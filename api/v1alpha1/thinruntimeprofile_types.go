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

type NodePublishSecretPolicy string

const (
	NotMountNodePublishSecret                NodePublishSecretPolicy = "NotMountNodePublishSecret"
	MountNodePublishSecretIfExists           NodePublishSecretPolicy = "MountNodePublishSecretIfExists"
	CopyNodePublishSecretAndMountIfNotExists NodePublishSecretPolicy = "CopyNodePublishSecretAndMountIfNotExists"
)

// ThinRuntimeProfileSpec defines the desired state of ThinRuntimeProfile
type ThinRuntimeProfileSpec struct {
	// file system of thinRuntime
	// +required
	FileSystemType string `json:"fileSystemType"`

	// The component spec of worker
	Worker ThinCompTemplateSpec `json:"worker,omitempty"`

	// The component spec of thinRuntime
	Fuse ThinFuseSpec `json:"fuse,omitempty"`

	// Volumes is the list of Kubernetes volumes that can be mounted by runtime components and/or fuses.
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`

	// NodePublishSecretPolicy describes the policy to decide which to do with node publish secret when mounting an existing persistent volume.
	// +kubebuilder:default=MountNodePublishSecretIfExists
	// +kubebuilder:validation:Enum=NotMountNodePublishSecret;MountNodePublishSecretIfExists;CopyNodePublishSecretAndMountIfNotExists
	NodePublishSecretPolicy NodePublishSecretPolicy `json:"nodePublishSecretPolicy,omitempty"`
}

// ThinRuntimeProfileStatus defines the observed state of ThinRuntimeProfile
type ThinRuntimeProfileStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// ThinRuntimeProfile is the Schema for the ThinRuntimeProfiles API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type ThinRuntimeProfile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ThinRuntimeProfileSpec   `json:"spec,omitempty"`
	Status ThinRuntimeProfileStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ThinRuntimeProfileList contains a list of ThinRuntimeProfile
type ThinRuntimeProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ThinRuntimeProfile `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ThinRuntimeProfile{}, &ThinRuntimeProfileList{})
}
