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

type Policy string

const (
	// Once run data migrate once, default policy is Once
	Once Policy = "Once"

	// Cron run data migrate by cron
	Cron Policy = "Cron"

	// OnEvent run data migrate when event occurs
	OnEvent Policy = "OnEvent"
)

// InitUsersSpec is a description of the initialize the users for runtime
type InitUsersSpec struct {

	// Image for initialize the users for runtime(e.g. alluxio/alluxio-User init)
	Image string `json:"image,omitempty"`

	// Image Tag for initialize the users for runtime(e.g. 2.3.0-SNAPSHOT)
	ImageTag string `json:"imageTag,omitempty"`

	// One of the three policies: `Always`, `IfNotPresent`, `Never`
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	// Environment variables that will be used by initialize the users for runtime
	Env map[string]string `json:"env,omitempty"`

	// Resources that will be requested by initialize the users for runtime. <br>
	// <br>
	// Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources
	// already allocated to the pod.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// HCFS Endpoint info
type HCFSStatus struct {
	// Endpoint for accessing
	Endpoint string `json:"endpoint,omitempty"`

	// Underlayer HCFS Compatible Version
	UnderlayerFileSystemVersion string `json:"underlayerFileSystemVersion,omitempty"`
}

// API Gateway
type APIGatewayStatus struct {
	// Endpoint for accessing
	Endpoint string `json:"endpoint,omitempty"`
}

// VersionSpec represents the settings for the  version that fluid is orchestrating.
type VersionSpec struct {
	// Image (e.g. alluxio/alluxio)
	Image string `json:"image,omitempty"`

	// Image tag (e.g. 2.3.0-SNAPSHOT)
	ImageTag string `json:"imageTag,omitempty"`

	// One of the three policies: `Always`, `IfNotPresent`, `Never`
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`
}

// Condition explains the transitions on phase
type Condition struct {
	// Type of condition, either `Complete` or `Failed`
	Type common.ConditionType `json:"type"`
	// Status of the condition, one of `True`, `False` or `Unknown`
	Status corev1.ConditionStatus `json:"status"`
	// Reason for the condition's last transition
	Reason string `json:"reason,omitempty"`
	// Message is a human-readable message indicating details about the transition
	Message string `json:"message,omitempty"`
	// LastProbeTime describes last time this condition was updated.
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`
	// LastTransitionTime describes last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// User explains the user and group to run a Container
type User struct {
	// The uid to run the alluxio runtime
	UID *int64 `json:"uid"`
	// The gid to run the alluxio runtime
	GID *int64 `json:"gid"`
	// The user name to run the alluxio runtime
	UserName string `json:"user"`
	// The group name to run the alluxio runtime
	GroupName string `json:"group"`
}

// Metadata defines subgroup properties of metav1.ObjectMeta
type Metadata struct {
	PodMetadata `json:",inline"`

	Selector metav1.GroupKind `json:"selector,omitempty"`
}

// PodMetadata defines subgroup properties of metav1.ObjectMeta
type PodMetadata struct {
	// Labels are labels of pod specification
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations are annotations of pod specification
	Annotations map[string]string `json:"annotations,omitempty"`
}

// VolumeSource defines volume source and volume claim template.
type VolumeSource struct {
	corev1.VolumeSource `json:",inline"`
}

// CleanCachePolicy defines policies when cleaning cache
type CleanCachePolicy struct {
	// Optional duration in seconds the cache needs to clean gracefully. May be decreased in delete runtime request.
	// Value must be non-negative integer. The value zero indicates clean immediately via the timeout
	// command (no opportunity to shut down).
	// If this value is nil, the default grace period will be used instead.
	// The grace period is the duration in seconds after the processes running in the pod are sent
	// a termination signal and the time when the processes are forcibly halted with timeout command.
	// Set this value longer than the expected cleanup time for your process.
	// +kubebuilder:default=60
	// +optional
	GracePeriodSeconds *int32 `json:"gracePeriodSeconds,omitempty"`

	// Optional max retry Attempts when cleanCache function returns an error after execution, runtime attempts
	// to run it three more times by default. With Maximum Retry Attempts, you can customize the maximum number
	// of retries. This gives you the option to continue processing retries.
	// +kubebuilder:default=3
	// +optional
	MaxRetryAttempts *int32 `json:"maxRetryAttempts,omitempty"`
}

// MetadataSyncPolicy defines policies when syncing metadata
type MetadataSyncPolicy struct {
	// AutoSync enables automatic metadata sync when setting up a runtime. If not set, it defaults to true.
	// +kubebuilder:default=true
	// +optional
	AutoSync *bool `json:"autoSync,omitempty"`
}

func (msb *MetadataSyncPolicy) AutoSyncEnabled() bool {
	return msb.AutoSync == nil || *msb.AutoSync
}
