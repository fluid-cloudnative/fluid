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

import "github.com/fluid-cloudnative/fluid/pkg/common"

// RuntimeStatus defines the observed state of AlluxioRuntime
type RuntimeStatus struct {
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
