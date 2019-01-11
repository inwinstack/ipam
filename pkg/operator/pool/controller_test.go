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
			Address:                   "172.22.132.150-172.22.132.200",
			IgnoreNamespaceAnnotation: true,
			AutoAssignToNamespace:     false,
			IgnoreNamespaces:          []string{"kube-system", "kube-public", "default"},
		},
	}

	createPool, err := client.InwinstackV1().Pools().Create(pool)
	assert.Nil(t, err)

	controller.onAdd(createPool)

	onAddPool, err := client.InwinstackV1().Pools().Get(pool.Name, metav1.GetOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, onAddPool.Status.Phase, inwinv1.PoolActive)
	assert.Equal(t, onAddPool.Status.AllocatedIPs, []string{})
	assert.Equal(t, onAddPool.Status.Capacity, 51)
	assert.Equal(t, onAddPool.Status.Allocatable, 51)

	// Test onUpdate
	// TODO: The pool needs to implement onUpdate function.
	onAddPool.Spec.Address = "172.22.132.0/24"
	controller.onUpdate(createPool, onAddPool)

	onUpdatePool, err := client.InwinstackV1().Pools().Get(onAddPool.Name, metav1.GetOptions{})
	assert.Nil(t, err)

	// Test onDelete
	controller.onDelete(onUpdatePool)
}
