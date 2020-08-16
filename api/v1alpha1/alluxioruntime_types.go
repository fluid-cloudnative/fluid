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
	// Master is the type for master of Alluxio cluster.
	Master AlluxioRuntimeRole = "master"

	// Worker is the type for workers of Alluxio cluster.
	Worker AlluxioRuntimeRole = "worker"

	// Fuse is the type for chief worker of Alluxio cluster.
	Fuse AlluxioRuntimeRole = "fuse"
)

// VersionSpec represents the settings for the Alluxio version that fluid is orchestrating.
type AlluxioVersionSpec struct {
	// Image for Alluxio(e.g. alluxio/alluxio)
	Image string `json:"image,omitempty"`

	// Image tag for Alluxio(e.g. 2.3.0-SNAPSHOT)
	ImageTag string `json:"imageTag,omitempty"`

	// One of the three policies: `Always`, `IfNotPresent`, `Never`
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`
}

// AlluxioCompTemplateSpec is a description of the Alluxio commponents
type AlluxioCompTemplateSpec struct {
	// Replicas is the desired number of replicas of the given template.
	// If unspecified, defaults to 1.
	// +kubebuilder:validation:Minimum=1
	// replicas is the min replicas of dataset in the cluster
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Options for JVM
	JvmOptions []string `json:"jvmOptions,omitempty"`

	// Configurable properties for the Alluxio component. <br>
	// Refer to <a href="https://docs.alluxio.io/os/user/stable/en/reference/Properties-List.html">Alluxio Configuration Properties</a> for more info
	// +optional
	Properties map[string]string `json:"properties,omitempty"`

	// Ports used by Alluxio(e.g. rpc: 19998 for master)
	// +optional
	Ports map[string]int `json:"ports,omitempty"`

	// Resources that will be requested by the Alluxio component. <br>
	// <br>
	// Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources
	// already allocated to the pod.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Environment variables that will be used by Alluxio component. <br>
	Env map[string]string `json:"env,omitempty"`
}

// AlluxioFuseSpec is a description of the Alluxio Fuse
type AlluxioFuseSpec struct {

	// Image for Alluxio Fuse(e.g. alluxio/alluxio-fuse)
	Image string `json:"image,omitempty"`

	// Image Tag for Alluxio Fuse(e.g. 2.3.0-SNAPSHOT)
	ImageTag string `json:"imageTag,omitempty"`

	// One of the three policies: `Always`, `IfNotPresent`, `Never`
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	// Options for JVM
	JvmOptions []string `json:"jvmOptions,omitempty"`

	// Configurable properties for Alluxio System. <br>
	// Refer to <a href="https://docs.alluxio.io/os/user/stable/en/reference/Properties-List.html">Alluxio Configuration Properties</a> for more info
	Properties map[string]string `json:"properties,omitempty"`

	// Environment variables that will be used by Alluxio Fuse
	Env map[string]string `json:"env,omitempty"`

	// ShortCircuitPolicy string            `json:"shortCircuitPolicy,omitempty"`

	// Resources that will be requested by Alluxio Fuse. <br>
	// <br>
	// Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources
	// already allocated to the pod.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// The mountPath in the host machine
	// MountPath string `json:"mountPath,omitempty"`
	// Placement rook.PlacementSpec `json:"placement,omitempty"`

	// Arguments that will be passed to Alluxio Fuse
	Args []string `json:"args,omitempty"`
}

// Level describes configurations a tier needs. <br>
// Refer to <a href="https://docs.alluxio.io/os/user/stable/en/core-services/Caching.html#configuring-tiered-storage">Configuring Tiered Storage</a> for more info
type Level struct {
	// Alias string `json:"alias,omitempty"`

	// Medium Type of the tier. One of the three types: `MEM`, `SSD`, `HDD`
	// +kubebuilder:validation:Enum=MEM;SSD;HDD
	// +required
	MediumType common.MediumType `json:"mediumtype"`

	// File path to be used for the tier (e.g. /mnt/ramdisk)
	// +kubebuilder:validation:MinLength=1
	// +required
	Path string `json:"path,omitempty"`

	// Quota for the tier. (e.g. 100GB)
	// +required
	Quota *resource.Quantity `json:"quota,omitempty"`

	// StorageType common.CacheStoreType `json:"storageType,omitempty"`
	// float64 is not supported, https://github.com/kubernetes-sigs/controller-tools/issues/245

	// Ratio of high watermark of the tier (e.g. 0.9)
	High string `json:"high,omitempty"`

	// Ratio of low watermark of the tier (e.g. 0.7)
	Low string `json:"low,omitempty"`
}

// Tieredstore is a description of the tiered store
type Tieredstore struct {
	// configurations for multiple tiers
	Levels []Level `json:"levels,omitempty"`
}

// Data management strategies
type Data struct {
	// The copies of the dataset
	// +optional
	Replicas int32 `json:"replicas"`

	// Pin the dataset or not. Refer to <a href="https://docs.alluxio.io/os/user/stable/en/operation/User-CLI.html#pin">Alluxio User-CLI pin</a>
	// +optional
	Pin bool `json:"pin"`
}

// AlluxioRuntimeSpec defines the desired state of AlluxioRuntime
type AlluxioRuntimeSpec struct {
	// The version information that instructs fluid to orchestrate a particular version of Alluxio.
	AlluxioVersion AlluxioVersionSpec `json:"alluxioVersion,omitempty"`

	// The placement-related configuration to pass to kubernetes (affinity, node selector, tolerations).
	// Placement rook.PlacementSpec `json:"placement,omitempty"`

	// Desired state for Alluxio master
	Master AlluxioCompTemplateSpec `json:"master,omitempty"`

	// Desired state for Alluxio job master
	JobMaster AlluxioCompTemplateSpec `json:"jobMaster,omitempty"`

	// Desired state for Alluxio worker
	Worker AlluxioCompTemplateSpec `json:"worker,omitempty"`

	// Desired state for Alluxio job Worker
	JobWorker AlluxioCompTemplateSpec `json:"jobWorker,omitempty"`

	// Desired state for Alluxio Fuse
	Fuse AlluxioFuseSpec `json:"fuse,omitempty"`

	// Configurable properties for Alluxio system. <br>
	// Refer to <a href="https://docs.alluxio.io/os/user/stable/en/reference/Properties-List.html">Alluxio Configuration Properties</a> for more info
	Properties map[string]string `json:"properties,omitempty"`

	// Options for JVM
	JvmOptions []string `json:"jvmOptions,omitempty"`

	// Tiered storage used by Alluxio
	Tieredstore Tieredstore `json:"tieredstore,omitempty"`

	//Management strategies for the dataset to which the runtime is bound
	Data Data `json:"data,omitempty"`

	// The replicas of the worker, need to be specified
	Replicas int32 `json:"replicas,omitempty"`
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
	// config map used to set configurations
	ValueFileConfigmap string `json:"valueFile"`

	// MasterPhase is the master running phase
	MasterPhase RuntimePhase `json:"masterPhase"`

	// Reason for Alluxio Master's condition transition
	MasterReason string `json:"masterReason,omitempty"`

	// WorkerPhase is the worker running phase
	WorkerPhase RuntimePhase `json:"workerPhase"`

	// Reason for Alluxio Worker's condition transition
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
// +kubebuilder:subresource:status
// +genclient

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

// Replicas gets the replicas of runtime worker
func (runtime *AlluxioRuntime) Replicas() int32 {
	return runtime.Spec.Replicas
}
