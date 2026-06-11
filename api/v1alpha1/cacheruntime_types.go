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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// CacheRuntimeKind represents the kind of CacheRuntime resource
	CacheRuntimeKind = "CacheRuntime"
)

// RuntimeComponentCommonSpec describes the common configuration for CacheRuntime components
type RuntimeComponentCommonSpec struct {
	// Disabled indicates whether the component should be disabled.
	// If set to true, the component will not be created.
	// +optional
	Disabled bool `json:"disabled,omitempty"`

	// RuntimeVersion is the version information that instructs Fluid to orchestrate a particular version of the runtime.
	// +optional
	RuntimeVersion VersionSpec `json:"runtimeVersion,omitempty"`

	// Options is a set of key-value pairs that provide additional configuration for the cache system.
	// +optional
	Options map[string]string `json:"options,omitempty"`

	// PodMetadata contains labels and annotations that will be propagated to the component's pods.
	// +optional
	PodMetadata PodMetadata `json:"podMetadata,omitempty"`

	// Resources describes the compute resource requirements.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Env is a list of environment variables to set in the container.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// VolumeMounts are the volumes to mount into the container's filesystem.
	// Cannot be updated.
	// +optional
	// +patchMergeKey=mountPath
	// +patchStrategy=merge
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty" patchStrategy:"merge" patchMergeKey:"mountPath"`

	// Args are arguments to the entrypoint.
	// The container image's CMD is used if this is not provided.
	// +optional
	// +listType=atomic
	Args []string `json:"args,omitempty"`

	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	// +mapType=atomic
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations are the pod's tolerations.
	// If specified, the pod's tolerations.
	// More info: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// CacheRuntimeMasterSpec describes the desired state of CacheRuntime master component
type CacheRuntimeMasterSpec struct {
	RuntimeComponentCommonSpec `json:",inline"`

	// Replicas is the desired number of replicas of the master component.
	// If unspecified, defaults to 1.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=1
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
}

// CacheRuntimeWorkerSpec describes the desired state of CacheRuntime worker component
type CacheRuntimeWorkerSpec struct {
	RuntimeComponentCommonSpec `json:",inline"`

	// Replicas is the desired number of replicas of the worker component.
	// If unspecified, defaults to 1.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=1
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// TieredStore describes the tiered storage configuration used by the worker component.
	// +optional
	TieredStore RuntimeTieredStore `json:"tieredStore,omitempty"`
}

// CacheRuntimeClientSpec describes the desired state of CacheRuntime client component
type CacheRuntimeClientSpec struct {
	RuntimeComponentCommonSpec `json:",inline"`

	// TieredStore describes the tiered storage configuration used by the client component.
	// +optional
	TieredStore RuntimeTieredStore `json:"tieredStore,omitempty"`

	// CleanPolicy determines when to clean up the FUSE client pods.
	// Currently supports two policies:
	// - "OnDemand": Clean up the FUSE pod when it is no longer needed on a node
	// - "OnRuntimeDeleted": Clean up the FUSE pod only when the CacheRuntime is deleted
	// Defaults to "OnRuntimeDeleted".
	// +kubebuilder:validation:Enum=OnRuntimeDeleted;OnDemand
	// +kubebuilder:default=OnRuntimeDeleted
	// +optional
	CleanPolicy FuseCleanPolicy `json:"cleanPolicy,omitempty"`
}

// CacheRuntimeSpec describes the desired state of CacheRuntime
type CacheRuntimeSpec struct {
	// RuntimeClassName is the name of the CacheRuntimeClass to use for this runtime.
	// The CacheRuntimeClass defines the implementation details of the cache runtime.
	// +kubebuilder:validation:Required
	RuntimeClassName string `json:"runtimeClassName"`

	// Master is the desired state of the master component.
	// +optional
	Master CacheRuntimeMasterSpec `json:"master,omitempty"`

	// Worker is the desired state of the worker component.
	// +optional
	Worker CacheRuntimeWorkerSpec `json:"worker,omitempty"`

	// Client is the desired state of the client component.
	// +optional
	Client CacheRuntimeClientSpec `json:"client,omitempty"`

	// Options is a set of key-value pairs that provide additional configuration for the cache system.
	// These options will be propagated to all components.
	// +optional
	Options map[string]string `json:"options,omitempty"`

	// PodMetadata contains labels and annotations that will be propagated to all component pods.
	// +optional
	PodMetadata PodMetadata `json:"podMetadata,omitempty"`

	// ImagePullSecrets is an optional list of references to secrets in the same namespace
	// to use for pulling any of the images used by this PodSpec.
	// More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// Volumes is the list of volumes that can be mounted by containers belonging to the cache runtime components.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Volumes []corev1.Volume `json:"volumes,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name"`
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

	Spec   CacheRuntimeSpec   `json:"spec,omitempty"`
	Status CacheRuntimeStatus `json:"status,omitempty"`
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

// RuntimeTieredStore describes the tiered storage configuration for cache runtime
type RuntimeTieredStore struct {
	// Levels is the list of cache storage tiers from highest priority to lowest.
	// Each tier can use different storage media (e.g., memory, SSD, HDD).
	// +optional
	Levels []RuntimeTieredStoreLevel `json:"levels,omitempty"`
}

// ProcessMemoryMediumSource describes process memory as a storage medium.
// Cache data will be stored in the process's memory space.
type ProcessMemoryMediumSource struct {
	// Quota specifies the amount of memory to allocate for caching.
	// This value will be added to the container's resource requests and limits.
	// Example: "4Gi" allocates 4GiB memory for cache.
	// +kubebuilder:validation:Required
	Quota resource.Quantity `json:"quota"`
}

// HostPathMediumSource describes hostPath volumes as a storage medium.
// Multiple paths can be configured to distribute cache across different disks.
type HostPathMediumSource struct {
	// Paths is a list of file paths on the host machine to be used for caching.
	// Multiple paths allow distributing cache across different mount points or disks.
	// Example: ["/mnt/cache1", "/mnt/cache2"]
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:Required
	Paths []string `json:"paths"`

	// Quotas is a list of storage quotas corresponding to each path.
	// The length of Quotas must match the length of Paths.
	// Each quota defines the maximum cache size for the corresponding path.
	// Example: ["100Gi", "50Gi"] allocates 100GiB to the first path and 50GiB to the second.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:Required
	Quotas []resource.Quantity `json:"quotas"`

	// Type specifies the type of hostPath volume.
	// Defaults to empty string (no validation).
	// More info: https://kubernetes.io/docs/concepts/storage/volumes#hostpath
	// +optional
	Type *corev1.HostPathType `json:"type,omitempty"`
}

// EmptyDirMediumSource describes emptyDir volume as a storage medium.
// Can be backed by node's default storage or memory (tmpfs).
type EmptyDirMediumSource struct {
	// Quota specifies the maximum storage capacity for the emptyDir volume.
	// For Memory medium, this limit is also constrained by the sum of container memory limits.
	// The actual limit = min(Quota, sum of container memory limits).
	// Example: "100Gi" limits the emptyDir to 100GiB.
	// +kubebuilder:validation:Required
	Quota resource.Quantity `json:"quota"`

	// medium represents what type of storage medium should back this directory.
	// The default is "" which means to use the node's default medium.
	// Must be an empty string (default) or Memory.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes#emptydir
	// +optional
	Medium corev1.StorageMedium `json:"medium,omitempty"`
}

// RuntimeTieredStoreLevel describes the configuration for a single tier in the tiered storage.
// Each tier can use different storage media (e.g., memory, SSD, HDD).
// Only one of ProcessMemory, EmptyDir, or HostPath should be specified.
type RuntimeTieredStoreLevel struct {
	// ProcessMemory indicates that process memory should be used as one storage medium.
	// When specified, cache data will be stored in the process's memory space,
	// and the quota will be added to the container's resource requests and limits.
	// +optional
	ProcessMemory *ProcessMemoryMediumSource `json:"processMemory,omitempty"`

	// EmptyDir indicates that an emptyDir volume should be used as the storage medium.
	// +optional
	EmptyDir *EmptyDirMediumSource `json:"emptyDir,omitempty"`

	// HostPath indicates that one or more hostPath volumes should be used as a storage medium.
	// +optional
	HostPath *HostPathMediumSource `json:"hostPath,omitempty"`

	// High is the ratio of high watermark of the tier (e.g., "0.9").
	// When cache usage exceeds this ratio, eviction will be triggered.
	// +optional
	High string `json:"high,omitempty"`

	// Low is the ratio of low watermark of the tier (e.g., "0.7").
	// Eviction will continue until cache usage falls below this ratio.
	// +optional
	Low string `json:"low,omitempty"`
}

func (runtime *CacheRuntime) GetStatus() *CacheRuntimeStatus {
	return &runtime.Status
}
