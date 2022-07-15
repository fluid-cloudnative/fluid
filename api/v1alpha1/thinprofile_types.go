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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ThinProfileSpec defines the desired state of ThinProfile
type ThinProfileSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Version VersionSpec `json:"version,omitempty"`

	// Environment variables that will be used by thinRuntime Fuse
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Command that will be passed to thinRuntime Fuse
	Command []string `json:"command,omitempty"`

	// Arguments that will be passed to thinRuntime Fuse
	Args []string `json:"args,omitempty"`

	// Options configurable options of FUSE client, performance parameters usually.
	// will be merged with Dataset.spec.mounts.options into fuse pod.
	Options map[string]string `json:"options,omitempty"`

	// Resources that will be requested by thinRuntime Fuse.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// ThinProfileStatus defines the observed state of ThinProfile
type ThinProfileStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ThinProfile is the Schema for the thinprofiles API
type ThinProfile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ThinProfileSpec   `json:"spec,omitempty"`
	Status ThinProfileStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ThinProfileList contains a list of ThinProfile
type ThinProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ThinProfile `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ThinProfile{}, &ThinProfileList{})
}
