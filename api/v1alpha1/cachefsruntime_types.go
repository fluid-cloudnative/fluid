/*
Copyright 2024 The Fluid Authors.

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
	CacheFSRuntimeKind = "CacheFSRuntime"
)

// CacheFSRuntimeSpec defines the desired state of CacheFSRuntime
type CacheFSRuntimeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The version information that instructs fluid to orchestrate a particular version of CacheFS.
	CacheFSVersion VersionSpec `json:"cachefsVersion,omitempty"`

	// The spec of CacheFS master, used as CacheFS meta engine like redis.
	// +optional
	Master CacheFSMasterSpec `json:"master,omitempty"`

	// The component spec of CacheFS worker
	Worker CacheFSCompSpec `json:"worker,omitempty"`

	// Desired state for CacheFS Fuse
	Fuse CacheFSFuseSpec `json:"fuse,omitempty"`

	// Tiered storage used by CacheFS
	TieredStore TieredStore `json:"tieredstore,omitempty"`

	// Configs of CacheFS, options for cachefs format sub-command
	Configs map[string]string `json:"configs,omitempty"`

	// The replicas of the worker, need to be specified
	Replicas int32 `json:"replicas,omitempty"`

	// Manage the user to run Cachefs Runtime
	RunAs *User `json:"runAs,omitempty"`

	// Volumes is the list of Kubernetes volumes that can be mounted by the cachefs runtime components and/or fuses.
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`

	// PodMetadata defines labels and annotations that will be propagated to CacheFs's pods.
	// +optional
	PodMetadata PodMetadata `json:"podMetadata,omitempty"`
}

// CacheFSMasterSpec defines the configurations for CacheFS Master, which is a redis server.
type CacheFSMasterSpec struct {
	// Image for CacheFS master
	Image string `json:"image,omitempty"`

	// Image for CacheFS master
	ImageTag string `json:"imageTag,omitempty"`

	// One of the three policies: `Always`, `IfNotPresent`, `Never`
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	// Environment variables that will be used by CacheFS master
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Resources that will be requested by CacheFS master.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Options master instance will use, all valid redis options except `protected-mode` will be
	// accepted. e.g.
	// master:
	//   options:
	//     maxmemory: "1gb"
	//     port: 6379
	//     loglevel: notice
	// +optional
	Options map[string]string `json:"options,omitempty"`

	// NodeSelector is a selector which must be true for the master to fit on a node,
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// VolumeMounts specifies the volumes listed in ".spec.volumes" to mount into master's filesystem.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`

	// PodMetadata defines labels and annotations that will be propagated to CacheFS's master.
	// +optional
	PodMetadata PodMetadata `json:"podMetadata,omitempty"`

	// Whether to use hostnetwork or not
	// +kubebuilder:validation:Enum=HostNetwork;"";ContainerNetwork
	// +optional
	NetworkMode NetworkMode `json:"networkMode,omitempty"`

	// ExternalEndpoint defines the configurations for external master instance, e.g. external
	// redis hosted on Aliyun.
	// Default is not set.
	// If set, the CacheFS Master component will not be deployed, and the external master
	// instance will be used.
	// E,g.
	//   endpoint:
	//     metaurl: "redis://instance-name.redis.rds.aliyuncs.com:6379/1"
	// Or
	//   endpoint:
	//     encryptOptions:
	//     - name: metaurl
	//       valueFrom:
	//         secretKeyRef:
	//           name: cfs-secret
	//           key: metaurl
	//
	// +optional
	ExternalEndpoint EndpointSpec `json:"endpoint,omitempty"`
}

// EndpointSpec defines the configurations for external etcd cluster
type EndpointSpec struct {
	// MetaUrl specifies the endpoint of external master instance
	// e,g. "redis://host:6379/1"
	// +optional
	MetaUrl string `json:"metaUrl,omitempty"`

	// encrypted info for accessing the external master instance, only accept "metaurl" option
	// name and this will override `MetaUrl` above, if both are specified.
	// +optional
	EncryptOptions []EncryptOption `json:"encryptOptions,omitempty"`
}

// CacheFSCompSpec is a description of the CacheFS components
type CacheFSCompSpec struct {
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

	// Options mount options that fuse pod will use
	// +optional
	Options map[string]string `json:"options,omitempty"`

	// NodeSelector is a selector which must be true for the fuse client to fit on a node,
	// this option only effect when global is enabled
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// VolumeMounts specifies the volumes listed in ".spec.volumes" to mount into runtime component's filesystem.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`

	// CleanPolicy decides when to clean Cachefs Fuse pods.
	// Currently Fluid supports two policies: OnDemand and OnRuntimeDeleted
	// OnDemand cleans fuse pod once the fuse pod on some node is not needed
	// OnRuntimeDeleted cleans fuse pod only when the cache runtime is deleted
	// Defaults to OnRuntimeDeleted
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
// +kubebuilder:resource:scope=Namespaced

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
func (c *CacheFSRuntime) Replicas() int32 {
	return c.Spec.Replicas
}

func (c *CacheFSRuntime) GetStatus() *RuntimeStatus {
	return &c.Status
}
