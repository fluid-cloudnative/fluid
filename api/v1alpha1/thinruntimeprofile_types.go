/*
Copyright 2022 The Fluid Authors.

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
