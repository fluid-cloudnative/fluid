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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ThinRuntimeProfileSpec defines the desired state of ThinRuntimeProfile
type ThinRuntimeProfileSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The version information that instructs fluid to orchestrate a particular version,
	Version VersionSpec `json:"version,omitempty"`

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
}

// ThinRuntimeProfileStatus defines the observed state of ThinRuntimeProfile
type ThinRuntimeProfileStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// ThinRuntimeProfile is the Schema for the ThinRuntimeProfiles API
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
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
