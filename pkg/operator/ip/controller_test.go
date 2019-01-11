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
	"testing"
	"time"

	inwinv1 "github.com/inwinstack/blended/apis/inwinstack/v1"
	fake "github.com/inwinstack/blended/client/clientset/versioned/fake"
	"github.com/inwinstack/ipam/pkg/constants"
	opkit "github.com/inwinstack/operator-kit"

	"k8s.io/api/core/v1"
	extensionsfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	corefake "k8s.io/client-go/kubernetes/fake"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const namespace = "test"

func TestIPController(t *testing.T) {
	client := fake.NewSimpleClientset()
	coreClient := corefake.NewSimpleClientset()
	extensionsClient := extensionsfake.NewSimpleClientset()

	pool := &inwinv1.Pool{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: inwinv1.PoolSpec{
			Address:                   "172.22.132.150-172.22.132.200",
			IgnoreNamespaceAnnotation: false,
			AutoAssignToNamespace:     false,
			IgnoreNamespaces:          []string{"kube-system", "kube-public", "default"},
		},
		Status: inwinv1.PoolStatus{
			Phase:          inwinv1.PoolActive,
			AllocatedIPs:   []string{},
			Capacity:       51,
			Allocatable:    51,
			LastUpdateTime: metav1.NewTime(time.Now()),
		},
	}

	createPool, err := client.InwinstackV1().Pools().Create(pool)
	assert.Nil(t, err)

	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        namespace,
			Annotations: map[string]string{},
		},
	}

	_, nserr := coreClient.CoreV1().Namespaces().Create(ns)
	assert.Nil(t, nserr)

	ctx := &opkit.Context{
		Clientset:             coreClient,
		APIExtensionClientset: extensionsClient,
		Interval:              500 * time.Millisecond,
		Timeout:               60 * time.Second,
	}

	controller := NewController(ctx, client)
	assert.NotNil(t, controller)

	// Test onAdd
	ip := &inwinv1.IP{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ip-1",
			Namespace: namespace,
		},
		Spec: inwinv1.IPSpec{
			PoolName:        createPool.Name,
			UpdateNamespace: true,
		},
	}
	createIP, err := client.InwinstackV1().IPs(namespace).Create(ip)
	assert.Nil(t, err)

	controller.onAdd(createIP)

	onAddIP, err := client.InwinstackV1().IPs(namespace).Get(ip.Name, metav1.GetOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, onAddIP.Status.Phase, inwinv1.IPActive)
	assert.Equal(t, onAddIP.Status.Address, "172.22.132.150")

	onAddPool, err := client.InwinstackV1().Pools().Get(pool.Name, metav1.GetOptions{})
	assert.Nil(t, err)
	assert.Equal(t, onAddPool.Status.AllocatedIPs, []string{"172.22.132.150"})
	assert.Equal(t, onAddPool.Status.Capacity, 51)
	assert.Equal(t, onAddPool.Status.Allocatable, 50)

	// Test onUpdate
	// TODO: The IP needs to implement onUpdate function.
	controller.onUpdate(createIP, onAddIP)

	onUpdateNs, err := coreClient.CoreV1().Namespaces().Get(ns.Name, metav1.GetOptions{})
	assert.Nil(t, err)
	assert.Equal(t, onUpdateNs.Annotations[constants.AnnKeyNamespaceRefresh], "true")

	// Test onUpdate
	controller.onDelete(onAddIP)

	onDeletePool, err := client.InwinstackV1().Pools().Get(pool.Name, metav1.GetOptions{})
	assert.Nil(t, err)
	assert.Equal(t, onDeletePool.Status.AllocatedIPs, []string{})
	assert.Equal(t, onDeletePool.Status.Capacity, 51)
	assert.Equal(t, onDeletePool.Status.Allocatable, 51)
}
