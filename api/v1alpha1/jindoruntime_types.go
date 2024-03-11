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
	JindoRuntimeKind = "JindoRuntime"
)

// JindoCompTemplateSpec is a description of the Jindo commponents
type JindoCompTemplateSpec struct {
	// Replicas is the desired number of replicas of the given template.
	// If unspecified, defaults to 1.
	// +kubebuilder:validation:Minimum=1
	// replicas is the min replicas of dataset in the cluster
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Configurable properties for the Jindo component. <br>
	// +optional
	Properties map[string]string `json:"properties,omitempty"`

	// +optional
	Ports map[string]int `json:"ports,omitempty"`

	// Resources that will be requested by the Jindo component. <br>
	// <br>
	// Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources
	// already allocated to the pod.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Environment variables that will be used by Jindo component. <br>
	Env map[string]string `json:"env,omitempty"`

	// NodeSelector is a selector which must be true for the master to fit on a node
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Labels will be added on JindoFS Master or Worker pods.
	// DEPRECATED: This is a deprecated field. Please use PodMetadata instead.
	// Note: this field is set to be exclusive with PodMetadata.Labels
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// PodMetadata defines labels and annotations that will be propagated to Jindo's pods
	// +optional
	PodMetadata PodMetadata `json:"podMetadata,omitempty"`

	// If disable JindoFS master or worker
	// +optional
	Disabled bool `json:"disabled,omitempty"`

	// VolumeMounts specifies the volumes listed in ".spec.volumes" to mount into the jindo runtime component's filesystem.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
}

// JindoFuseSpec is a description of the Jindo Fuse
type JindoFuseSpec struct {

	// Image for Jindo Fuse(e.g. jindo/jindo-fuse)
	Image string `json:"image,omitempty"`

	// Image Tag for Jindo Fuse(e.g. 2.3.0-SNAPSHOT)
	ImageTag string `json:"imageTag,omitempty"`

	// One of the three policies: `Always`, `IfNotPresent`, `Never`
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	// Configurable properties for Jindo System. <br>
	Properties map[string]string `json:"properties,omitempty"`

	// Environment variables that will be used by Jindo Fuse
	Env map[string]string `json:"env,omitempty"`

	// ShortCircuitPolicy string            `json:"shortCircuitPolicy,omitempty"`

	// Resources that will be requested by Jindo Fuse. <br>
	// <br>
	// Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources
	// already allocated to the pod.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Arguments that will be passed to Jindo Fuse
	Args []string `json:"args,omitempty"`

	// If the fuse client should be deployed in global mode,
	// otherwise the affinity should be considered
	// +optional
	Global bool `json:"global,omitempty"`

	// NodeSelector is a selector which must be true for the fuse client to fit on a node,
	// this option only effect when global is enabled
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Labels will be added on all the JindoFS pods.
	// DEPRECATED: this is a deprecated field. Please use PodMetadata.Labels instead.
	// Note: this field is set to be exclusive with PodMetadata.Labels
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// PodMetadata defines labels and annotations that will be propagated to Jindo's fuse pods
	// +optional
	PodMetadata PodMetadata `json:"podMetadata,omitempty"`

	// CleanPolicy decides when to clean JindoFS Fuse pods.
	// Currently Fluid supports two policies: OnDemand and OnRuntimeDeleted
	// OnDemand cleans fuse pod once th fuse pod on some node is not needed
	// OnRuntimeDeleted cleans fuse pod only when the cache runtime is deleted
	// Defaults to OnRuntimeDeleted
	// +optional
	CleanPolicy FuseCleanPolicy `json:"cleanPolicy,omitempty"`

	// If disable JindoFS fuse
	// +optional
	Disabled bool `json:"disabled,omitempty"`

	// +optional
	LogConfig map[string]string `json:"logConfig,omitempty"`

	// Use the host's pid namespace, default false.
	// +optional
	HostPID bool `json:"hostPID,omitempty"`
}

