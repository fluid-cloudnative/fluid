/*
Copyright 2021 The Fluid Authors.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DataBackupSpec defines the desired state of DataBackup
type DataBackupSpec struct {
	// Dataset defines the target dataset of the DataBackup
	Dataset string `json:"dataset,omitempty"`
	// BackupPath defines the target path to save data of the DataBackup
	BackupPath string `json:"backupPath,omitempty"`
	// Manage the user to run Alluxio DataBackup
	RunAs *User `json:"runAs,omitempty"`
	// Specifies that the preceding operation in a workflow
	// +optional
	RunAfter *OperationRef `json:"runAfter,omitempty"`
	// TTLSecondsAfterFinished is the time second to clean up data operations after finished or failed
	// +optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty"`
}

// +kubebuilder:printcolumn:name="Dataset",type="string",JSONPath=`.spec.dataset`
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Path",type="string",JSONPath=`.status.infos.BackupLocationPath`
// +kubebuilder:printcolumn:name="NodeName",type="string",JSONPath=`.status.infos.BackupLocationNodeName`
// +kubebuilder:printcolumn:name="Duration",type="string",JSONPath=`.status.duration`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:resource:categories={fluid},shortName=backup
// +genclient

// DataBackup is the Schema for the backup API
type DataBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataBackupSpec  `json:"spec,omitempty"`
	Status OperationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced

// DataBackupList contains a list of DataBackup
type DataBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataBackup{}, &DataBackupList{})
}
