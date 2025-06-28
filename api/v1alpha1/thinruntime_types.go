/*
Copyright 2022 The Fluid Author.

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

const (
	ThinRuntimeKind = "ThinRuntime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ThinRuntimeSpec defines the desired state of ThinRuntime
type ThinRuntimeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The specific runtime profile name, empty value is used for handling datasets which mount another dataset
	ThinRuntimeProfileName string `json:"profileName,omitempty"`

	// The component spec of worker
	Worker ThinCompTemplateSpec `json:"worker,omitempty"`

	// The component spec of thinRuntime
	Fuse ThinFuseSpec `json:"fuse,omitempty"`

	// Tiered storage
	TieredStore TieredStore `json:"tieredstore,omitempty"`

	// The replicas of the worker, need to be specified
	Replicas int32 `json:"replicas,omitempty"`

	// Manage the user to run Runtime
	RunAs *User `json:"runAs,omitempty"`

	// Disable monitoring for Runtime
	// Prometheus is enabled by default
	// +optional
	DisablePrometheus bool `json:"disablePrometheus,omitempty"`

	// Volumes is the list of Kubernetes volumes that can be mounted by runtime components and/or fuses.
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`

	// RuntimeManagement defines policies when managing the runtime
	// +optional
	RuntimeManagement RuntimeManagement `json:"management,omitempty"`

	// ImagePullSecrets that will be used to pull images
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

// ThinCompTemplateSpec is a description of the thinRuntime components
type ThinCompTemplateSpec struct {
	// Image for thinRuntime fuse
	Image string `json:"image,omitempty"`

	// Image for thinRuntime fuse
	ImageTag string `json:"imageTag,omitempty"`

	// One of the three policies: `Always`, `IfNotPresent`, `Never`
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	// ImagePullSecrets that will be used to pull images
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// Replicas is the desired number of replicas of the given template.
	// If unspecified, defaults to 1.
	// +kubebuilder:validation:Minimum=1
	// replicas is the min replicas of dataset in the cluster
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Ports used thinRuntime
	// +optional
	Ports []corev1.ContainerPort `json:"ports,omitempty"`

	// Resources that will be requested by thinRuntime component.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Environment variables that will be used by thinRuntime component.
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Enabled or Disabled for the components.
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// NodeSelector is a selector
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// VolumeMounts specifies the volumes listed in ".spec.volumes" to mount into runtime component's filesystem.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`

	// livenessProbe of thin fuse pod
	// +optional
	LivenessProbe *corev1.Probe `json:"livenessProbe,omitempty"`

	// readinessProbe of thin fuse pod
	// +optional
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty"`

	// Whether to use hostnetwork or not
	// +kubebuilder:validation:Enum=HostNetwork;"";ContainerNetwork
	// +optional
	NetworkMode NetworkMode `json:"networkMode,omitempty"`
}

type ThinFuseSpec struct {
	// Image for thinRuntime fuse
	Image string `json:"image,omitempty"`

	// Image for thinRuntime fuse
	ImageTag string `json:"imageTag,omitempty"`

	// One of the three policies: `Always`, `IfNotPresent`, `Never`
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	// ImagePullSecrets that will be used to pull images
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// Ports used thinRuntime
	// +optional
	Ports []corev1.ContainerPort `json:"ports,omitempty"`

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

	// NodeSelector is a selector which must be true for the fuse client to fit on a node,
	// this option only effect when global is enabled
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// CleanPolicy decides when to clean thinRuntime Fuse pods.
	// Currently Fluid supports two policies: OnDemand and OnRuntimeDeleted
	// OnDemand cleans fuse pod once the fuse pod on some node is not needed
	// OnRuntimeDeleted cleans fuse pod only when the cache runtime is deleted
	// Defaults to OnDemand
	// +optional
	CleanPolicy FuseCleanPolicy `json:"cleanPolicy,omitempty"`

	// Whether to use hostnetwork or not
	// +kubebuilder:validation:Enum=HostNetwork;"";ContainerNetwork
	// +optional
	NetworkMode NetworkMode `json:"networkMode,omitempty"`

	// livenessProbe of thin fuse pod
	// +optional
	LivenessProbe *corev1.Probe `json:"livenessProbe,omitempty"`

	// readinessProbe of thin fuse pod
	// +optional
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty"`

	// VolumeMounts specifies the volumes listed in ".spec.volumes" to mount into the thinruntime component's filesystem.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`

	// Lifecycle describes actions that the management system should take in response to container lifecycle events.
	Lifecycle *corev1.Lifecycle `json:"lifecycle,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +genclient

// ThinRuntime is the Schema for the thinruntimes API
type ThinRuntime struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ThinRuntimeSpec `json:"spec,omitempty"`
	Status RuntimeStatus   `json:"status,omitempty"`
}

func (in *ThinRuntime) Replicas() int32 {
	return in.Spec.Replicas
}

func (in *ThinRuntime) GetStatus() *RuntimeStatus {
	return &in.Status
}

//+kubebuilder:object:root=true

// ThinRuntimeList contains a list of ThinRuntime
type ThinRuntimeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ThinRuntime `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ThinRuntime{}, &ThinRuntimeList{})
}
