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
}

// JindoRuntimeSpec defines the desired state of JindoRuntime
type JindoRuntimeSpec struct {
	// The version information that instructs fluid to orchestrate a particular version of Jindo.
	JindoVersion VersionSpec `json:"jindoVersion,omitempty"`

	// Desired state for Jindo master
	Master JindoCompTemplateSpec `json:"master,omitempty"`

	// Desired state for Jindo worker
	Worker JindoCompTemplateSpec `json:"worker,omitempty"`

	// Desired state for Jindo Fuse
	Fuse JindoFuseSpec `json:"fuse,omitempty"`

	// Configurable properties for Jindo system. <br>
	Properties map[string]string `json:"properties,omitempty"`

	// Tiered storage used by Jindo
	Tieredstore Tieredstore `json:"tieredstore,omitempty"`

	// The replicas of the worker, need to be specified
	Replicas int32 `json:"replicas,omitempty"`

	// Manage the user to run Jindo Runtime
	RunAs *User `json:"runAs,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
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
// +genclient

// JindoRuntime is the Schema for the jindoruntimes API
type JindoRuntime struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JindoRuntimeSpec `json:"spec,omitempty"`
	Status RuntimeStatus    `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

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
