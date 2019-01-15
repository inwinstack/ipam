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
	"testing"
	"time"

	inwinv1 "github.com/inwinstack/blended/apis/inwinstack/v1"
	fake "github.com/inwinstack/blended/client/clientset/versioned/fake"
	opkit "github.com/inwinstack/operator-kit"

	extensionsfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	corefake "k8s.io/client-go/kubernetes/fake"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPoolController(t *testing.T) {
	client := fake.NewSimpleClientset()
	coreClient := corefake.NewSimpleClientset()
	extensionsClient := extensionsfake.NewSimpleClientset()

	ctx := &opkit.Context{
		Clientset:             coreClient,
		APIExtensionClientset: extensionsClient,
		Interval:              500 * time.Millisecond,
		Timeout:               60 * time.Second,
	}

	controller := NewController(ctx, client)
	assert.NotNil(t, controller)

	// Test onAdd
	pool := &inwinv1.Pool{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pool",
		},
		Spec: inwinv1.PoolSpec{
			Addresses:         []string{"172.22.132.0-172.22.132.5"},
			AssignToNamespace: false,
			AvoidBuggyIPs:     true,
			AvoidGatewayIPs:   false,
			IgnoreNamespaces:  []string{"kube-system", "kube-public", "default"},
		},
	}

	createPool, err := client.InwinstackV1().Pools().Create(pool)
	assert.Nil(t, err)

	controller.onAdd(createPool)

	onAddPool, err := client.InwinstackV1().Pools().Get(pool.Name, metav1.GetOptions{})
	assert.Nil(t, err)
	assert.Equal(t, onAddPool.Status.Phase, inwinv1.PoolActive)
	assert.Equal(t, onAddPool.Status.AllocatedIPs, []string{})
	assert.Equal(t, onAddPool.Status.Capacity, 5)
	assert.Equal(t, onAddPool.Status.Allocatable, 5)

	// Test onUpdate
	onAddPool.Spec.Addresses = append(onAddPool.Spec.Addresses, "172.22.132.250-172.22.132.255")
	controller.onUpdate(createPool, onAddPool)

	onUpdatePool, err := client.InwinstackV1().Pools().Get(onAddPool.Name, metav1.GetOptions{})
	assert.Nil(t, err)
	assert.Equal(t, onUpdatePool.Status.Capacity, 10)
	assert.Equal(t, onUpdatePool.Status.Allocatable, 10)

	// Test onUpdate failed
	onUpdatePool.Spec.Addresses = []string{"172.22.132.250-172.22.132.267"}
	controller.onUpdate(onAddPool, onUpdatePool)

	onUpdateFailedPool, err := client.InwinstackV1().Pools().Get(onAddPool.Name, metav1.GetOptions{})
	assert.Nil(t, err)
	assert.Equal(t, onUpdateFailedPool.Status.Phase, inwinv1.PoolFailed)
	assert.Equal(t, onUpdateFailedPool.Status.Capacity, 0)
	assert.Equal(t, onUpdateFailedPool.Status.Allocatable, 0)
	assert.NotNil(t, onUpdateFailedPool.Status.Reason)
}
