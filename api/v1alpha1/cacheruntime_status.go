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
	// SetupValueFile used to set runtime configurations
	SetupValueFile string `json:"setupValueFile"`

	// EngineConfigFile used to set cacheruntime configurations
	EngineConfigFile string `json:"engineConfigFile"` //TODO： 改一下

	// SetupDuration tell user how much time was spent to setup the runtime
	SetupDuration string `json:"setupDuration,omitempty"`

	// Represents the latest available observations of a ddc runtime's current state.
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []RuntimeCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// Selector is used for auto-scaling
	Selector string `json:"selector,omitempty"` // this must be the string form of the selector

	// CacheAffinity represents the runtime worker pods node affinity including node selector
	CacheAffinity *corev1.NodeAffinity `json:"cacheAffinity,omitempty"`

	// ComponentsStatus is a map of RuntimeComponent and its status
	RuntimeComponentStatusCollection `json:",inline"`

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

type RuntimeComponentStatusCollection struct {
	Master RuntimeComponentStatus `json:"master,omitempty"`
	Worker RuntimeComponentStatus `json:"worker,omitempty"`
	Client RuntimeComponentStatus `json:"client,omitempty"`
}

type RuntimeComponentStatus struct {
	// Phase is the running phase of a component
	Phase RuntimePhase `json:"phase"`

	// Reason for component's condition transition
	Reason string `json:"reason,omitempty"`

	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	CurrentReplicas int32 `json:"currentReplicas,omitempty"`

	DesiredReplicas int32 `json:"desiredReplicas,omitempty"`

	UnavailableReplicas int32 `json:"unavailableReplicas,omitempty"`

	AvailableReplicas int32 `json:"availableReplicas,omitempty"`
}
