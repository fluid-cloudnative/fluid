/*
Copyright 2021 The Fluid Authors.

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

// JuiceFSRuntimeSpec defines the desired state of JuiceFSRuntime
type JuiceFSRuntimeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The version information that instructs fluid to orchestrate a particular version of JuiceFS.
	JuiceFSVersion VersionSpec `json:"juicefsVersion,omitempty"`

	// The spec of init users
	InitUsers InitUsersSpec `json:"initUsers,omitempty"`

	// The component spec of JuiceFS master
	Master JuiceFSCompTemplateSpec `json:"master,omitempty"`

	// The component spec of JuiceFS worker
	Worker JuiceFSCompTemplateSpec `json:"worker,omitempty"`

	// The component spec of JuiceFS job Worker
	JobWorker JuiceFSCompTemplateSpec `json:"jobWorker,omitempty"`

	// Desired state for JuiceFS Fuse
	Fuse JuiceFSFuseSpec `json:"fuse,omitempty"`

	// Tiered storage used by JuiceFS
	TieredStore TieredStore `json:"tieredstore,omitempty"`

	// The replicas of the worker, need to be specified
	Replicas int32 `json:"replicas,omitempty"`

	// Manage the user to run Juicefs Runtime
	RunAs *User `json:"runAs,omitempty"`

	// Disable monitoring for JuiceFS Runtime
	// Prometheus is enabled by default
	// +optional
	DisablePrometheus bool `json:"disablePrometheus,omitempty"`
}

// JuiceFSCompTemplateSpec is a description of the JuiceFS components
type JuiceFSCompTemplateSpec struct {
	// Replicas is the desired number of replicas of the given template.
	// If unspecified, defaults to 1.
	// +kubebuilder:validation:Minimum=1
	// replicas is the min replicas of dataset in the cluster
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Ports used by JuiceFS
	// +optional
	Ports []corev1.ContainerPort `json:"ports,omitempty"`

	// Resources that will be requested by the JuiceFS component.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Environment variables that will be used by JuiceFS component.
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Enabled or Disabled for the components.
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// NodeSelector is a selector
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

type JuiceFSFuseSpec struct {
	// Image for JuiceFS fuse
	Image string `json:"image,omitempty"`

	// Image for JuiceFS fuse
	ImageTag string `json:"image_tag,omitempty"`

	// One of the three policies: `Always`, `IfNotPresent`, `Never`
	ImagePullPolicy string `json:"image_pull_policy,omitempty"`

	// Environment variables that will be used by JuiceFS Fuse
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Resources that will be requested by JuiceFS Fuse.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// If the fuse client should be deployed in global mode,
	// otherwise the affinity should be considered
	// +optional
	Global bool `json:"global,omitempty"`

	// NodeSelector is a selector which must be true for the fuse client to fit on a node,
	// this option only effect when global is enabled
	// +optional
	NodeSelector map[string]string `json:"node_selector,omitempty"`
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
// +genclient

// JuiceFSRuntime is the Schema for the juicefsruntimes API
type JuiceFSRuntime struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JuiceFSRuntimeSpec `json:"spec,omitempty"`
	Status RuntimeStatus      `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// JuiceFSRuntimeList contains a list of JuiceFSRuntime
type JuiceFSRuntimeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JuiceFSRuntime `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JuiceFSRuntime{}, &JuiceFSRuntimeList{})
}

// Replicas gets the replicas of runtime worker
func (r *JuiceFSRuntime) Replicas() int32 {
	return r.Spec.Replicas
}
