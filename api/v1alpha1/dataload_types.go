/*

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TargetDataset defines the target dataset of the DataLoad
type TargetDataset struct {
	// Name defines name of the target dataset
	Name string `json:"name"`

	// todo(xuzhihao): Namespace may be unnecessary for the reason that we assume DataLoad is in the same namespace with its target Dataset

	// Namespace defines namespace of the target dataset
	Namespace string `json:"namespace,omitempty"`
}

// TargetPath defines the target path of the DataLoad
type TargetPath struct {
	// Path defines path to be load
	Path string `json:"path"`

	// Replicas defines how many replicas will be loaded
	Replicas int32 `json:"replicas,omitempty"`
}

// DataLoadSpec defines the desired state of DataLoad
type DataLoadSpec struct {
	// Dataset defines the target dataset of the DataLoad
	Dataset TargetDataset `json:"dataset,omitempty"`

	// LoadMetadata specifies if the dataload job should load metadata
	LoadMetadata bool `json:"loadMetadata,omitempty"`

	// Target defines target paths that needs to be loaded
	Target []TargetPath `json:"target,omitempty"`

	// Options specifies the extra dataload properties for runtime
	Options map[string]string `json:"options,omitempty"`

	// PodMetadata defines labels and annotations that will be propagated to DataLoad pods
	PodMetadata PodMetadata `json:"podMetadata,omitempty"`

	// +optional
	// Affinity defines affinity for DataLoad pod
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// +optional
	// Tolerations defines tolerations for DataLoad pod
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// +optional
	// NodeSelector defiens node selector for DataLoad pod
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	// SchedulerName sets the scheduler to be used for DataLoad pod
	SchedulerName string `json:"schedulerName,omitempty"`
}

// +kubebuilder:printcolumn:name="Dataset",type="string",JSONPath=`.spec.dataset.name`
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Duration",type="string",JSONPath=`.status.duration`
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:resource:categories={fluid},shortName=load
// +genclient

// DataLoad is the Schema for the dataloads API
type DataLoad struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataLoadSpec    `json:"spec,omitempty"`
	Status OperationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced

// DataLoadList contains a list of DataLoad
type DataLoadList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataLoad `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataLoad{}, &DataLoadList{})
}