// JindoRuntimeSpec defines the desired state of JindoRuntime
type JindoRuntimeSpec struct {
	// The version information that instructs fluid to orchestrate a particular version of Jindo.
	JindoVersion VersionSpec `json:"jindoVersion,omitempty"`

	// The component spec of Jindo master
	Master JindoCompTemplateSpec `json:"master,omitempty"`

	// The component spec of Jindo worker
	Worker JindoCompTemplateSpec `json:"worker,omitempty"`

	// The component spec of Jindo Fuse
	Fuse JindoFuseSpec `json:"fuse,omitempty"`

	// Configurable properties for Jindo system. <br>
	Properties map[string]string `json:"properties,omitempty"`

	// Tiered storage used by Jindo
	TieredStore TieredStore `json:"tieredstore,omitempty"`

	// The replicas of the worker, need to be specified
	Replicas int32 `json:"replicas,omitempty"`

	// Manage the user to run Jindo Runtime
	RunAs *User `json:"runAs,omitempty"`

	User string `json:"user,omitempty"`

	// Name of the configMap used to support HDFS configurations when using HDFS as Jindo's UFS. The configMap
	// must be in the same namespace with the JindoRuntime. The configMap should contain user-specific HDFS conf files in it.
	// For now, only "hdfs-site.xml" and "core-site.xml" are supported. It must take the filename of the conf file as the key and content
	// of the file as the value.
	// +optional
	HadoopConfig string `json:"hadoopConfig,omitempty"`

	Secret string `json:"secret,omitempty"`

	// Labels will be added on all the JindoFS pods.
	// DEPRECATED: this is a deprecated field. Please use PodMetadata.Labels instead.
	// Note: this field is set to be exclusive with PodMetadata.Labels
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// PodMetadata defines labels and annotations that will be propagated to all Jindo's fuse pods
	// +optional
	PodMetadata PodMetadata `json:"podMetadata,omitempty"`

	// +optional
	LogConfig map[string]string `json:"logConfig,omitempty"`

	// Whether to use hostnetwork or not
	// +kubebuilder:validation:Enum=HostNetwork;"";ContainerNetwork
	// +optional
	NetworkMode NetworkMode `json:"networkmode,omitempty"`

	// CleanCachePolicy defines cleanCache Policy
	// +optional
	CleanCachePolicy CleanCachePolicy `json:"cleanCachePolicy,omitempty"`

	// Volumes is the list of Kubernetes volumes that can be mounted by the jindo runtime components and/or fuses.
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.currentWorkerNumberScheduled
// +kubebuilder:printcolumn:name="Ready Masters",type="integer",JSONPath=`.status.masterNumberReady`,priority=10
// +kubebuilder:printcolumn:name="Desired Masters",type="integer",JSONPath=`.status.desiredMasterNumberScheduled`,priority=10
// +kubebuilder:printcolumn:name="Master Phase",type="string",JSONPath=`.status.masterPhase`,priority=0
// +kubebuilder:printcolumn:name="Ready Workers",type="integer",JSONPath=`.status.workerNumberReady`,priority=10
// +kubebuilder:printcolumn:name="Desired Workers",type="integer",JSONPath=`.status.desiredWorkerNumberScheduled`,priority=10
// +kubebuilder:printcolumn:name="Worker Phase",type="string",JSONPath=`.status.workerPhase`,priority=0
// +kubebuilder:printcolumn:name="Ready Fuses",type="integer",JSONPath=`.status.fuseNumberReady`,priority=10
// +kubebuilder:printcolumn:name="Desired Fuses",type="integer",JSONPath=`.status.desiredFuseNumberScheduled`,priority=10
// +kubebuilder:printcolumn:name="Fuse Phase",type="string",JSONPath=`.status.fusePhase`,priority=0
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`,priority=0
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:resource:categories={fluid},shortName=jindo
// +genclient

// JindoRuntime is the Schema for the jindoruntimes API
type JindoRuntime struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JindoRuntimeSpec `json:"spec,omitempty"`
	Status RuntimeStatus    `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced

// JindoRuntimeList contains a list of JindoRuntime
type JindoRuntimeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JindoRuntime `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JindoRuntime{}, &JindoRuntimeList{})
}

// Replicas gets the replicas of runtime worker
func (runtime *JindoRuntime) Replicas() int32 {
	return runtime.Spec.Replicas
}

func (runtime *JindoRuntime) GetStatus() *RuntimeStatus {
	return &runtime.Status
}
