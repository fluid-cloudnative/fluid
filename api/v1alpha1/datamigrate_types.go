/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/
package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DataMigrateSpec defines the desired state of DataMigrate
type DataMigrateSpec struct {
	// The version information that instructs fluid to orchestrate a particular version for data migrate.
	// +optional
	VersionSpec `json:",inline,omitempty"`

	// data to migrate source, including dataset and external storage
	From DataToMigrate `json:"from"`

	// data to migrate destination, including dataset and external storage
	To DataToMigrate `json:"to"`

	// if dataMigrate blocked dataset usage, default is false
	// +optional
	Block bool `json:"block,omitempty"`

	// using which runtime to migrate data; if none, take dataset runtime as default
	// +optional
	RuntimeType string `json:"runtimeType,omitempty"`

	// options for migrate, different for each runtime
	// +optional
	Options map[string]string `json:"options,omitempty"`

	//+kubebuilder:default:=Once
	//+kubebuilder:validation:Enum=Once;Cron;OnEvent
	// policy for migrate, including Once, Cron, OnEvent
	// +optional
	Policy Policy `json:"policy,omitempty"`

	// The schedule in Cron format, only set when policy is cron, see https://en.wikipedia.org/wiki/Cron.
	// +optional
	Schedule string `json:"schedule,omitempty"`

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

	// Specifies that the preceding operation in a workflow
	// +optional
	RunAfter *OperationRef `json:"runAfter,omitempty"`

	// TTLSecondsAfterFinished is the time second to clean up data operations after finished or failed
	// +optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty"`
}

type DataToMigrate struct {
	// dataset to migrate
	DataSet *DatasetToMigrate `json:"dataset,omitempty"`

	// external storage for data migrate
	ExternalStorage *ExternalStorage `json:"externalStorage,omitempty"`
}

type DatasetToMigrate struct {
	// name of dataset
	Name string `json:"name"`

	// namespace of dataset
	Namespace string `json:"namespace"`

	// path to migrate
	Path string `json:"path,omitempty"`
}

type ExternalStorage struct {
	// type of external storage, including s3, oss, gcs, ceph, nfs, pvc, etc. (related to runtime)
	URI string `json:"uri"`

	// encrypt info for external storage
	// +optional
	EncryptOptions []EncryptOption `json:"encryptOptions,omitempty"`
}

// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Duration",type="string",JSONPath=`.status.duration`
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:resource:categories={fluid},shortName=migrate
// +genclient

// DataMigrate is the Schema for the datamigrates API
type DataMigrate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataMigrateSpec `json:"spec,omitempty"`
	Status OperationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced

// DataMigrateList contains a list of DataMigrate
type DataMigrateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataMigrate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataMigrate{}, &DataMigrateList{})
}
