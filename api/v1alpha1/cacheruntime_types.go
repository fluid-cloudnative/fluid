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

// RuntimeTieredStoreLevel describes the configuration for a single tier in the tiered storage
type RuntimeTieredStoreLevel struct {
	// Medium describes the storage medium type for this tier.
	// Supported types include process memory and various volume types.
	// +optional
	Medium MediumSource `json:"medium,omitempty"`

	// Path is a list of file paths to be used for the cache tier.
	// Multiple paths can be specified to distribute cache across different mount points.
	// For example: ["/mnt/cache1", "/mnt/cache2"].
	// +kubebuilder:validation:MinItems=1
	// +optional
	Path []string `json:"path,omitempty"`

	// Quota is a list of storage quotas for each path in the tier.
	// The length of Quota should match the length of Path.
	// Each quota corresponds to the path at the same index.
	// For example: ["100Gi", "50Gi"] allocates 100GiB to the first path and 50GiB to the second path.
	// +optional
	Quota []resource.Quantity `json:"quota,omitempty"`

	// High is the ratio of high watermark of the tier (e.g., "0.9").
	// When cache usage exceeds this ratio, eviction will be triggered.
	// +optional
	High string `json:"high,omitempty"`

	// Low is the ratio of low watermark of the tier (e.g., "0.7").
	// Eviction will continue until cache usage falls below this ratio.
	// +optional
	Low string `json:"low,omitempty"`
}

// MediumSource describes the storage medium type for tiered store.
// Only one of its members may be specified.
type MediumSource struct {
	// ProcessMemory indicates that process memory should be used as the storage medium.
	// The cache will be stored in the process's memory space.
	// +optional
	ProcessMemory *ProcessMemoryMediumSource `json:"processMemory,omitempty"`

	// Volume indicates that a Kubernetes volume should be used as the storage medium.
	// Supported volume types include hostPath, emptyDir, and ephemeral volumes.
	// +optional
	Volume *VolumeMediumSource `json:"volume,omitempty"`
}

//tieredStore:
//  levels:
//	- quota: 8Gi # quota will add to component.container.resource.request/limit.memory
//	  high: "0.99"
//    low: "0.99"
//	  mediumSource:
//      processMemory: {}

// ProcessMemoryMediumSource describes process memory as a storage medium.
// When specified, cache data will be stored in the process's memory space,
// and the quota will be added to the container's resource requests and limits.
type ProcessMemoryMediumSource struct {
}

//tieredStore:
//  levels:
//	- quota: 8Gi
//	  high: "0.99"
//    low: "0.99"
// 	  path: /dev/shm
//	  mediumSource:
//		emptyDir:{}
//	    # Or one of the following:
//      # ephemeral:
//      #   volumeClaimTemplate:{}
//		# hostPath:{}

// VolumeMediumSource describes a Kubernetes volume as a storage medium.
// Only one of its members may be specified.
type VolumeMediumSource struct {
	// HostPath represents a pre-existing file or directory on the host machine that is directly exposed to the container.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes#hostpath
	// +optional
	HostPath *corev1.HostPathVolumeSource `json:"hostPath,omitempty"`

	// EmptyDir represents a temporary directory that shares a pod's lifetime.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes#emptydir
	// +optional
	EmptyDir *corev1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`

	// Ephemeral represents a volume that is handled by a cluster storage driver.
	// The volume's lifecycle is tied to the pod that defines it.
	// More info: https://kubernetes.io/docs/concepts/storage/ephemeral-volumes/
	// +optional
	Ephemeral *corev1.EphemeralVolumeSource `json:"ephemeral,omitempty"`
}
