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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VersionSpec represents the settings for the Jindo version that fluid is orchestrating.
type JindoVersionSpec struct {
	// Image for Jindo(e.g. jindo/jindo)
	Image string `json:"image,omitempty"`

	// Image tag for Jindo(e.g. 2.3.0-SNAPSHOT)
	ImageTag string `json:"imageTag,omitempty"`

	// One of the three policies: `Always`, `IfNotPresent`, `Never`
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`
}

// JindoCompTemplateSpec is a description of the Jindo commponents
type JindoCompTemplateSpec struct {
	// Replicas is the desired number of replicas of the given template.
	// If unspecified, defaults to 1.
	// +kubebuilder:validation:Minimum=1
	// replicas is the min replicas of dataset in the cluster
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Configurable properties for the Jindo component. <br>
	// Refer to <a href="https://docs.jindo.io/os/user/stable/en/reference/Properties-List.html">Jindo Configuration Properties</a> for more info
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
	// Refer to <a href="https://docs.jindo.io/os/user/stable/en/reference/Properties-List.html">Jindo Configuration Properties</a> for more info
	Properties map[string]int `json:"properties,omitempty"`

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
	JindoVersion JindoVersionSpec `json:"jindoVersion,omitempty"`

	// Desired state for Jindo master
	Master JindoCompTemplateSpec `json:"master,omitempty"`

	// Desired state for Jindo job master
	JobMaster JindoCompTemplateSpec `json:"jobMaster,omitempty"`

	// Desired state for Jindo worker
	Worker JindoCompTemplateSpec `json:"worker,omitempty"`

	// Desired state for Jindo job Worker
	JobWorker JindoCompTemplateSpec `json:"jobWorker,omitempty"`

	// The spec of init users
	InitUsers InitUsersSpec `json:"initUsers,omitempty"`

	// Desired state for Jindo Fuse
	Fuse JindoFuseSpec `json:"fuse,omitempty"`

	// Configurable properties for Jindo system. <br>
	// Refer to <a href="https://docs.jindo.io/os/user/stable/en/reference/Properties-List.html">Jindo Configuration Properties</a> for more info
	Properties map[string]string `json:"properties,omitempty"`

	// Tiered storage used by Jindo
	Tieredstore Tieredstore `json:"tieredstore,omitempty"`

	// The replicas of the worker, need to be specified
	Replicas int32 `json:"replicas,omitempty"`

	// Manage the user to run Jindo Runtime
	RunAs *User `json:"runAs,omitempty"`
}

// JindoRuntimeStatus defines the observed state of JindoRuntime
type JindoRuntimeStatus struct {
	// config map used to set configurations
	ValueFileConfigmap string `json:"valueFile"`

	// MasterPhase is the master running phase
	MasterPhase RuntimePhase `json:"masterPhase"`

	// Reason for Jindo Master's condition transition
	MasterReason string `json:"masterReason,omitempty"`

	// WorkerPhase is the worker running phase
	WorkerPhase RuntimePhase `json:"workerPhase"`

	// Reason for Jindo Worker's condition transition
	WorkerReason string `json:"workerReason,omitempty"`

	// The total number of nodes that should be running the runtime worker
	// pod (including nodes correctly running the runtime worker pod).
	DesiredWorkerNumberScheduled int32 `json:"desiredWorkerNumberScheduled"`

	// The total number of nodes that can be running the runtime worker
	// pod (including nodes correctly running the runtime worker pod).
	CurrentWorkerNumberScheduled int32 `json:"currentWorkerNumberScheduled"`

	// The number of nodes that should be running the runtime worker pod and have one
	// or more of the runtime worker pod running and ready.
	WorkerNumberReady int32 `json:"workerNumberReady"`

	// The number of nodes that should be running the
	// runtime worker pod and have one or more of the runtime worker pod running and
	// available (ready for at least spec.minReadySeconds)
	// +optional
	WorkerNumberAvailable int32 `json:"workerNumberAvailable,omitempty"`

	// The number of nodes that should be running the
	// runtime worker pod and have none of the runtime worker pod running and available
	// (ready for at least spec.minReadySeconds)
	// +optional
	WorkerNumberUnavailable int32 `json:"workerNumberUnavailable,omitempty"`

	// The total number of nodes that should be running the runtime
	// pod (including nodes correctly running the runtime master pod).
	DesiredMasterNumberScheduled int32 `json:"desiredMasterNumberScheduled"`

	// The total number of nodes that should be running the runtime
	// pod (including nodes correctly running the runtime master pod).
	CurrentMasterNumberScheduled int32 `json:"currentMasterNumberScheduled"`

	// The number of nodes that should be running the runtime worker pod and have zero
	// or more of the runtime master pod running and ready.
	MasterNumberReady int32 `json:"masterNumberReady"`

	// FusePhase is the Fuse running phase
	FusePhase RuntimePhase `json:"fusePhase"`

	// Reason for the condition's last transition.
	FuseReason string `json:"fuseReason,omitempty"`

	// The total number of nodes that can be running the runtime Fuse
	// pod (including nodes correctly running the runtime Fuse pod).
	CurrentFuseNumberScheduled int32 `json:"currentFuseNumberScheduled"`

	// The total number of nodes that should be running the runtime Fuse
	// pod (including nodes correctly running the runtime Fuse pod).
	DesiredFuseNumberScheduled int32 `json:"desiredFuseNumberScheduled"`

	// The number of nodes that should be running the runtime Fuse pod and have one
	// or more of the runtime Fuse pod running and ready.
	FuseNumberReady int32 `json:"fuseNumberReady"`

	// The number of nodes that should be running the
	// runtime fuse pod and have none of the runtime fuse pod running and available
	// (ready for at least spec.minReadySeconds)
	// +optional
	FuseNumberUnavailable int32 `json:"fuseNumberUnavailable,omitempty"`

	// The number of nodes that should be running the
	// runtime Fuse pod and have one or more of the runtime Fuse pod running and
	// available (ready for at least spec.minReadySeconds)
	// +optional
	FuseNumberAvailable int32 `json:"fuseNumberAvailable,omitempty"`

	// Represents the latest available observations of a ddc runtime's current state.
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []RuntimeCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// CacheStatus represents the total resources of the dataset.
	CacheStates common.CacheStateList `json:"cacheStates,omitempty"`
}

// +kubebuilder:object:root=true

// JindoRuntime is the Schema for the jindoruntimes API
type JindoRuntime struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JindoRuntimeSpec   `json:"spec,omitempty"`
	Status JindoRuntimeStatus `json:"status,omitempty"`
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
