/*
Copyright © 2018 inwinSTACK.inc

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pool

import (
	"fmt"
	"reflect"
	"time"

	"github.com/golang/glog"
	inwinalphav1 "github.com/inwinstack/ipam-operator/pkg/apis/inwinstack/v1alpha1"
	inwinclientset "github.com/inwinstack/ipam-operator/pkg/client/clientset/versioned/typed/inwinstack/v1alpha1"
	"github.com/inwinstack/ipam-operator/pkg/util"
	opkit "github.com/inwinstack/operator-kit"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	customResourceName       = "ippool"
	customResourceNamePlural = "ippools"
)

var Resource = opkit.CustomResource{
	Name:       customResourceName,
	Plural:     customResourceNamePlural,
	Group:      inwinalphav1.CustomResourceGroup,
	Version:    inwinalphav1.Version,
	Scope:      apiextensionsv1beta1.ClusterScoped,
	Kind:       reflect.TypeOf(inwinalphav1.IPPool{}).Name(),
	ShortNames: []string{"ipp"},
}

type PoolController struct {
	ctx       *opkit.Context
	clientset inwinclientset.InwinstackV1alpha1Interface
}

func NewController(ctx *opkit.Context, clientset inwinclientset.InwinstackV1alpha1Interface) *PoolController {
	return &PoolController{ctx: ctx, clientset: clientset}
}

func (c *PoolController) StartWatch(namespace string, stopCh chan struct{}) error {
	resourceHandlerFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAdd,
		UpdateFunc: c.onUpdate,
		DeleteFunc: c.onDelete,
	}

	glog.Infof("Start watching pool resources.")
	watcher := opkit.NewWatcher(Resource, namespace, resourceHandlerFuncs, c.clientset.RESTClient())
	go watcher.Watch(&inwinalphav1.IPPool{}, stopCh)
	return nil
}

func (c *PoolController) initStatus(pool *inwinalphav1.IPPool) error {
	if len(pool.Spec.Address) == 0 {
		pool.Status.Message = "IP pool has no prefixes defined."
	}

	if pool.Spec.Address != "" {
		nets, err := util.ParseCIDR(pool.Spec.Address)
		if err != nil {
			pool.Status.Message = fmt.Sprintf("Invalid parse CIDR from %s.", pool.Spec.Address)
		}

		if pool.Status.AllocatableIPs == nil && pool.Status.AllocatedIPs == nil {
			var newips []string
			for _, net := range nets {
				ips := util.GetAllIP(net)
				newips = append([]string{}, append(newips, ips...)...)
			}

			pool.Status.AllocatableIPs = newips
			pool.Status.Capacity = len(newips)
			pool.Status.AllocatedIPs = []string{}
			pool.Status.Message = "IP pool has been actived."
		}
	}

	pool.Status.LastUpdateTime = metav1.NewTime(time.Now())
	if _, err := c.clientset.IPPools().Update(pool); err != nil {
		return err
	}
	return nil
}

func (c *PoolController) updateStatus(pool *inwinalphav1.IPPool) error {
	if len(pool.Spec.Address) == 0 {
		pool.Status.Message = "IP pool has no prefixes defined."
	}

	total := len(pool.Status.AllocatedIPs) + len(pool.Status.AllocatableIPs)
	if pool.Spec.Address != "" && total != pool.Status.Capacity {
		for _, ip := range pool.Status.AllocatedIPs {
			pool.Status.AllocatableIPs = remove(pool.Status.AllocatableIPs, ip)
		}
	}

	pool.Status.AllocatableIPs = uniqueIPs(pool.Status.AllocatableIPs)
	pool.Status.AllocatedIPs = uniqueIPs(pool.Status.AllocatedIPs)
	pool.Status.LastUpdateTime = metav1.NewTime(time.Now())
	if _, err := c.clientset.IPPools().Update(pool); err != nil {
		return err
	}
	return nil
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func uniqueIPs(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func (c *PoolController) onAdd(obj interface{}) {
	pool := obj.(*inwinalphav1.IPPool).DeepCopy()
	glog.V(2).Infof("Pool %s resource has added.", pool.Name)

	if err := c.initStatus(pool); err != nil {
		glog.Infof("Failed to init status in %s pool: %s.", pool.Name, err)
	}
}

func (c *PoolController) onUpdate(oldObj, newObj interface{}) {
	newPool := newObj.(*inwinalphav1.IPPool).DeepCopy()
	glog.V(2).Infof("Pool %s resource has updated.", newPool.Name)

	if err := c.updateStatus(newPool); err != nil {
		glog.Infof("Failed to update status in %s pool: %s.", newPool.Name, err)
	}
}

func (c *PoolController) onDelete(obj interface{}) {
	glog.V(2).Infof("Pool %s resource has deleted: .", obj.(*inwinalphav1.IPPool).Name)
}
