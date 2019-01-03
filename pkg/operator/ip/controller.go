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
	clientset "github.com/inwinstack/blended/client/clientset/versioned/typed/inwinstack/v1"
	"github.com/inwinstack/ipam/pkg/constants"
	"github.com/inwinstack/ipam/pkg/util"
	"github.com/inwinstack/ipam/pkg/util/slice"
	opkit "github.com/inwinstack/operator-kit"
	"k8s.io/api/core/v1"
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
	clientset clientset.InwinstackV1Interface
}

func NewController(ctx *opkit.Context, clientset clientset.InwinstackV1Interface) *IPController {
	return &IPController{ctx: ctx, clientset: clientset}
}

func (c *IPController) StartWatch(namespace string, stopCh chan struct{}) error {
	resourceHandlerFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAdd,
		UpdateFunc: c.onUpdate,
		DeleteFunc: c.onDelete,
	}

	glog.Infof("Start watching IP resources.")
	watcher := opkit.NewWatcher(Resource, namespace, resourceHandlerFuncs, c.clientset.RESTClient())
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
	ip := newObj.(*inwinv1.IP).DeepCopy()
	glog.V(2).Infof("Received update on IP %s in namespace %s.", ip.Name, ip.Namespace)

	if ip.Status.Phase == inwinv1.IPActive {
		if err := c.makeNamespaceRefresh(ip); err != nil {
			glog.Errorf("Failed to update namespace annotations for %s in %s namespace: %+v.", ip.Name, ip.Namespace, err)
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

		if err := c.makeNamespaceRefresh(ip); err != nil {
			glog.Errorf("Failed to update namespace annotations for %s in %s namespace: %+v.", ip.Name, ip.Namespace, err)
		}
	}
}

func (c *IPController) allocate(ip *inwinv1.IP) error {
	pool, err := c.clientset.Pools().Get(ip.Spec.PoolName, metav1.GetOptions{})
	if err != nil {
		return c.makeFailedStatus(ip, err)
	}

	ip.Status.Phase = inwinv1.IPFailed
	if pool.Status.Phase == inwinv1.PoolActive {
		nets, _ := util.ParseCIDR(pool.Spec.Address)

		var ips []string
		for _, net := range nets {
			ips = append([]string{}, append(ips, util.GetAllIP(net)...)...)
		}

		// Filter allocated IPs
		ips = slice.RemoveItems(ips, pool.Status.AllocatedIPs)

		if len(ips) != 0 {
			ip.Status.Address = ips[0]
			ip.Status.Ports = []int{}
			ip.Status.Phase = inwinv1.IPActive
			pool.Status.AllocatedIPs = append(pool.Status.AllocatedIPs, ips[0])
			pool.Status.Allocatable = pool.Status.Capacity - len(pool.Status.AllocatedIPs)
			pool.Status.LastUpdateTime = metav1.NewTime(time.Now())
			if _, err := c.clientset.Pools().Update(pool); err != nil {
				return c.makeFailedStatus(ip, err)
			}
		}

		if len(ips) == 0 {
			ip.Status.Reason = fmt.Sprintf("Pool \"%s\" has been exhausted for IP", pool.Name)
		}
	}

	ip.Status.LastUpdateTime = metav1.NewTime(time.Now())
	if _, err := c.clientset.IPs(ip.Namespace).Update(ip); err != nil {
		return err
	}
	return nil
}

func (c *IPController) deallocate(ip *inwinv1.IP) error {
	pool, err := c.clientset.Pools().Get(ip.Spec.PoolName, metav1.GetOptions{})
	if err != nil {
		return c.makeFailedStatus(ip, err)
	}

	if pool.Status.Phase == inwinv1.PoolActive {
		pool.Status.AllocatedIPs = slice.RemoveItem(pool.Status.AllocatedIPs, ip.Status.Address)
		pool.Status.Allocatable = pool.Status.Capacity - len(pool.Status.AllocatedIPs)
		pool.Status.LastUpdateTime = metav1.NewTime(time.Now())
		if _, err := c.clientset.Pools().Update(pool); err != nil {
			return c.makeFailedStatus(ip, err)
		}
	}
	return nil
}

func (c *IPController) makeNamespaceRefresh(ip *inwinv1.IP) error {
	if ip.Spec.UpdateNamespace {
		ns, err := c.ctx.Clientset.CoreV1().Namespaces().Get(ip.Namespace, metav1.GetOptions{})
		if err != nil {
			// When a namespace has been deleted, we don't do anything
			return nil
		}

		if ns.Status.Phase != v1.NamespaceTerminating {
			ns.Annotations[constants.AnnKeyNamespaceRefresh] = "true"
			if _, err := c.ctx.Clientset.CoreV1().Namespaces().Update(ns); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *IPController) makeFailedStatus(ip *inwinv1.IP, e error) error {
	ip.Status.Phase = inwinv1.IPFailed
	ip.Status.Reason = e.Error()
	ip.Status.LastUpdateTime = metav1.NewTime(time.Now())
	if _, err := c.clientset.IPs(ip.Namespace).Update(ip); err != nil {
		return err
	}
	return nil
}
