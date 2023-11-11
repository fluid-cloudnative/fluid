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
	// ServiceAccountName defiens the serviceAccountName of the container
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// PodMetadata defines labels and annotations on the processor pod.
	// +optional
	PodMetadata PodMetadata `json:"podMetadata,omitempty"`

	// Job represents a processor which runs DataProcess as a job.
	// +optional
	Job *JobProcessor `json:"job,omitempty"`

	// Shell represents a processor which executes shell script
	Script *ScriptProcessor `json:"script,omitempty"`
}

type JobProcessor struct {
	// PodSpec defines Pod specification of the DataProcess job.
	// +optional
	PodSpec *corev1.PodSpec `json:"podSpec,omitempty"`
}

type ScriptProcessor struct {
	// VersionSpec specifies the container's image info.
	VersionSpec `json:",inline,omitempty"`

	// RestartPolicy specifies the processor job's restart policy. Only "Never", "OnFailure" is allowed.
	// +optional
	// +kubebuilder:default="Never"
	// +kubebuilder:validation:Enum=Never;OnFailure
	RestartPolicy corev1.RestartPolicy `json:"restartPolicy,omitempty"`

	// Entrypoint command for ScriptProcessor.
	// +optional
	Command []string `json:"command,omitempty"`

	// Script source for ScriptProcessor
	// +required
	Source string `json:"source"`

	// List of environment variables to set in the container.
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Pod volumes to mount into the container's filesystem.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`

	// List of volumes that can be mounted by containers belonging to the pod.
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`
}

// DataProcessSpec defines the desired state of DataProcess
type DataProcessSpec struct {
	// Dataset specifies the target dataset and its mount path.
	// +requried
	Dataset TargetDatasetWithMountPath `json:"dataset"`

	// Processor specify how to process data.
	// +required
	Processor Processor `json:"processor"`

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
