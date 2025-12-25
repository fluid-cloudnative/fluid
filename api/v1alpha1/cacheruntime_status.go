/*
Copyright 2025 The Fluid Authors.

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

// GenericCacheRuntimeStatus defines the observed state of GenericCacheRuntime
type GenericCacheRuntimeStatus struct {
	// SetupValueFile used to set runtime configurations
	SetupValueFile string `json:"setupValueFile,omitempty"`

	// EngineConfigFile used to set cacheruntime configurations
	EngineConfigFile string `json:"engineConfigFile,omitempty"`

	// SetupDuration tells user how much time was spent to setup the runtime
	SetupDuration string `json:"setupDuration,omitempty"`

	// Represents the latest available observations of a ddc runtime's current state.
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []RuntimeCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// Selector is used for auto-scaling
	Selector string `json:"selector,omitempty"` // this must be the string form of the selector

	// CacheAffinity represents the runtime worker pods node affinity including node selector
	CacheAffinity *corev1.NodeAffinity `json:"cacheAffinity,omitempty"`

	// RuntimeComponentStatusCollection contains the status of runtime components
	RuntimeComponentStatusCollection `json:",inline"`

	// MountPointsStatuses represents the status of mount points specified in the bounded dataset
	MountPointsStatuses []MountPointStatus `json:"mountPointsStatus,omitempty"`
}

// MountPointStatus represents the status of a mount point
type MountPointStatus struct {
	// Mount represents the mount point configuration from the bounded dataset
	Mount `json:"mount,omitempty"`

	// MountTime represents the time when the last mount operation occurred
	// If MountTime is earlier than master starting time, remount will be required
	MountTime *metav1.Time `json:"mountTime,omitempty"`
}

// RuntimeComponentStatusCollection defines the status collection for all components of a runtime.
// It includes statuses for master, worker, and client components.
type RuntimeComponentStatusCollection struct {
	// Master represents the status of the master component
	Master RuntimeComponentStatus `json:"master,omitempty"`

	// Worker represents the status of the worker component
	Worker RuntimeComponentStatus `json:"worker,omitempty"`

	// Client represents the status of the client component
	Client RuntimeComponentStatus `json:"client,omitempty"`
}

// RuntimeComponentStatus defines the observed state of a specific runtime component.
type RuntimeComponentStatus struct {
	// Phase is the current running phase of the component
	Phase RuntimePhase `json:"phase"`

	// Reason indicates the reason for the component's condition transition
	Reason string `json:"reason,omitempty"`

	// ReadyReplicas is the number of replicas that are ready to serve requests
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// CurrentReplicas is the current number of replicas running for the component
	CurrentReplicas int32 `json:"currentReplicas,omitempty"`

	// DesiredReplicas is the desired number of replicas for the component
	DesiredReplicas int32 `json:"desiredReplicas,omitempty"`

	// UnavailableReplicas is the number of replicas that are currently unavailable
	UnavailableReplicas int32 `json:"unavailableReplicas,omitempty"`

	// AvailableReplicas is the number of replicas that are available and ready
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`
}
