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
	// Job represents a processor which runs DataProcess as a job.
	// +optional
	Job *JobProcessor `json:"job,omitempty"`

	// Shell represents a processor which executes shell script
	Shell *ShellProcessor `json:"shell,omitempty"`
}

type JobProcessor struct {
	// PodTemplate defines Pod specification of the DataProcess job.
	// +optional
	PodTemplate *corev1.PodTemplate `json:"podTemplate,omitempty"`
}

type ShellProcessor struct {
	// Script specifies a shell script that will be executed in DataProcess.
	// +requried
	Script string `json:"script"`

	// Image specifies the container image where the shell script will be executed.
	// +required
	Image string `json:"image"`

	// ServiceAccountName defiens the serviceAccountName of the container
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

// DataProcessSpec defines the desired state of DataProcess
type DataProcessSpec struct {
	// Dataset specifies the target dataset and its mount path.
	// +requried
	Dataset TargetDatasetWithMountPath `json:"dataset"`

	// Processor specify how to process data.
	// +required
	Processor Processor `json:"processor"`
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