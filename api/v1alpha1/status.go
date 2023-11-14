/*
Copyright 2022 The Fluid Authors.

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

// RuntimeStatus defines the observed state of Runtime
type RuntimeStatus struct {
	// config map used to set configurations
	ValueFileConfigmap string `json:"valueFile"`

	// MasterPhase is the master running phase
	MasterPhase RuntimePhase `json:"masterPhase"`

	// Reason for Master's condition transition
	MasterReason string `json:"masterReason,omitempty"`

	// WorkerPhase is the worker running phase
	WorkerPhase RuntimePhase `json:"workerPhase"`

	// Reason for Worker's condition transition
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

	// Duration tell user how much time was spent to setup the runtime
	SetupDuration string `json:"setupDuration,omitempty"`

	// Represents the latest available observations of a ddc runtime's current state.
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []RuntimeCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// CacheStatus represents the total resources of the dataset.
	CacheStates common.CacheStateList `json:"cacheStates,omitempty"`

	// Selector is used for auto-scaling
	Selector string `json:"selector,omitempty"` // this must be the string form of the selector

	// APIGatewayStatus represents rest api gateway status
	APIGatewayStatus *APIGatewayStatus `json:"apiGateway,omitempty"`

	// MountTime represents time last mount happened
	// if Mounttime is earlier than master starting time, remount will be required
	MountTime *metav1.Time `json:"mountTime,omitempty"`

	// MountPoints represents the mount points specified in the bounded dataset
	Mounts []Mount `json:"mounts,omitempty"`

	// CacheAffinity represents the runtime worker pods node affinity including node selector
	CacheAffinity *corev1.NodeAffinity `json:"cacheAffinity,omitempty"`
}

// OperationStatus defines the observed state of operation
type OperationStatus struct {
	// Phase describes current phase of operation
	Phase common.Phase `json:"phase"`
	// Duration tell user how much time was spent to operation
	Duration string `json:"duration"`
	// Conditions consists of transition information on operation's Phase
	Conditions []Condition `json:"conditions"`

	// Infos operation customized name-value
	Infos map[string]string `json:"infos,omitempty"`

	// LastScheduleTime is the last time the cron operation was scheduled
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`
	// LastSuccessfulTime is the last time the cron operation successfully completed
	LastSuccessfulTime *metav1.Time `json:"lastSuccessfulTime,omitempty"`
	// WaitingStatus stores information about waiting operation.
	WaitingFor WaitingStatus `json:"waitingFor,omitempty"`
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
	// RuntimeMasterInitialized means the master of runtime is initialized
	RuntimeMasterInitialized RuntimeConditionType = "MasterInitialized"
	// RuntimeMasterReady means the master of runtime is ready
	RuntimeMasterReady RuntimeConditionType = "MasterReady"
	// RuntimeWorkersInitialized means the workers of runtime are initialized
	RuntimeWorkersInitialized RuntimeConditionType = "WorkersInitialized"
	// RuntimeWorkersReady means the workers of runtime are ready
	RuntimeWorkersReady RuntimeConditionType = "WorkersReady"
	// RuntimeWorkerScaledIn means the workers of runtime just scaled in
	RuntimeWorkerScaledIn RuntimeConditionType = "WorkersScaledIn"
	// RuntimeWorkerScaledIn means the workers of runtime just scaled out
	RuntimeWorkerScaledOut RuntimeConditionType = "WorkersScaledOut"
	// RuntimeFusesInitialized means the fuses of runtime are initialized
	RuntimeFusesInitialized RuntimeConditionType = "FusesInitialized"
	// RuntimeFusesReady means the fuses of runtime are ready
	RuntimeFusesReady RuntimeConditionType = "FusesReady"
	// RuntimeFusesScaledIn means the fuses of runtime just scaled in
	RuntimeFusesScaledIn RuntimeConditionType = "FusesScaledIn"
	// RuntimeFusesScaledOut means the fuses of runtime just scaled out
	RuntimeFusesScaledOut RuntimeConditionType = "FusesScaledOut"
)

const (
	// RuntimeMasterInitializedReason means the master of runtime is initialized
	RuntimeMasterInitializedReason = "Master is initialized"
	// RuntimeMasterReadyReason means the master of runtime is ready
	RuntimeMasterReadyReason = "Master is ready"
	// RuntimeWorkersInitializedReason means the workers of runtime are initialized
	RuntimeWorkersInitializedReason = "Workers are initialized"
	// RuntimeWorkersReadyReason means the workers of runtime are ready
	RuntimeWorkersReadyReason = "Workers are ready"
	// RuntimeWorkersScaledInReason means the workers of runtime just scaled in
	RuntimeWorkersScaledInReason = "Workers scaled in"
	// RuntimeWorkersScaledInReason means the workers of runtime just scaled out
	RuntimeWorkersScaledOutReason = "Workers scaled out"
	// RuntimeFusesInitializedReason means the fuses of runtime are initialized
	RuntimeFusesInitializedReason = "Fuses are initialized"
	// RuntimeFusesReadyReason means the fuses of runtime are ready
	RuntimeFusesReadyReason = "Fuses are ready"
	// RuntimeFusesScaledInReason means the fuses of runtime just scaled in
	RuntimeFusesScaledInReason = "Fuses scaled in"
	// RuntimeFusesScaledInReason means the fuses of runtime just scaled out
	RuntimeFusesScaledOutReason = "Fuses scaled out"
)

// Condition describes the state of the cache at a certain point.
type RuntimeCondition struct {
	// Type of cache condition.
	Type RuntimeConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
	// The last time this condition was updated.
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}
