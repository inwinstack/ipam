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
	"testing"
	"time"

	blendedv1 "github.com/inwinstack/blended/apis/inwinstack/v1"
	blendedfake "github.com/inwinstack/blended/generated/clientset/versioned/fake"
	blendedinformers "github.com/inwinstack/blended/generated/informers/externalversions"
	"github.com/inwinstack/ipam/pkg/config"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const timeout = 3 * time.Second

func TestPoolController(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := &config.Config{Threads: 2}
	blendedset := blendedfake.NewSimpleClientset()
	informer := blendedinformers.NewSharedInformerFactory(blendedset, 0)

	controller := NewController(blendedset, informer.Inwinstack().V1().Pools())
	go informer.Start(ctx.Done())
	assert.Nil(t, controller.Run(ctx, cfg.Threads))

	pool := &blendedv1.Pool{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pool",
		},
		Spec: blendedv1.PoolSpec{
			Addresses:         []string{"172.22.132.0-172.22.132.5"},
			AssignToNamespace: false,
			AvoidBuggyIPs:     true,
			AvoidGatewayIPs:   false,
			IgnoreNamespaces:  []string{"kube-system", "kube-public", "default"},
		},
	}

	// Create the pool
	_, err := blendedset.InwinstackV1().Pools().Create(pool)
	assert.Nil(t, err)

	failed := true
	for start := time.Now(); time.Since(start) < timeout; {
		p, err := blendedset.InwinstackV1().Pools().Get(pool.Name, metav1.GetOptions{})
		assert.Nil(t, err)

		if p.Status.Phase == blendedv1.PoolActive {
			assert.Equal(t, []string{}, p.Status.AllocatedIPs)
			assert.Equal(t, 5, p.Status.Capacity)
			assert.Equal(t, 5, p.Status.Allocatable)
			failed = false
			break
		}
	}
	assert.Equal(t, false, failed, "The pool object failed to make status.")

	// Success to update the pool
	gpool, err := blendedset.InwinstackV1().Pools().Get(pool.Name, metav1.GetOptions{})
	assert.Nil(t, err)

	gpool.Spec.Addresses = append(gpool.Spec.Addresses, "172.22.132.250-172.22.132.255")
	_, err = blendedset.InwinstackV1().Pools().Update(gpool)
	assert.Nil(t, err)

	failed = true
	for start := time.Now(); time.Since(start) < timeout; {
		p, err := blendedset.InwinstackV1().Pools().Get(pool.Name, metav1.GetOptions{})
		assert.Nil(t, err)

		if p.Status.Capacity == 10 {
			assert.Equal(t, 10, p.Status.Allocatable)
			failed = false
			break
		}
	}
	assert.Equal(t, false, failed, "The service object failed to sync status.")

	// Failed to update the pool
	gpool, err = blendedset.InwinstackV1().Pools().Get(pool.Name, metav1.GetOptions{})
	assert.Nil(t, err)
	gpool.Spec.Addresses = []string{"172.22.132.250-172.22.132.267"}

	_, err = blendedset.InwinstackV1().Pools().Update(gpool)
	assert.Nil(t, err)

	failed = true
	for start := time.Now(); time.Since(start) < timeout; {
		p, err := blendedset.InwinstackV1().Pools().Get(pool.Name, metav1.GetOptions{})
		assert.Nil(t, err)

		if p.Status.Phase == blendedv1.PoolFailed {
			assert.NotNil(t, p.Status.Reason)
			failed = false
			break
		}
	}
	assert.Equal(t, false, failed, "The service object failed to get error status.")

	// Delete the pool
	assert.Nil(t, blendedset.InwinstackV1().Pools().Delete(pool.Name, nil))

	_, err = blendedset.InwinstackV1().Pools().Get(pool.Name, metav1.GetOptions{})
	assert.NotNil(t, err)

	cancel()
	controller.Stop()
}
