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

package ip

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
	customResourceName       = "ip"
	customResourceNamePlural = "ips"
)

var Resource = opkit.CustomResource{
	Name:    customResourceName,
	Plural:  customResourceNamePlural,
	Group:   inwinv1.CustomResourceGroup,
	Version: inwinv1.Version,
	Scope:   apiextensionsv1beta1.NamespaceScoped,
	Kind:    reflect.TypeOf(inwinv1.IP{}).Name(),
}

type IPController struct {
	ctx       *opkit.Context
	clientset clientset.Interface
}

func NewController(ctx *opkit.Context, clientset clientset.Interface) *IPController {
	return &IPController{ctx: ctx, clientset: clientset}
}

func (c *IPController) StartWatch(namespace string, stopCh chan struct{}) error {
	resourceHandlerFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAdd,
		UpdateFunc: c.onUpdate,
		DeleteFunc: c.onDelete,
	}

	glog.Infof("Start watching IP resources.")
	watcher := opkit.NewWatcher(Resource, namespace, resourceHandlerFuncs, c.clientset.InwinstackV1().RESTClient())
	go watcher.Watch(&inwinv1.IP{}, stopCh)
	return nil
}

func (c *IPController) onAdd(obj interface{}) {
	ip := obj.(*inwinv1.IP).DeepCopy()
	glog.V(2).Infof("Received add on IP %s in %s namespace.", ip.Name, ip.Namespace)

	if ip.Status.Phase != inwinv1.IPActive {
		if err := c.allocate(ip); err != nil {
			glog.Errorf("Failed to allocate IP for %s in %s namespace: %+v.", ip.Name, ip.Namespace, err)
		}
	}
}

func (c *IPController) onUpdate(oldObj, newObj interface{}) {
	old := oldObj.(*inwinv1.IP).DeepCopy()
	ip := newObj.(*inwinv1.IP).DeepCopy()
	glog.V(2).Infof("Received update on IP %s in namespace %s.", ip.Name, ip.Namespace)

	if old.Spec.PoolName != ip.Spec.PoolName {
		if err := c.allocate(ip); err != nil {
			glog.Errorf("Failed to allocate new IP for %s in %s namespace: %+v.", ip.Name, ip.Namespace, err)
		}

		if err := c.deallocate(old); err != nil {
			glog.Errorf("Failed to deallocate old IP for %s in %s namespace: %+v.", old.Name, old.Namespace, err)
		}
	}
}

func (c *IPController) onDelete(obj interface{}) {
	ip := obj.(*inwinv1.IP).DeepCopy()
	glog.V(2).Infof("Received delete on IP %s in %s namespace.", ip.Name, ip.Namespace)

	if ip.Status.Phase == inwinv1.IPActive {
		if err := c.deallocate(ip); err != nil {
			glog.Errorf("Failed to deallocate IP for %s in %s namespace: %+v.", ip.Name, ip.Namespace, err)
		}
	}
}

func (c *IPController) makeFailedStatus(ip *inwinv1.IP, e error) error {
	ip.Status.Phase = inwinv1.IPFailed
	ip.Status.Reason = fmt.Sprintf("%+v.", e)
	ip.Status.Address = ""
	ip.Status.LastUpdateTime = metav1.NewTime(time.Now())
	if _, err := c.clientset.InwinstackV1().IPs(ip.Namespace).Update(ip); err != nil {
		return err
	}
	return nil
}

func (c *IPController) updatePool(pool *inwinv1.Pool) error {
	pool.Status.LastUpdateTime = metav1.NewTime(time.Now())
	if _, err := c.clientset.InwinstackV1().Pools().Update(pool); err != nil {
		return err
	}
	return nil
}

func (c *IPController) allocate(ip *inwinv1.IP) error {
	pool, err := c.clientset.InwinstackV1().Pools().Get(ip.Spec.PoolName, metav1.GetOptions{})
	if err != nil {
		return c.makeFailedStatus(ip, err)
	}

	if pool.Status.Phase == inwinv1.PoolFailed {
		err := fmt.Errorf("Unable to allocate IP from failed pool")
		return c.makeFailedStatus(ip, err)
	}

	if pool.Status.Allocatable == 0 {
		err := fmt.Errorf("The pool \"%s\" has been exhausted", pool.Name)
		return c.makeFailedStatus(ip, err)
	}

	np := util.NewNetworkParser(pool.Spec.Addresses, pool.Spec.AvoidBuggyIPs, pool.Spec.AvoidGatewayIPs)
	ips, _ := np.IPs()

	filterIPs := pool.Status.AllocatedIPs
	if pool.Spec.FilterIPs != nil {
		filterIPs = append([]string{}, append(filterIPs, pool.Spec.FilterIPs...)...)
	}

	// Filter IPs
	for _, rem := range filterIPs {
		ips = slice.FilterString(ips, func(v string) bool {
			return v != rem
		})
	}

	pool.Status.AllocatedIPs = append(pool.Status.AllocatedIPs, ips[0])
	pool.Status.Allocatable = pool.Status.Capacity - len(pool.Status.AllocatedIPs)
	if err := c.updatePool(pool); err != nil {
		return c.makeFailedStatus(ip, err)
	}

	ip.Status.Address = ips[0]
	ip.Status.Phase = inwinv1.IPActive
	ip.Status.Reason = ""
	ip.Status.LastUpdateTime = metav1.NewTime(time.Now())
	if _, err := c.clientset.InwinstackV1().IPs(ip.Namespace).Update(ip); err != nil {
		return err
	}
	return nil
}

func (c *IPController) deallocate(ip *inwinv1.IP) error {
	pool, err := c.clientset.InwinstackV1().Pools().Get(ip.Spec.PoolName, metav1.GetOptions{})
	if err != nil {
		return c.makeFailedStatus(ip, err)
	}

	if pool.Status.Phase == inwinv1.PoolActive {
		pool.Status.AllocatedIPs = slice.FilterString(pool.Status.AllocatedIPs, func(v string) bool {
			return v != ip.Status.Address
		})
		pool.Status.Allocatable = pool.Status.Capacity - len(pool.Status.AllocatedIPs)
		if err := c.updatePool(pool); err != nil {
			return c.makeFailedStatus(ip, err)
		}
	}
	return nil
}
