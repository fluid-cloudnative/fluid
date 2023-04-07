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

const (
	// CacheFS is a fuse filesystem for every posix filesystem with distributed cache
	CacheFSRuntimeKind = "CacheFSRuntime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CacheFSRuntimeSpec defines the desired state of CacheFSRuntime
type CacheFSRuntimeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The version information that instructs fluid to orchestrate a particular version of CacheFS.
	CacheFSVersion VersionSpec `json:"cachefsVersion,omitempty"`

	// The spec of init users
	InitUsers InitUsersSpec `json:"initUsers,omitempty"`

	// The component spec of CacheFS master
	Master CacheFSCompTemplateSpec `json:"master,omitempty"`

	// The component spec of CacheFS worker
	Worker CacheFSCompTemplateSpec `json:"worker,omitempty"`

	// The component spec of CacheFS job Worker
	JobWorker CacheFSCompTemplateSpec `json:"jobWorker,omitempty"`

	// Desired state for CacheFS Fuse
	Fuse CacheFSFuseSpec `json:"fuse,omitempty"`

	// Tiered storage used by CacheFS
	TieredStore TieredStore `json:"tieredstore,omitempty"`

	// Configs of CacheFS
	Configs *[]string `json:"configs,omitempty"`

	// The replicas of the worker, need to be specified
	Replicas int32 `json:"replicas,omitempty"`

	// Manage the user to run CacheFS Runtime
	RunAs *User `json:"runAs,omitempty"`

	// Disable monitoring for CacheFS Runtime
	// Prometheus is enabled by default
	// +optional
	DisablePrometheus bool `json:"disablePrometheus,omitempty"`

	// Volumes is the list of Kubernetes volumes that can be mounted by the alluxio runtime components and/or fuses.
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`

	// PodMetadata defines labels and annotations that will be propagated to CacheFS's pods.
	// +optional
	PodMetadata PodMetadata `json:"podMetadata,omitempty"`

	// CleanCachePolicy defines cleanCache Policy
	// +optional
	CleanCachePolicy CleanCachePolicy `json:"cleanCachePolicy,omitempty"`

	// MetadataList defines labels and annotations that will be propagated to resources created by runtime
	// +optional
	MetadataList []Metadata `json:"metadataList,omitempty"`
}

type CacheFSFuseSpec struct {
	// Image for CacheFS fuse
	Image string `json:"image,omitempty"`

	// Image for CacheFS fuse
	ImageTag string `json:"imageTag,omitempty"`

	// One of the three policies: `Always`, `IfNotPresent`, `Never`
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	// Environment variables that will be used by CacheFS Fuse
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Resources that will be requested by CacheFS Fuse.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// If the fuse client should be deployed in global mode,
	// otherwise the affinity should be considered
	// +optional
	Global bool `json:"global,omitempty"`

	// NodeSelector is a selector which must be true for the fuse client to fit on a node,
	// this option only effect when global is enabled
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// VolumeMounts specifies the volumes listed in ".spec.volumes" to mount into runtime component's filesystem.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`

	// CleanPolicy decides when to clean CacheFS Fuse pods.
	// Currently Fluid supports two policies: OnDemand and OnRuntimeDeleted
	// OnDemand cleans fuse pod once th fuse pod on some node is not needed
	// OnRuntimeDeleted cleans fuse pod only when the cache runtime is deleted
	// Defaults to OnDemand
	// +optional
	CleanPolicy FuseCleanPolicy `json:"cleanPolicy,omitempty"`

	// PodMetadata defines labels and annotations that will be propagated to CacheFS's pods.
	// +optional
	PodMetadata PodMetadata `json:"podMetadata,omitempty"`

	// Whether to use hostnetwork or not
	// +kubebuilder:validation:Enum=HostNetwork;"";ContainerNetwork
	// +optional
	NetworkMode NetworkMode `json:"networkMode,omitempty"`
}

// CacheFSCompTemplateSpec is a description of the CacheFS components
type CacheFSCompTemplateSpec struct {
	// Replicas is the desired number of replicas of the given template.
	// If unspecified, defaults to 1.
	// +kubebuilder:validation:Minimum=1
	// replicas is the min replicas of dataset in the cluster
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Ports used by CacheFS
	// +optional
	Ports []corev1.ContainerPort `json:"ports,omitempty"`

	// Resources that will be requested by the CacheFS component.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Options
	Options map[string]string `json:"options,omitempty"`

	// Environment variables that will be used by CacheFS component.
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

	// PodMetadata defines labels and annotations that will be propagated to CacheFS's pods.
	// +optional
	PodMetadata PodMetadata `json:"podMetadata,omitempty"`

	// Whether to use hostnetwork or not
	// +kubebuilder:validation:Enum=HostNetwork;"";ContainerNetwork
	// +optional
	NetworkMode NetworkMode `json:"networkMode,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.currentWorkerNumberScheduled,selectorpath=.status.selector
// +kubebuilder:printcolumn:name="Ready Workers",type="integer",JSONPath=`.status.workerNumberReady`,priority=10
// +kubebuilder:printcolumn:name="Desired Workers",type="integer",JSONPath=`.status.desiredWorkerNumberScheduled`,priority=10
// +kubebuilder:printcolumn:name="Worker Phase",type="string",JSONPath=`.status.workerPhase`,priority=0
// +kubebuilder:printcolumn:name="Ready Fuses",type="integer",JSONPath=`.status.fuseNumberReady`,priority=10
// +kubebuilder:printcolumn:name="Desired Fuses",type="integer",JSONPath=`.status.desiredFuseNumberScheduled`,priority=10
// +kubebuilder:printcolumn:name="Fuse Phase",type="string",JSONPath=`.status.fusePhase`,priority=0
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`,priority=0
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:resource:categories={fluid},shortName=cachefs
// +genclient

// CacheFSRuntime is the Schema for the cachefsruntimes API
type CacheFSRuntime struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CacheFSRuntimeSpec `json:"spec,omitempty"`
	Status RuntimeStatus      `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CacheFSRuntimeList contains a list of CacheFSRuntime
type CacheFSRuntimeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CacheFSRuntime `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CacheFSRuntime{}, &CacheFSRuntimeList{})
}

// Replicas gets the replicas of runtime worker
func (j *CacheFSRuntime) Replicas() int32 {
	return j.Spec.Replicas
}

func (j *CacheFSRuntime) GetStatus() *RuntimeStatus {
	return &j.Status
}
