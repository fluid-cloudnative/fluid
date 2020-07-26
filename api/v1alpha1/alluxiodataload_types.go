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
	"github.com/cloudnativefluid/fluid/pkg/common"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AlluxioDataLoadSpec defines the desired state of AlluxioDataLoad
type AlluxioDataLoadSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// DatasetName is an example field of AlluxioDataLoad. Edit AlluxioDataLoad_types.go to remove/update
	// +kubebuilder:validation:MinLength=1
	// +required
	DatasetName string `json:"datasetName"`

	// the Path in alluxio
	// +optional
	Path string `json:"path,omitempty"`

	// Specifies the number of slots per worker used in hostfile.
	// Defaults to 1.
	// +optional
	SlotsPerNode *int32 `json:"slotsPerNode,omitempty"`
}

// AlluxioDataLoadStatus defines the observed state of AlluxioDataLoad
type AlluxioDataLoadStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase common.DataloadPhase `json:"phase"`

	// The latest available observations of an object's current state.
	// +optional
	Conditions []DataloadCondition `json:"conditions"`
}

// DataloadCondition describes current state of a Dataload.
type DataloadCondition struct {
	// Type of Dataload condition, Complete or Failed.
	Type common.DataloadConditionType
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus
	// Last time the condition was checked.
	// +optional
	LastProbeTime metav1.Time
	// Last time the condition transit from one status to another.
	// +optional
	LastTransitionTime metav1.Time
	// (brief) reason for the condition's last transition.
	// +optional
	Reason string
	// Human readable message indicating details about last transition.
	// +optional
	Message string
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
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
