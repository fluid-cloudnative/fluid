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

	//+kubebuilder:default:=Once
	//+kubebuilder:validation:Enum=Once;Cron;OnEvent
	// including Once, Cron, OnEvent
	// +optional
	Policy Policy `json:"policy,omitempty"`

	// The schedule in Cron format, only set when policy is cron, see https://en.wikipedia.org/wiki/Cron.
	// +optional
	Schedule string `json:"schedule,omitempty"`

	// Specifies that the preceding operation in a workflow
	// +optional
	RunAfter *OperationRef `json:"runAfter,omitempty"`

	// TTLSecondsAfterFinished is the time second to clean up data operations after finished or failed
	// +optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty"`
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
