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
	// EAC(Elastic Accelerate Client) is a fuse filesystem for NAS with distributed cache
	EACRuntimeKind = "EACRuntime"
)

// InitFuseSpec is a description of initialize the fuse kernel module for runtime
type InitFuseSpec struct {
	// The version information that instructs fluid to orchestrate a particular version of Alifuse
	Version VersionSpec `json:"version,omitempty"`
}

// OSAdvise is a description of choices to have optimization on specific operating system
type OSAdvise struct {
	// Specific operating system version that can have optimization.
	// +optional
	OSVersion string `json:"osVersion,omitempty"`

	// Enable operating system optimization
	// not enabled by default.
	// +optional
	Enabled bool `json:"enabled,omitempty"`
}

// EACCompTemplateSpec is a description of the EAC components
type EACCompTemplateSpec struct {
	// Replicas is the desired number of replicas of the given template.
	// If unspecified, defaults to 1.
	// +kubebuilder:validation:Minimum=1
	// replicas is the min replicas of dataset in the cluster
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// The version information that instructs fluid to orchestrate a particular version of EAC Comp
	Version VersionSpec `json:"version,omitempty"`

	// Configurable properties for the EAC component.
	// +optional
	Properties map[string]string `json:"properties,omitempty"`

	// Ports used by EAC(e.g. rpc: 19998 for master).
	// +optional
	Ports map[string]int `json:"ports,omitempty"`

	// Resources that will be requested by the EAC component. <br>
	// <br>
	// Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources
	// already allocated to the pod.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Enabled or Disabled for the components.
	// Default enable.
	// +optional
	Disabled bool `json:"disabled,omitempty"`

	// NodeSelector is a selector which must be true for the component to fit on a node.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Whether to use host network or not.
	// +kubebuilder:validation:Enum=HostNetwork;"";ContainerNetwork
	// +optional
	NetworkMode NetworkMode `json:"networkMode,omitempty"`

	// PodMetadata defines labels and annotations that will be propagated to EAC's master and worker pods
	// +optional
	PodMetadata PodMetadata `json:"podMetadata,omitempty"`
}

// EACFuseSpec is a description of the EAC Fuse
type EACFuseSpec struct {
	// The version information that instructs fluid to orchestrate a particular version of EAC Fuse
	Version VersionSpec `json:"version,omitempty"`

	// Configurable properties for EAC fuse
	// +optional
	Properties map[string]string `json:"properties,omitempty"`

	// Resources that will be requested by EAC Fuse. <br>
	// <br>
	// Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources
	// already allocated to the pod.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// NodeSelector is a selector which must be true for the fuse client to fit on a node,
	// this option only effect when global is enabled
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// CleanPolicy decides when to clean EAC Fuse pods.
	// Currently Fluid supports two policies: OnDemand and OnRuntimeDeleted
	// OnDemand cleans fuse pod once th fuse pod on some node is not needed
	// OnRuntimeDeleted cleans fuse pod only when the cache runtime is deleted
	// Defaults to OnRuntimeDeleted
	// +optional
	CleanPolicy FuseCleanPolicy `json:"cleanPolicy,omitempty"`

	// Whether to use hostnetwork or not
	// +kubebuilder:validation:Enum=HostNetwork;"";ContainerNetwork
	// +optional
	NetworkMode NetworkMode `json:"networkMode,omitempty"`

	// PodMetadata defines labels and annotations that will be propagated to EAC's fuse pods
	// +optional
	PodMetadata PodMetadata `json:"podMetadata,omitempty"`
}

// EACRuntimeSpec defines the desired state of EACRuntime
type EACRuntimeSpec struct {
	// The component spec of EAC master
	Master EACCompTemplateSpec `json:"master,omitempty"`

	// The component spec of EAC worker
	Worker EACCompTemplateSpec `json:"worker,omitempty"`

	// The spec of init alifuse
	InitFuse InitFuseSpec `json:"initFuse,omitempty"`

	// The component spec of EAC Fuse
	Fuse EACFuseSpec `json:"fuse,omitempty"`

	// Tiered storage used by EAC worker
	TieredStore TieredStore `json:"tieredstore,omitempty"`

	// The replicas of the worker, need to be specified
	Replicas int32 `json:"replicas,omitempty"`

	// Operating system optimization for EAC
	OSAdvise OSAdvise `json:"osAdvise,omitempty"`

	// CleanCachePolicy defines cleanCache Policy
	// +optional
	CleanCachePolicy CleanCachePolicy `json:"cleanCachePolicy,omitempty"`

	// PodMetadata defines labels and annotations that will be propagated to all EAC's pods
	// +optional
	PodMetadata PodMetadata `json:"podMetadata,omitempty"`

	// PVCMetadata defines labels and annotations that will be propagated to pvc created by Alluxio
	// +optional
	PVCMetadata Metadata `json:"pvcMetadata,omitempty"`
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
// +kubebuilder:resource:categories={fluid},shortName=eac
// +genclient

// EACRuntime is the Schema for the eacruntimes API
type EACRuntime struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EACRuntimeSpec `json:"spec,omitempty"`
	Status RuntimeStatus  `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced

// EACRuntimeList contains a list of EACRuntime
type EACRuntimeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EACRuntime `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EACRuntime{}, &EACRuntimeList{})
}

func (runtime *EACRuntime) Enabled() bool {
	return !runtime.Spec.Worker.Disabled
}

// Replicas gets the replicas of runtime worker
func (runtime *EACRuntime) Replicas() int32 {
	if !runtime.Enabled() {
		return 0
	}
	return runtime.Spec.Replicas
}

func (runtime *EACRuntime) GetStatus() *RuntimeStatus {
	return &runtime.Status
}

func (runtime *EACRuntime) MasterEnabled() bool {
	return !runtime.Spec.Master.Disabled
}

func (runtime *EACRuntime) MasterReplicas() int32 {
	if !runtime.MasterEnabled() {
		return 0
	}
	if runtime.Spec.Master.Replicas < 1 {
		return 1
	}
	return runtime.Spec.Master.Replicas
}
