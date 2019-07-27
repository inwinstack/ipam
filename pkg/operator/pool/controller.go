/*
Copyright © 2018 inwinSTACK Inc

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
	"context"
	"fmt"
	"time"

	"github.com/thoas/go-funk"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"
	blendedv1 "github.com/inwinstack/blended/apis/inwinstack/v1"
	"github.com/inwinstack/blended/constants"
	blended "github.com/inwinstack/blended/generated/clientset/versioned"
	informerv1 "github.com/inwinstack/blended/generated/informers/externalversions/inwinstack/v1"
	listerv1 "github.com/inwinstack/blended/generated/listers/inwinstack/v1"
	"github.com/inwinstack/blended/k8sutil"
	"github.com/inwinstack/ipam/pkg/ipaddr"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// Controller represents the controller of pool
type Controller struct {
	blendedset blended.Interface
	lister     listerv1.PoolLister
	synced     cache.InformerSynced
	queue      workqueue.RateLimitingInterface
}

// NewController creates an instance of the pool controller
func NewController(blendedset blended.Interface, informer informerv1.PoolInformer) *Controller {
	controller := &Controller{
		blendedset: blendedset,
		lister:     informer.Lister(),
		synced:     informer.Informer().HasSynced,
		queue:      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Pools"),
	}
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueue,
		UpdateFunc: func(old, new interface{}) {
			oo := old.(*blendedv1.Pool)
			no := new.(*blendedv1.Pool)
			k8sutil.MakeNeedToUpdate(&no.ObjectMeta, oo.Spec, no.Spec)
			controller.enqueue(no)
		},
	})
	return controller
}

// Run serves the pool controller
func (c *Controller) Run(ctx context.Context, threadiness int) error {
	glog.Info("Starting the pool controller")
	glog.Info("Waiting for the pool informer caches to sync")
	if ok := cache.WaitForCacheSync(ctx.Done(), c.synced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, ctx.Done())
	}
	return nil
}

// Stop stops the pool controller
func (c *Controller) Stop() {
	glog.Info("Stopping the pool controller")
	c.queue.ShutDown()
}

func (c *Controller) runWorker() {
	defer utilruntime.HandleCrash()
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.queue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.queue.Done(obj)
		key, ok := obj.(string)
		if !ok {
			c.queue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("Pool expected string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.reconcile(key); err != nil {
			c.queue.AddRateLimited(key)
			return fmt.Errorf("Pool error syncing '%s': %s, requeuing", key, err.Error())
		}

		c.queue.Forget(obj)
		glog.V(2).Infof("Pool successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

func (c *Controller) enqueue(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.queue.Add(key)
}

func (c *Controller) reconcile(key string) error {
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return err
	}

	pool, err := c.lister.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("pool '%s' in work queue no longer exists", key))
			return err
		}
		return err
	}

	if !pool.ObjectMeta.DeletionTimestamp.IsZero() {
		return c.cleanup(pool)
	}

	if err := c.checkAndUdateFinalizer(pool); err != nil {
		return err
	}

	need := k8sutil.IsNeedToUpdate(pool.ObjectMeta)
	if pool.Status.Phase != blendedv1.PoolActive || need {
		if err := c.makeStatus(pool); err != nil {
			return c.makeFailedStatus(pool, err)
		}
	}
	return nil
}

func (c *Controller) checkAndUdateFinalizer(pool *blendedv1.Pool) error {
	poolCopy := pool.DeepCopy()
	ok := funk.ContainsString(poolCopy.Finalizers, constants.CustomFinalizer)
	if poolCopy.Status.Phase == blendedv1.PoolActive && !ok {
		k8sutil.AddFinalizer(&poolCopy.ObjectMeta, constants.CustomFinalizer)
		if _, err := c.blendedset.InwinstackV1().Pools().Update(poolCopy); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) makeStatus(pool *blendedv1.Pool) error {
	poolCopy := pool.DeepCopy()
	if poolCopy.Status.AllocatedIPs == nil {
		poolCopy.Status.AllocatedIPs = []string{}
	}

	parser := ipaddr.NewParser(poolCopy.Spec.Addresses, poolCopy.Spec.AvoidBuggyIPs, poolCopy.Spec.AvoidGatewayIPs)
	ips, err := parser.FilterIPs(pool.Spec.FilterIPs)
	if err != nil {
		return err
	}

	poolCopy.Status.Reason = ""
	poolCopy.Status.Capacity = len(ips)
	poolCopy.Status.Allocatable = len(ips) - len(poolCopy.Status.AllocatedIPs)
	poolCopy.Status.LastUpdateTime = metav1.NewTime(time.Now())
	poolCopy.Status.Phase = blendedv1.PoolActive
	delete(poolCopy.Annotations, constants.NeedUpdateKey)
	k8sutil.AddFinalizer(&poolCopy.ObjectMeta, constants.CustomFinalizer)
	if _, err := c.blendedset.InwinstackV1().Pools().Update(poolCopy); err != nil {
		return err
	}
	return nil
}

func (c *Controller) makeFailedStatus(pool *blendedv1.Pool, e error) error {
	poolCopy := pool.DeepCopy()
	poolCopy.Status.Reason = e.Error()
	poolCopy.Status.Phase = blendedv1.PoolFailed
	poolCopy.Status.LastUpdateTime = metav1.NewTime(time.Now())
	delete(poolCopy.Annotations, constants.NeedUpdateKey)
	if _, err := c.blendedset.InwinstackV1().Pools().Update(poolCopy); err != nil {
		return err
	}
	glog.Errorf("Pool got an error: %+v.", e)
	return nil
}

func (c *Controller) cleanup(pool *blendedv1.Pool) error {
	poolCopy := pool.DeepCopy()
	poolCopy.Status.Phase = blendedv1.PoolTerminating
	if len(poolCopy.Status.AllocatedIPs) == 0 {
		k8sutil.RemoveFinalizer(&poolCopy.ObjectMeta, constants.CustomFinalizer)
	}

	if _, err := c.blendedset.InwinstackV1().Pools().Update(poolCopy); err != nil {
		return err
	}
	return nil
}
