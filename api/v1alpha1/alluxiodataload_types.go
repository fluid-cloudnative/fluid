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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AlluxioDataLoadSpec defines the desired state of AlluxioDataLoad
type AlluxioDataLoadSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Name of the dataset that will be prefetched
	// +kubebuilder:validation:MinLength=1
	// +required
	DatasetName string `json:"datasetName"`

	// Mount path of the dataset in Alluxio. Defaults to /{datasetName} if not specified. (e.g. /my-dataset/cifar10)
	// +optional
	Path string `json:"path,omitempty"`

	// Specifies the number of slots per worker used in hostfile.
	// Defaults to 2.
	// +optional
	SlotsPerNode *int32 `json:"slotsPerNode,omitempty"`
}

// AlluxioDataLoadStatus defines the observed state of AlluxioDataLoad
type AlluxioDataLoadStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The latest available observation of a dataload's running phase.
	// One of the four phases: `Pending`, `Loading`, `Complete` and `Failed`
	Phase common.DataloadPhase `json:"phase"`

	// The latest available observations of an object's current state.
	// +optional
	Conditions []DataloadCondition `json:"conditions"`
}

// DataloadCondition describes current state of a Dataload.
type DataloadCondition struct {
	// Type of Dataload condition, Complete or Failed.
	Type common.DataloadConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
	// The last time this condition was updated.
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +genclient

// AlluxioDataLoad is the Schema for the alluxiodataloads API
type AlluxioDataLoad struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlluxioDataLoadSpec   `json:"spec,omitempty"`
	Status AlluxioDataLoadStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AlluxioDataLoadList contains a list of AlluxioDataLoad
type AlluxioDataLoadList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AlluxioDataLoad `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AlluxioDataLoad{}, &AlluxioDataLoadList{})
}
