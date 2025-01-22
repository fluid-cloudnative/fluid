/*
Copyright 2020 The Fluid Author.

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

// +groupName=data
package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced

// CacheRuntimeList contains a list of CacheRuntime
type CacheRuntimeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CacheRuntime `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.runtimeComponentStatuses.worker.currentScheduledReplicas
// +kubebuilder:printcolumn:name="Ready Masters",type="integer",JSONPath=`.status.runtimeComponentStatuses.master.readyReplicas`,priority=10
// +kubebuilder:printcolumn:name="Desired Masters",type="integer",JSONPath=`.status.runtimeComponentStatuses.master.desiredScheduledReplicas`,priority=10
// +kubebuilder:printcolumn:name="Master Phase",type="string",JSONPath=`.status.runtimeComponentStatuses.master.phase`,priority=0
// +kubebuilder:printcolumn:name="Ready Workers",type="integer",JSONPath=`.status.runtimeComponentStatuses.worker.readyReplicas`,priority=10
// +kubebuilder:printcolumn:name="Desired Workers",type="integer",JSONPath=`.status.runtimeComponentStatuses.worker.desiredScheduledReplicas`,priority=10
// +kubebuilder:printcolumn:name="Worker Phase",type="string",JSONPath=`.status.runtimeComponentStatuses.worker.phase`,priority=0
// +kubebuilder:printcolumn:name="Ready Clients",type="integer",JSONPath=`.status.runtimeComponentStatuses.client.readyReplicas`,priority=10
// +kubebuilder:printcolumn:name="Desired Clients",type="integer",JSONPath=`.status.runtimeComponentStatuses.client.desiredScheduledReplica`,priority=10
// +kubebuilder:printcolumn:name="Client Phase",type="string",JSONPath=`.status.runtimeComponentStatuses.client.phase`,priority=0
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`,priority=0
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:resource:categories={fluid},shortName=Cache
// +genclient

// CacheRuntime is the Schema for the CacheRuntimes API
type CacheRuntime struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CacheRuntimeSpec   `json:"spec,omitempty"`
	Status CacheRuntimeStatus `json:"status,omitempty"`
}

// CacheRuntimeSpec defines the desired state of CacheRuntime
type CacheRuntimeSpec struct {
	// RuntimeClassName is the name of the RuntimeClass required by the cacheRuntime.
	RuntimeClassName string `json:"runtimeClassName,omitempty"`

	// The component spec of master
	// +optional
	Master CacheRuntimeMasterSpec `json:"master,omitempty"`

	// The component spec of worker
	// +optional
	Worker CacheRuntimeWorkerSpec `json:"worker,omitempty"`

	// The component spec of worker group
	// +optional
	WorkerGroup []CacheRuntimeWorkerSpec `json:"workerGroup,omitempty"`

	// The component spec of client
	// +optional
	Client CacheRuntimeClientSpec `json:"client,omitempty"`

	// PodMetadata defines labels and annotations that will be propagated to the all components' pods
	// +optional
	PodMetadata PodMetadata `json:"podMetadata,omitempty"`

	// Options defines the configurable options for Cache system.
	// +optional
	Options map[string]string `json:"options,omitempty"`

	// Volumes is the list of Kubernetes volumes that can be mounted by the cache runtime components.
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`

	// NetworkMode indicates whether to use ContainerNetwork or not
	// +kubebuilder:validation:Enum=HostNetwork;"";ContainerNetwork
	// +kubebuilder:default=ContainerNetwork
	// +optional
	NetworkMode NetworkMode `json:"networkMode,omitempty"`

	// RuntimeManagement defines policies when managing the runtime
	// +optional
	RuntimeManagement RuntimeManagement `json:"runtimeManagement,omitempty"`
}

// CacheRuntimeMasterSpec is a description of the CacheRuntime master component
type CacheRuntimeMasterSpec struct {
	CacheRuntimeComponentCommonSpec `json:",inline"`

	// Replicas is the desired number of replicas of the component.
	// If unspecified, defaults to 1.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=1
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
}

// CacheRuntimeWorkerSpec is a description of the CacheRuntime worker component
type CacheRuntimeWorkerSpec struct {
	CacheRuntimeComponentCommonSpec `json:",inline"`

	// Replicas is the desired number of replicas of the given template.
	// If unspecified, defaults to 1.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=1
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Tiered storage used by worker
	// +optional
	TieredStore TieredStore `json:"tieredStore,omitempty"`
}

// CacheRuntimeClientSpec is a description of the CacheRuntime client component
type CacheRuntimeClientSpec struct {
	CacheRuntimeComponentCommonSpec `json:",inline"`

	// CleanPolicy decides when to clean CacheFS Fuse pods.
	// Currently Fluid supports two policies: OnDemand and OnRuntimeDeleted
	// OnDemand cleans fuse pod once th fuse pod on some node is not needed
	// OnRuntimeDeleted cleans fuse pod only when the cache runtime is deleted
	// Defaults to OnRuntimeDeleted
	// +kubebuilder:validation:Enum=OnRuntimeDeleted;"";OnDemand
	// +kubebuilder:default=OnRuntimeDeleted
	// +optional
	CleanPolicy FuseCleanPolicy `json:"cleanPolicy,omitempty"`
}

// CacheRuntimeComponentCommonSpec is a common description of the CacheRuntime component
type CacheRuntimeComponentCommonSpec struct {
	// If disable CacheRuntime component
	// +optional
	Disabled bool `json:"disabled,omitempty"`

	// WorkloadType is the type of workload, a default workload type is defined in the CacheRuntimeClass
	// +optional
	WorkloadType metav1.TypeMeta `json:"workloadType,omitempty"`

	// The version information that instructs fluid to orchestrate a particular version.
	// +optional
	RuntimeVersion VersionSpec `json:"runtimeVersion,omitempty"`

	// Options is configurable options for Cache System.
	// +optional
	Options map[string]string `json:"options,omitempty"`

	// PodMetadata defines labels and annotations that will be propagated to all CacheRuntime components' pods
	// +optional
	PodMetadata PodMetadata `json:"podMetadata,omitempty"`

	// Resources that will be requested by the CacheRuntime component.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Environment variables that will be used by CacheRuntime component.
	// +optional
	Env map[string]string `json:"env,omitempty"`

	// VolumeMounts specifies the volumes listed in ".spec.volumes" to mount into the CacheRuntime component's filesystem.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`

	// Arguments to the entrypoint.
	// +optional
	// +listType=atomic
	Args []string `json:"args,omitempty"`

	// If specified, the pod will be dispatched by specified scheduler.
	// If not specified, the pod will be dispatched by default scheduler.
	// +optional
	SchedulerName string `json:"schedulerName,omitempty"`

	// NodeSelector is a selector which must be true for the component pods to fit on a node
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// NetworkMode indicates whether to use ContainerNetwork or not
	// +kubebuilder:validation:Enum=HostNetwork;"";ContainerNetwork
	// +kubebuilder:default=ContainerNetwork
	// +optional
	NetworkMode NetworkMode `json:"networkMode,omitempty"`

	// Advanced config of pod's containers
	// +optional
	AdvancedContainerConfigs CacheRuntimeComponentAdvancedContainerConfigs `json:"advancedContainerConfigs,omitempty"`
}

type CacheRuntimeComponentAdvancedContainerConfigs struct {
	AdvancedContainerConfigs map[string]AdvancedContainerConfig `json:"containerConfigs,omitempty"`
}

type AdvancedContainerConfig struct {
	// The version information that instructs fluid to orchestrate a particular version of CacheRuntime Component.
	// +optional
	RuntimeVersion VersionSpec `json:"runtimeVersion,omitempty"`

	// Resources that will be requested by the Cache component. <br>
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Environment variables that will be used by Cache component. <br>
	// +optional
	Env map[string]string `json:"env,omitempty"`

	// VolumeMounts specifies the volumes listed in ".spec.volumes" to mount into the CacheRuntime component's filesystem.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
}

func init() {
	SchemeBuilder.Register(&CacheRuntime{}, &CacheRuntimeList{})
}
