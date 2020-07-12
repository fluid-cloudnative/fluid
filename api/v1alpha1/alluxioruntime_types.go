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
	"github.com/cloudnativefluid/fluid/pkg/common"
	"k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AlluxioRuntimeRole common.RuntimeRole

const (
	// Master is the type for master of alluxio cluster.
	Master AlluxioRuntimeRole = "master"

	// Worker is the type for workers of alluxio cluster.
	Worker AlluxioRuntimeRole = "worker"

	// Fuse is the type for chief worker of alluxio cluster.
	Fuse AlluxioRuntimeRole = "fuse"
)

// VersionSpec represents the settings for the Alluxio version that fluid is orchestrating.
type AlluxioVersionSpec struct {
	Image           string `json:"image,omitempty"`
	ImageTag        string `json:"imageTag,omitempty"`
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`
}

// AlluxioCompTemplateSpec is a description of the alluxio commponents
type AlluxioCompTemplateSpec struct {
	// Replicas is the desired number of replicas of the given template.
	// If unspecified, defaults to 1.
	// +kubebuilder:validation:Minimum=1
	// replicas is the min replicas of dataset in the cluster
	// +optional
	Replicas   int32    `json:"replicas,omitempty"`
	JvmOptions []string `json:"jvmOptions,omitempty"`
	// +optional
	Properties map[string]string `json:"properties,omitempty"`
	// +optional
	Ports map[string]int `json:"ports,omitempty"`
	// Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources
	// already allocated to the pod.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Placement  rook.PlacementSpec `json:"placement,omitempty"`
	// +optional

}

// AlluxioFuseSpec is a description of the alluxio fuse
type AlluxioFuseSpec struct {
	Image           string            `json:"image,omitempty"`
	ImageTag        string            `json:"imageTag,omitempty"`
	ImagePullPolicy string            `json:"imagePullPolicy,omitempty"`
	JvmOptions      []string          `json:"jvmOptions,omitempty"`
	Properties      map[string]string `json:"properties,omitempty"`
	Env             map[string]string `json:"env,omitempty"`
	// ShortCircuitPolicy string            `json:"shortCircuitPolicy,omitempty"`
	// Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources
	// already allocated to the pod.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// The mountPath in the host machine
	// MountPath string `json:"mountPath,omitempty"`
	// Placement rook.PlacementSpec `json:"placement,omitempty"`
	Args []string `json:"args,omitempty"`
}

type Level struct {
	// Alias string `json:"alias,omitempty"`
	// +kubebuilder:validation:Enum=MEM;SSD;HDD
	// +required
	MediumType common.MediumType `json:"mediumtype"`
	// +kubebuilder:validation:MinLength=1
	// +required
	Path string `json:"path,omitempty"`
	// +required
	Quota *resource.Quantity `json:"quota,omitempty"`
	// StorageType common.CacheStoreType `json:"storageType,omitempty"`
	// float64 is not supported, https://github.com/kubernetes-sigs/controller-tools/issues/245
	High string `json:"high,omitempty"`
	Low  string `json:"low,omitempty"`
}

// Tieredstore is a description of the tiered store
type Tieredstore struct {
	Levels []Level `json:"levels,omitempty"`
}

// AlluxioRuntimeSpec defines the desired state of AlluxioRuntime
type AlluxioRuntimeSpec struct {
	// The version information that instructs fluid to orchestrate a particular version of Alluxio.
	AlluxioVersion AlluxioVersionSpec `json:"alluxioVersion,omitempty"`

	// The placement-related configuration to pass to kubernetes (affinity, node selector, tolerations).
	// Placement rook.PlacementSpec `json:"placement,omitempty"`

	// A spec for alluxio master
	Master AlluxioCompTemplateSpec `json:"master,omitempty"`

	// A spec for alluxio job master
	JobMaster AlluxioCompTemplateSpec `json:"jobMaster,omitempty"`

	// A spec for alluxio worker
	Worker AlluxioCompTemplateSpec `json:"worker,omitempty"`

	// A spec for alluxio job Worker
	JobWorker AlluxioCompTemplateSpec `json:"jobWorker,omitempty"`

	Fuse AlluxioFuseSpec `json:"fuse,omitempty"`

	Properties map[string]string `json:"properties,omitempty"`
	JvmOptions []string          `json:"jvmOptions,omitempty"`

	Tieredstore Tieredstore `json:"tieredstore,omitempty"`

	// The copies of the dataset
	DataReplicas int32 `json:"dataReplicas"`
}

type RuntimePhase string

const (
	RuntimePhaseNone         RuntimePhase = ""
	RuntimePhaseNotReady     RuntimePhase = "NotReady"
	RuntimePhasePartialReady RuntimePhase = "PartialReady"
	RuntimePhaseReady        RuntimePhase = "Ready"
)

// RuntimeConditionType indicates valid conditions type of a runtime
type RuntimeConditionType string

// These are valid conditions of a runtime.
const (
	// MasterInitialized means the master of runtime is initialized
	RuntimeMasterInitialized RuntimeConditionType = "MasterInitialized"
	// MasterReady means the master of runtime is ready
	RuntimeMasterReady RuntimeConditionType = "MasterReady"
	// WorkersInitialized means the Workers of runtime is initialized
	RuntimeWorkersInitialized RuntimeConditionType = "WorkersInitialized"
	// WorkersReady means the Workers of runtime is ready
	RuntimeWorkersReady RuntimeConditionType = "WorkersReady"
	// FusesInitialized means the fuses of runtime is initialized
	RuntimeFusesInitialized RuntimeConditionType = "FusesInitialized"
	// FusesReady means the fuses of runtime is ready
	RuntimeFusesReady RuntimeConditionType = "FusesReady"
)

const (
	RuntimeMasterInitializedReason = "Master is initialized"
	// MasterReady means the master of runtime is ready
	RuntimeMasterReadyReason = "Master is ready"
	// WorkersInitialized means the Workers of runtime is initialized
	RuntimeWorkersInitializedReason = "Workers are initialized"
	// WorkersReady means the Workers of runtime is ready
	RuntimeWorkersReadyReason = "Workers are ready"
	// WorkersInitialized means the Workers of runtime is initialized
	RuntimeFusesInitializedReason = "Fuses are initialized"
	// WorkersReady means the Workers of runtime is ready
	RuntimeFusesReadyReason = "Fuses are ready"
)

// Condition describes the state of the cache at a certain point.
type RuntimeCondition struct {
	// Type of cache condition.
	Type RuntimeConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
	// The last time this condition was updated.
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// AlluxioRuntimeStatus defines the observed state of AlluxioRuntime
type AlluxioRuntimeStatus struct {
	ValueFileConfigmap string `json:"valueFile"`
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// MasterPhase is the master running phase
	MasterPhase  RuntimePhase `json:"masterPhase"`
	MasterReason string       `json:"masterReason,omitempty"`

	// WorkerPhase is the worker running phase
	WorkerPhase  RuntimePhase `json:"workerPhase"`
	WorkerReason string       `json:"workerReason,omitempty"`

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
	FusePhase  RuntimePhase `json:"fusePhase"`
	FuseReason string       `json:"fuseReason,omitempty"`

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
// +kubebuilder:subresource:status

// AlluxioRuntime is the Schema for the alluxioruntimes API
type AlluxioRuntime struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlluxioRuntimeSpec   `json:"spec,omitempty"`
	Status AlluxioRuntimeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AlluxioRuntimeList contains a list of AlluxioRuntime
type AlluxioRuntimeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AlluxioRuntime `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AlluxioRuntime{}, &AlluxioRuntimeList{})
}
