/*
Copyright 2020 The Fluid Authors.

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

// CacheRuntimeStatus defines the observed state of CacheRuntime
type CacheRuntimeStatus struct {
	// ValueFileConfigmap used to set helm chart values configurations
	ValueFileConfigmap string `json:"valueFile"`

	// ConfigFileConfigmap used to set cacheruntime configurations
	ConfigFileConfigmap string `json:"configFile"`

	// SetupDuration tell user how much time was spent to setup the runtime
	SetupDuration string `json:"setupDuration,omitempty"`

	// Represents the latest available observations of a ddc runtime's current state.
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []RuntimeCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// CacheStatus represents the total resources of the dataset.
	// TODO: support by cacheManager
	// CacheStates common.CacheStateList `json:"cacheStates,omitempty"`

	// Selector is used for auto-scaling
	Selector string `json:"selector,omitempty"` // this must be the string form of the selector

	// CacheAffinity represents the runtime worker pods node affinity including node selector
	CacheAffinity *corev1.NodeAffinity `json:"cacheAffinity,omitempty"`

	// ComponentsStatus is a map of RuntimeComponent and its status
	ComponentsStatus []RuntimeComponentStatus `json:"componentsStatus,omitempty"`

	// MountPointsStatuses represents the mount points specified in the bounded dataset
	MountPointsStatuses []MountPointStatus `json:"mountPointsStatus,omitempty"`
}

type MountPointStatus struct {
	// MountPoints represents the mount points specified in the bounded dataset
	Mount `json:"mount,omitempty"`

	// MountTime represents time last mount happened
	// if MountTime is earlier than master starting time, remount will be required
	MountTime *metav1.Time `json:"mountTime,omitempty"`
}

type RuntimeComponentStatus struct {
	// Phase is the running phase of component
	Phase RuntimePhase `json:"phase"`

	// Reason for component's condition transition
	Reason string `json:"reason,omitempty"`

	// The total number of nodes that should be running the runtime
	// pod (including nodes correctly running the runtime master pod).
	DesiredScheduledReplicas int32 `json:"desiredScheduledReplicas"`

	// The total number of nodes that should be running the runtime
	// pod (including nodes correctly running the runtime master pod).
	CurrentScheduledReplicas int32 `json:"currentScheduledReplicas"`

	// The number of nodes that should be running the runtime worker pod and have zero
	// or more of the runtime master pod running and ready.
	ReadyReplicas int32 `json:"readyReplicas"`
}
