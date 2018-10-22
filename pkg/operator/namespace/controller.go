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

package namespace

import (
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	inwinclientset "github.com/inwinstack/ipam-operator/pkg/client/clientset/versioned/typed/inwinstack/v1alpha1"
	opkit "github.com/inwinstack/operator-kit"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

var Resource = opkit.CustomResource{
	Name:    "namespace",
	Plural:  "namespaces",
	Version: "v1",
	Kind:    reflect.TypeOf(v1.Namespace{}).Name(),
}

const (
	AllocatedIPs      = "inwinstack.com/allocated-ips"
	AllocatedLatestIP = "inwinstack.com/allocated-latest-ip"
	AllocateIPNumber  = "inwinstack.com/allocate-ip-number"
	AllocatePoolName  = "inwinstack.com/allocate-pool-name"
)

const defaultPoolName = "default-pool"

type NamespaceController struct {
	ctx       *opkit.Context
	clientset inwinclientset.InwinstackV1alpha1Interface
}

func NewController(ctx *opkit.Context, clientset inwinclientset.InwinstackV1alpha1Interface) *NamespaceController {
	return &NamespaceController{ctx: ctx, clientset: clientset}
}

func (c *NamespaceController) StartWatch(namespace string, stopCh chan struct{}) error {
	resourceHandlerFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAdd,
		UpdateFunc: c.onUpdate,
		DeleteFunc: c.onDelete,
	}

	glog.Infof("Start watching namespace resources.")
	watcher := opkit.NewWatcher(Resource, namespace, resourceHandlerFuncs, c.ctx.Clientset.CoreV1().RESTClient())
	go watcher.Watch(&v1.Namespace{}, stopCh)
	return nil
}

func (c *NamespaceController) initAnnotations(ns *v1.Namespace) error {
	if ns.Annotations == nil {
		ns.Annotations = map[string]string{}
	}

	if ns.Annotations[AllocateIPNumber] == "" {
		ns.Annotations[AllocateIPNumber] = strconv.Itoa(1)
	}

	if ns.Annotations[AllocatePoolName] == "" {
		ns.Annotations[AllocatePoolName] = defaultPoolName
	}

	_, err := c.ctx.Clientset.CoreV1().Namespaces().Update(ns)
	if err != nil {
		return err
	}
	return nil
}

func (c *NamespaceController) addOrDeleteIPs(name, poolName string, ipNumber int, ips []string) ([]string, error) {
	pool, err := c.clientset.IPPools().Get(poolName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if containsNamespace(pool.Spec.IgnoreNamespaces, name) {
		return nil, nil
	}

	count := len(ips)
	switch {
	case len(ips) < ipNumber:
		for _, ip := range pool.Status.AllocatableIPs {
			if count >= ipNumber {
				break
			}
			ips = append(ips, ip)
			pool.Status.AllocatedIPs = append(pool.Status.AllocatedIPs, ip)
			count++
		}
	case len(ips) > ipNumber:
		for _, ip := range ips {
			if count <= ipNumber {
				break
			}

			if ipNumber != 0 {
				ips = removeIP(ips, ip)
			}

			pool.Status.AllocatedIPs = removeIP(pool.Status.AllocatedIPs, ip)
			pool.Status.AllocatableIPs = append(pool.Status.AllocatableIPs, ip)
			count--
		}
	}

	pool.Status.LastUpdateTime = metav1.NewTime(time.Now())
	if _, err := c.clientset.IPPools().Update(pool); err != nil {
		return nil, err
	}
	return ips, nil
}

func (c *NamespaceController) handleAssignIPs(ns *v1.Namespace) error {
	ips := parseIPs(ns.Annotations[AllocatedIPs])
	poolName := ns.Annotations[AllocatePoolName]
	ipNumber, err := strconv.Atoi(ns.Annotations[AllocateIPNumber])
	if err != nil {
		return err
	}

	if len(ips) == ipNumber {
		return nil
	}

	newIPs, err := c.addOrDeleteIPs(ns.Name, poolName, ipNumber, ips)
	if err != nil {
		return err
	}

	if len(newIPs) > 0 && ns != nil {
		ns.Annotations[AllocatedIPs] = strings.Join(newIPs, ",")
		ns.Annotations[AllocatedLatestIP] = newIPs[len(newIPs)-1]
		if _, err := c.ctx.Clientset.CoreV1().Namespaces().Update(ns); err != nil {
			return err
		}
	}
	return nil
}

func (c *NamespaceController) handleDeleteIPs(ns *v1.Namespace) error {
	ips := parseIPs(ns.Annotations[AllocatedIPs])
	poolName := ns.Annotations[AllocatePoolName]
	ipNumber, err := strconv.Atoi(ns.Annotations[AllocateIPNumber])
	if err != nil {
		return err
	}

	if _, err := c.addOrDeleteIPs(ns.Name, poolName, ipNumber, ips); err != nil {
		return err
	}
	return nil
}

func parseIPs(v string) []string {
	if v == "" {
		return []string{}
	}
	return strings.Split(v, ",")
}

func removeIP(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func containsNamespace(sl []string, v string) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

func (c *NamespaceController) onAdd(obj interface{}) {
	ns := obj.(*v1.Namespace).DeepCopy()
	glog.V(2).Infof("Namespace %s resource has added.", ns.Name)

	if err := c.initAnnotations(ns); err != nil {
		glog.Infof("Failed to init annotations in %s namespace: %s.", ns.Name, err)
	}
}

func (c *NamespaceController) onUpdate(oldObj, newObj interface{}) {
	newns := newObj.(*v1.Namespace).DeepCopy()
	glog.V(2).Infof("Namespace %s resource has updated.", newns.Name)

	if err := c.handleAssignIPs(newns); err != nil {
		glog.Infof("Failed to assign IPs in %s namespace: %s.", newns.Name, err)
	}
}

func (c *NamespaceController) onDelete(obj interface{}) {
	ns := obj.(*v1.Namespace).DeepCopy()
	glog.V(2).Infof("Namespace %s resource has deleted.", ns.Name)

	ns.Annotations[AllocateIPNumber] = "0"
	if err := c.handleDeleteIPs(ns); err != nil {
		glog.Infof("Failed to delete IPs in %s namespace: %s.", ns.Name, err)
	}
}
