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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PoolList is a list of Pool.
type PoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Pool `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Pool represents a Kubernetes Pool Custom Resource.
// The Pool will be used as IP pools.
type Pool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   PoolSpec   `json:"spec"`
	Status PoolStatus `json:"status,omitempty"`
}

// PoolSpec is the spec for a pool resource.
type PoolSpec struct {
	Address                   string   `json:"address"`
	IgnoreNamespaces          []string `json:"ignoreNamespaces"`
	IgnoreNamespaceAnnotation bool     `json:"ignoreNamespaceAnnotation"`
	AutoAssignToNamespace     bool     `json:"autoAssignToNamespace"`
}

type PoolPhase string

// These are the valid phases of a pool.
const (
	PoolActive      PoolPhase = "Active"
	PoolFailed      PoolPhase = "Failed"
	PoolTerminating PoolPhase = "Terminating"
)

// PoolStatus represents the current state of a pool resource.
type PoolStatus struct {
	Phase          PoolPhase   `json:"phase"`
	Reason         string      `json:"reason,omitempty"`
	LastUpdateTime metav1.Time `json:"lastUpdateTime"`
	AllocatedIPs   []string    `json:"allocatedIPs"`
	Capacity       int         `json:"capacity"`
}
