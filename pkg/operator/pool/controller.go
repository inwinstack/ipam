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
	inwinv1 "github.com/inwinstack/blended/apis/inwinstack/v1"
	inwinclientset "github.com/inwinstack/blended/client/clientset/versioned/typed/inwinstack/v1"
	"github.com/inwinstack/ipam-operator/pkg/constants"
	"github.com/inwinstack/ipam-operator/pkg/k8sutil"
	"github.com/inwinstack/ipam-operator/pkg/util"
	opkit "github.com/inwinstack/operator-kit"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	customResourceName       = "pool"
	customResourceNamePlural = "pools"
)

var Resource = opkit.CustomResource{
	Name:    customResourceName,
	Plural:  customResourceNamePlural,
	Group:   inwinv1.CustomResourceGroup,
	Version: inwinv1.Version,
	Scope:   apiextensionsv1beta1.ClusterScoped,
	Kind:    reflect.TypeOf(inwinv1.Pool{}).Name(),
}

type PoolController struct {
	ctx       *opkit.Context
	clientset inwinclientset.InwinstackV1Interface
}

func NewController(ctx *opkit.Context, clientset inwinclientset.InwinstackV1Interface) *PoolController {
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
	go watcher.Watch(&inwinv1.Pool{}, stopCh)
	return nil
}

func (c *PoolController) CreateDefaultPool(address string, namespaces []string, auto, ignore bool) error {
	if address == "" && namespaces == nil {
		return fmt.Errorf("Miss address and namespaces flag")
	}

	_, err := c.clientset.Pools().Get(constants.DefaultPool, metav1.GetOptions{})
	if err == nil {
		glog.V(2).Infof("The default pool already exists.")
		return nil
	}

	pool := k8sutil.NewDefaultPool(address, namespaces, auto, ignore)
	if _, err := c.clientset.Pools().Create(pool); err != nil {
		return err
	}
	glog.Infof("The default pool has created.")
	return nil
}

func (c *PoolController) onAdd(obj interface{}) {
	pool := obj.(*inwinv1.Pool).DeepCopy()
	glog.V(2).Infof("Pool %s has added.", pool.Name)

	if err := c.makeStatus(pool); err != nil {
		glog.Errorf("Failed to init status in %s pool: %s.", pool.Name, err)
	}
}

func (c *PoolController) onUpdate(oldObj, newObj interface{}) {
	pool := newObj.(*inwinv1.Pool).DeepCopy()
	glog.V(2).Infof("Received update on Pool %s.", pool.Name)
}

func (c *PoolController) onDelete(obj interface{}) {
	glog.V(2).Infof("Pool %s has deleted.", obj.(*inwinv1.Pool).Name)
}

func (c *PoolController) makeStatus(pool *inwinv1.Pool) error {
	if pool.Status.Capacity == 0 && pool.Status.Phase != inwinv1.PoolActive {
		nets, err := util.ParseCIDR(pool.Spec.Address)
		if err != nil {
			pool.Status.Phase = inwinv1.PoolFailed
			pool.Status.Reason = fmt.Sprintf("Invalid parse CIDR from %s.", pool.Spec.Address)
		}

		var ips []string
		for _, net := range nets {
			ips = append([]string{}, append(ips, util.GetAllIP(net)...)...)
		}

		pool.Status.Capacity = len(ips)
		pool.Status.Allocatable = len(ips)
		pool.Status.AllocatedIPs = []string{}
		pool.Status.Phase = inwinv1.PoolActive
		pool.Status.LastUpdateTime = metav1.NewTime(time.Now())
		if _, err := c.clientset.Pools().Update(pool); err != nil {
			return err
		}
	}
	return nil
}
