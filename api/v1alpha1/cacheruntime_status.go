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

// CacheRuntimeStatus describes the observed state of a CacheRuntime
type CacheRuntimeStatus struct {
	// ValueFile is the path to the generated values file used for runtime configuration.
	// This file is generated during the setup phase and contains the resolved configuration.
	// +optional
	ValueFile string `json:"valueFile,omitempty"`

	// ConfigFile is the path to the engine-specific configuration file.
	// This file contains the cache engine's native configuration format.
	// +optional
	ConfigFile string `json:"configFile,omitempty"`

	// SetupDuration is the duration spent setting up the runtime, in human-readable format (e.g., "2m30s").
	// This helps users understand the setup time for the runtime.
	// +optional
	SetupDuration string `json:"setupDuration,omitempty"`

	// Conditions represent the latest available observations of the runtime's state.
	// Known condition types include "Ready", "Available", etc.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []RuntimeCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// Selector is the label selector in string form, used for service discovery and auto-scaling.
	// This is the serialized form of a label selector that can be used to find runtime pods.
	// +optional
	Selector string `json:"selector,omitempty"`

	// CacheAffinity describes the node affinity for cache worker pods.
	// This includes node selector terms that specify which nodes are suitable for cache placement.
	// +optional
	CacheAffinity *corev1.NodeAffinity `json:"cacheAffinity,omitempty"`

	// RuntimeComponentStatusCollection contains the status of runtime components (master, worker, client).
	RuntimeComponentStatusCollection `json:",inline"`

	// MountPoints represents the status of mount points specified in the bound dataset.
	// Each entry tracks the mount configuration and the time of the last successful mount.
	// +optional
	MountPoints []MountPointStatus `json:"mountPoints,omitempty"`
}

// MountPointStatus describes the status of a single mount point in the dataset
type MountPointStatus struct {
	// Mount contains the mount point configuration from the bound dataset.
	// This includes the remote path, mount options, and other mount-specific settings.
	Mount `json:"mount,omitempty"`

	// MountTime is the timestamp of the last successful mount operation.
	// If MountTime is earlier than the master component's start time, a remount will be required.
	// +optional
	MountTime *metav1.Time `json:"mountTime,omitempty"`
}

// RuntimeComponentStatusCollection describes the status of all runtime components.
// It provides a unified view of master, worker, and client component states.
type RuntimeComponentStatusCollection struct {
	// Master is the observed state of the master component.
	// +optional
	Master RuntimeComponentStatus `json:"master,omitempty"`

	// Worker is the observed state of the worker component.
	// +optional
	Worker RuntimeComponentStatus `json:"worker,omitempty"`

	// Client is the observed state of the client (FUSE) component.
	// +optional
	Client RuntimeComponentStatus `json:"client,omitempty"`
}

// RuntimeComponentStatus describes the observed state of a runtime component.
// It follows the standard Kubernetes pattern for tracking workload status.
type RuntimeComponentStatus struct {
	// Phase is the current lifecycle phase of the component.
	// Known phases include: "Pending", "Running", "Failed", etc.
	// +kubebuilder:validation:Required
	Phase RuntimePhase `json:"phase"`

	// Reason is a brief, machine-readable string that gives the reason for the current phase.
	// This is useful for understanding why a component is in a particular state.
	// +optional
	Reason string `json:"reason,omitempty"`

	// Message is a human-readable message indicating details about the component's status.
	// +optional
	Message string `json:"message,omitempty"`

	// ReadyReplicas is the number of pods with a Ready condition.
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// CurrentReplicas is the current number of replicas running for the component
	// +optional
	CurrentReplicas int32 `json:"currentReplicas,omitempty"`

	// DesiredReplicas is the desired number of replicas as specified in the spec.
	// +optional
	DesiredReplicas int32 `json:"desiredReplicas,omitempty"`

	// UnavailableReplicas is the number of pods that are not available.
	// A pod is considered unavailable if it is not ready or has been terminated.
	// +optional
	UnavailableReplicas int32 `json:"unavailableReplicas,omitempty"`

	// AvailableReplicas is the number of pods that are available and ready to serve requests.
	// This count includes only pods that have passed readiness checks.
	// +optional
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`
}
