/*
Copyright 2025 The Fluid Author.

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

// +k8s:deepcopy-gen=package
// +groupName=data.fluid.io
package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// CacheRuntimeKind represents the kind of CacheRuntime resource
	CacheRuntimeKind = "CacheRuntime"
)

// GenericCacheRuntimeComponentCommonSpec is a common description of the GenericCacheRuntime component
type GenericCacheRuntimeComponentCommonSpec struct {
	// Disabled determines whether to disable the GenericCacheRuntime component
	// +optional
	Disabled bool `json:"disabled,omitempty"`

	// RuntimeVersion specifies the version information that instructs fluid to orchestrate a particular version
	// +optional
	RuntimeVersion VersionSpec `json:"runtimeVersion,omitempty"`

	// Options defines configurable options for the cache system
	// +optional
	Options map[string]string `json:"options,omitempty"`

	// PodMetadata defines labels and annotations that will be propagated to all components' pods
	// +optional
	PodMetadata PodMetadata `json:"podMetadata,omitempty"`

	// Resources specifies the resources that will be requested by the CacheRuntime component
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Env defines environment variables that will be used by CacheRuntime component
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// VolumeMounts specifies the volumes listed in ".spec.volumes" to mount into the CacheRuntime component's filesystem
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`

	// Args defines arguments to the entrypoint
	// +optional
	// +listType=atomic
	Args []string `json:"args,omitempty"`

	// NodeSelector is a selector that must be true for the component pods to fit on a node
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations specify the pod's tolerations if specified
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// CacheRuntimeMasterSpec defines the specification of the CacheRuntime master component
type CacheRuntimeMasterSpec struct {
	GenericCacheRuntimeComponentCommonSpec `json:",inline"`

	// Replicas is the desired number of replicas of the component
	// If unspecified, defaults to 1
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=1
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
}

// CacheRuntimeWorkerSpec defines the specification of the CacheRuntime worker component
type CacheRuntimeWorkerSpec struct {
	GenericCacheRuntimeComponentCommonSpec `json:",inline"`

	// Replicas is the desired number of replicas of the given template
	// If unspecified, defaults to 1
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=1
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// TieredStore defines tiered storage configuration used by worker
	// +optional
	TieredStore GenericCacheRuntimeTieredStore `json:"tieredStore,omitempty"`
}

// CacheRuntimeClientSpec defines the specification of the CacheRuntime client component
type CacheRuntimeClientSpec struct {
	GenericCacheRuntimeComponentCommonSpec `json:",inline"`

	// TieredStore defines tiered storage configuration used by worker
	// +optional
	TieredStore GenericCacheRuntimeTieredStore `json:"tieredStore,omitempty"`

	// CleanPolicy decides when to clean CacheFS Fuse pods
	// Currently Fluid supports two policies: OnDemand and OnRuntimeDeleted
	// OnDemand cleans fuse pod once the fuse pod on some node is not needed
	// OnRuntimeDeleted cleans fuse pod only when the cache runtime is deleted
	// Defaults to OnRuntimeDeleted
	// +kubebuilder:validation:Enum=OnRuntimeDeleted;"";OnDemand
	// +kubebuilder:default=OnRuntimeDeleted
	// +optional
	CleanPolicy FuseCleanPolicy `json:"cleanPolicy,omitempty"`
}

// CacheRuntimeSpec defines the desired state of CacheRuntime
type CacheRuntimeSpec struct {
	// RuntimeClassName is the name of the RuntimeClass required by the CacheRuntime
	RuntimeClassName string `json:"runtimeClassName,omitempty"`

	// Master defines the component spec of master
	// +optional
	Master CacheRuntimeMasterSpec `json:"master,omitempty"`

	// Worker defines the component spec of worker
	// +optional
	Worker CacheRuntimeWorkerSpec `json:"worker,omitempty"`

	// Client defines the component spec of client
	// +optional
	Client CacheRuntimeClientSpec `json:"client,omitempty"`

	// Options defines configurable options for the cache system
	// +optional
	Options map[string]string `json:"options,omitempty"`

	// PodMetadata defines labels and annotations that will be propagated to all components' pods
	// +optional
	PodMetadata PodMetadata `json:"podMetadata,omitempty"`

	// ImagePullSecrets specifies secrets that will be used to pull images
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// Volumes is the list of Kubernetes volumes that can be mounted by the cache runtime components
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",priority=0
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase",priority=0
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:resource:categories={fluid}
// +genclient

// CacheRuntime is the Schema for the CacheRuntimes API
type CacheRuntime struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CacheRuntimeSpec          `json:"spec,omitempty"`
	Status GenericCacheRuntimeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced

// CacheRuntimeList contains a list of CacheRuntime
type CacheRuntimeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CacheRuntime `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CacheRuntime{}, &CacheRuntimeList{})
}
