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

package k8sutil

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"fmt"
	"reflect"

	inwinv1 "github.com/inwinstack/ipam-operator/pkg/apis/inwinstack/v1"
	"github.com/inwinstack/ipam-operator/pkg/constants"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func GetRestConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("master", kubeconfig)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}

	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func SetPoolOwnerReference(refs []metav1.OwnerReference, pool *inwinv1.Pool) {
	refs = []metav1.OwnerReference{
		*metav1.NewControllerRef(pool, schema.GroupVersionKind{
			Group:   inwinv1.SchemeGroupVersion.Group,
			Version: inwinv1.SchemeGroupVersion.Version,
			Kind:    reflect.TypeOf(inwinv1.Pool{}).Name(),
		}),
	}
}

func NewIPWithNamespace(ns *v1.Namespace, poolName string) *inwinv1.IP {
	return &inwinv1.IP{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s", uuid.NewUUID()),
			Namespace: ns.Name,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(ns, schema.GroupVersionKind{
					Group:   v1.SchemeGroupVersion.Group,
					Version: v1.SchemeGroupVersion.Version,
					Kind:    reflect.TypeOf(v1.Namespace{}).Name(),
				}),
			},
		},
		Spec: inwinv1.IPSpec{
			PoolName:        poolName,
			UpdateNamespace: true,
		},
	}
}

func NewDefaultPool(addr string, namespaces []string) *inwinv1.Pool {
	return &inwinv1.Pool{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.DefaultPoolName,
		},
		Spec: inwinv1.PoolSpec{
			Address:                   addr,
			IgnoreNamespaces:          namespaces,
			IgnoreNamespaceAnnotation: false,
			AutoAssignToNamespace:     true,
		},
	}
}
