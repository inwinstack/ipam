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

// IPList is a list of IP.
type IPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []IP `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IP represents a Kubernetes IP Custom Resource.
// The IP will be used as IP.
type IP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   IPSpec   `json:"spec"`
	Status IPStatus `json:"status,omitempty"`
}

// IPSpec is the spec for a IP resource.
type IPSpec struct {
	PoolName        string `json:"poolName"`
	UpdateNamespace bool   `json:"updateNamespace"`
}

type IPPhase string

// These are the valid phases of a IP.
const (
	IPActive      IPPhase = "Active"
	IPFailed      IPPhase = "Failed"
	IPTerminating IPPhase = "Terminating"
)

// IPStatus represents the current state of a IP resource.
type IPStatus struct {
	Phase          IPPhase     `json:"phase"`
	Reason         string      `json:"reason,omitempty"`
	LastUpdateTime metav1.Time `json:"lastUpdateTime"`
	Address        string      `json:"address,omitempty"`
}
