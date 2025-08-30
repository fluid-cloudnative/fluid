/*
Copyright 2025 The Fluid Authors.

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

// PersistentPodStateSpec defines the desired state of PersistentPodState
type PersistentPodStateSpec struct {
}

type PodState struct {
	NodeName string `json:"nodeName"`
}

// PersistentPodStateStatus defines the observed state of PersistentPodState
type PersistentPodStateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// 	PodStates is the pod name mapping to state.
	PodStates      map[string]PodState `json:"podStates,omitempty"`
	LastUpdateTime *metav1.Time        `json:"lastUpdateTime,omitempty"`
}

// PersistentPodState is the Schema for the PersistentPodState API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
type PersistentPodState struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PersistentPodStateSpec   `json:"spec,omitempty"`
	Status PersistentPodStateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PersistentPodStateList contains a list of PersistentPodState
type PersistentPodStateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PersistentPodState `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PersistentPodState{}, &PersistentPodStateList{})
}
