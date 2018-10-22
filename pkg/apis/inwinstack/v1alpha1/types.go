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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IPPoolList is a list of Pool.
type IPPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []IPPool `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IPPool represents a Kubernetes IPPool Custom Resource.
// The Pool will be used as IP pools.
type IPPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   IPPoolSpec   `json:"spec"`
	Status IPPoolStatus `json:"status,omitempty"`
}

// IPPoolSpec is the spec for a Pool resource.
type IPPoolSpec struct {
	Address          string   `json:"address"`
	IgnoreNamespaces []string `json:"ignoreNamespaces"`
}

// IPPoolStatus represents the current state of a resource.
type IPPoolStatus struct {
	Message        string      `json:"message"`
	LastUpdateTime metav1.Time `json:"lastUpdateTime"`
	AllocatedIPs   []string    `json:"allocatedIPs"`
	AllocatableIPs []string    `json:"allocatableIPs"`
	Capacity       int         `json:"capacity"`
}
