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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TargetDataset defines which dataset will be processed by DataProcess.
// Under the hood, the dataset's pvc will be mounted to the given mountPath of the DataProcess's containers.
type TargetDatasetWithMountPath struct {
	TargetDataset `json:",inline"`

	// MountPath defines where the Dataset should be mounted in DataProcess's containers.
	// +required
	MountPath string `json:"mountPath"`

	// SubPath defines subpath of the target dataset to mount.
	// +optional
	SubPath string `json:"subPath,omitempty"`
}

// Processor defines the actual processor for DataProcess. Processor can be either of a Job or a Shell script.
type Processor struct {
	Job *ProcessorJob `json:"job,omitempty"`
}

// DataProcessSpec defines the desired state of DataProcess
type DataProcessSpec struct {
	// +requried
	Dataset TargetDatasetWithMountPath `json:"dataset"`

	// +required
	Processor Processor `json:"processor"`
}

type ProcessorJob struct {
	// +required
	Image string `json:"image"`

	// +required
	Script string `json:"script"`

	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DataProcess is the Schema for the dataprocesses API
type DataProcess struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataProcessSpec `json:"spec,omitempty"`
	Status OperationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DataProcessList contains a list of DataProcess
type DataProcessList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataProcess `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataProcess{}, &DataProcessList{})
}
