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

package ip

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

	controller := NewController(blendedset, informer.Inwinstack().V1().IPs())
	go informer.Start(ctx.Done())
	assert.Nil(t, controller.Run(ctx, cfg.Threads))

	pool := &blendedv1.Pool{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: blendedv1.PoolSpec{
			Addresses:         []string{"172.22.132.0-172.22.132.5"},
			AssignToNamespace: false,
			AvoidBuggyIPs:     true,
			AvoidGatewayIPs:   false,
		},
		Status: blendedv1.PoolStatus{
			Phase:          blendedv1.PoolActive,
			AllocatedIPs:   []string{},
			Capacity:       5,
			Allocatable:    5,
			LastUpdateTime: metav1.NewTime(time.Now()),
		},
	}
	_, err := blendedset.InwinstackV1().Pools().Create(pool)
	assert.Nil(t, err)

	ip := &blendedv1.IP{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ip-1",
			Namespace: "default",
		},
		Spec: blendedv1.IPSpec{
			PoolName: pool.Name,
		},
	}
	_, err = blendedset.InwinstackV1().IPs(ip.Namespace).Create(ip)
	assert.Nil(t, err)

	failed := true
	for start := time.Now(); time.Since(start) < timeout; {
		gip, err := blendedset.InwinstackV1().IPs(ip.Namespace).Get(ip.Name, metav1.GetOptions{})
		assert.Nil(t, err)

		if gip.Status.Phase == blendedv1.IPActive {
			assert.Equal(t, "172.22.132.1", gip.Status.Address)
			gpool, err := blendedset.InwinstackV1().Pools().Get(pool.Name, metav1.GetOptions{})
			assert.Nil(t, err)
			assert.Equal(t, []string{"172.22.132.1"}, gpool.Status.AllocatedIPs)
			assert.Equal(t, 4, gpool.Status.Allocatable)
			assert.Equal(t, 5, gpool.Status.Capacity)
			failed = false
			break
		}
	}
	assert.Equal(t, false, failed, "The service object failed to allocate IP.")

	// Test to deallocate IP
	gip, err := blendedset.InwinstackV1().IPs(ip.Namespace).Get(ip.Name, metav1.GetOptions{})
	assert.Nil(t, err)
	assert.Nil(t, controller.deallocate(gip))

	gpool, err := blendedset.InwinstackV1().Pools().Get(pool.Name, metav1.GetOptions{})
	assert.Nil(t, err)
	assert.Equal(t, []string{}, gpool.Status.AllocatedIPs)
	assert.Equal(t, 5, gpool.Status.Allocatable)
	assert.Equal(t, 5, gpool.Status.Capacity)

	cancel()
	controller.Stop()
}
