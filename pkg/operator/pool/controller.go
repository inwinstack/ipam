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
	clientset "github.com/inwinstack/blended/client/clientset/versioned"
	"github.com/inwinstack/ipam/pkg/util"
	opkit "github.com/inwinstack/operator-kit"
	slice "github.com/thoas/go-funk"
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
	clientset clientset.Interface
}

func NewController(ctx *opkit.Context, clientset clientset.Interface) *PoolController {
	return &PoolController{ctx: ctx, clientset: clientset}
}

func (c *PoolController) StartWatch(namespace string, stopCh chan struct{}) error {
	resourceHandlerFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAdd,
		UpdateFunc: c.onUpdate,
	}

	glog.Infof("Start watching pool resources.")
	watcher := opkit.NewWatcher(Resource, namespace, resourceHandlerFuncs, c.clientset.InwinstackV1().RESTClient())
	go watcher.Watch(&inwinv1.Pool{}, stopCh)
	return nil
}

func (c *PoolController) onAdd(obj interface{}) {
	pool := obj.(*inwinv1.Pool).DeepCopy()
	glog.V(2).Infof("Received add on Pool %s.", pool.Name)

	if err := c.makeStatus(pool); err != nil {
		glog.Errorf("Failed to init status on Pool %s: %+v.", pool.Name, err)
	}
}

func (c *PoolController) onUpdate(oldObj, newObj interface{}) {
	old := oldObj.(*inwinv1.Pool).DeepCopy()
	pool := newObj.(*inwinv1.Pool).DeepCopy()
	glog.V(2).Infof("Received update on Pool %s.", pool.Name)

	if !reflect.DeepEqual(old.Spec, pool.Spec) {
		if err := c.updateStatus(pool); err != nil {
			glog.Errorf("Failed to update status on Pool %s: %+v.", pool.Name, err)
		}
	}
}

func (c *PoolController) makeStatus(pool *inwinv1.Pool) error {
	if pool.Status.Capacity == 0 && pool.Status.Phase != inwinv1.PoolActive {
		if err := c.setStatus(true, pool); err != nil {
			return err
		}
	}
	return nil
}

func (c *PoolController) updateStatus(pool *inwinv1.Pool) error {
	return c.setStatus(false, pool)
}

func (c *PoolController) setStatus(init bool, pool *inwinv1.Pool) error {
	pool.Status.Phase = inwinv1.PoolActive
	np := util.NewNetworkParser(pool.Spec.Addresses, pool.Spec.AvoidBuggyIPs, pool.Spec.AvoidGatewayIPs)
	ips, err := np.IPs()
	if err != nil {
		pool.Status.Phase = inwinv1.PoolFailed
		pool.Status.Reason = fmt.Sprintf("%+v.", err)
	}

	if pool.Spec.FilterIPs != nil {
		for _, rem := range pool.Spec.FilterIPs {
			ips = slice.FilterString(ips, func(v string) bool {
				return v != rem
			})
		}
	}

	if init {
		pool.Status.AllocatedIPs = []string{}
	}

	if pool.Status.Phase == inwinv1.PoolActive {
		pool.Status.Reason = ""
	}

	pool.Status.Capacity = len(ips)
	pool.Status.Allocatable = len(ips)
	pool.Status.LastUpdateTime = metav1.NewTime(time.Now())
	if _, err := c.clientset.InwinstackV1().Pools().Update(pool); err != nil {
		return err
	}
	return nil
}
