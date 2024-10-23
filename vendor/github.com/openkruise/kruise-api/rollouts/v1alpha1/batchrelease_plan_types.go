/*
Copyright 2022 The Kruise Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ReleasePlan fines the details of the release plan
type ReleasePlan struct {
	// Batches is the details on each batch of the ReleasePlan.
	//Users can specify their batch plan in this field, such as:
	// batches:
	// - canaryReplicas: 1  # batches 0
	// - canaryReplicas: 2  # batches 1
	// - canaryReplicas: 5  # batches 2
	// Not that these canaryReplicas should be a non-decreasing sequence.
	// +optional
	Batches []ReleaseBatch `json:"batches"`
	// All pods in the batches up to the batchPartition (included) will have
	// the target resource specification while the rest still is the stable revision.
	// This is designed for the operators to manually rollout.
	// Default is nil, which means no partition and will release all batches.
	// BatchPartition start from 0.
	// +optional
	BatchPartition *int32 `json:"batchPartition,omitempty"`
	// RolloutID indicates an id for each rollout progress
	RolloutID string `json:"rolloutID,omitempty"`
	// FailureThreshold indicates how many failed pods can be tolerated in all upgraded pods.
	// Only when FailureThreshold are satisfied, Rollout can enter ready state.
	// If FailureThreshold is nil, Rollout will use the MaxUnavailable of workload as its
	// FailureThreshold.
	// Defaults to nil.
	FailureThreshold *intstr.IntOrString `json:"failureThreshold,omitempty"`
	// FinalizingPolicy define the behavior of controller when phase enter Finalizing
	// Defaults to "Immediate"
	FinalizingPolicy FinalizingPolicyType `json:"finalizingPolicy,omitempty"`
}

type FinalizingPolicyType string

const (
	// WaitResumeFinalizingPolicyType will wait workload to be resumed, which means
	// controller will be hold at Finalizing phase util all pods of workload is upgraded.
	// WaitResumeFinalizingPolicyType only works in canary-style BatchRelease controller.
	WaitResumeFinalizingPolicyType FinalizingPolicyType = "WaitResume"
	// ImmediateFinalizingPolicyType will not to wait workload to be resumed.
	ImmediateFinalizingPolicyType FinalizingPolicyType = "Immediate"
)

// ReleaseBatch is used to describe how each batch release should be
type ReleaseBatch struct {
	// CanaryReplicas is the number of upgraded pods that should have in this batch.
	// it can be an absolute number (ex: 5) or a percentage of workload replicas.
	// batches[i].canaryReplicas should less than or equal to batches[j].canaryReplicas if i < j.
	CanaryReplicas intstr.IntOrString `json:"canaryReplicas"`
}

// BatchReleaseStatus defines the observed state of a release plan
type BatchReleaseStatus struct {
	// Conditions represents the observed process state of each phase during executing the release plan.
	Conditions []RolloutCondition `json:"conditions,omitempty"`
	// CanaryStatus describes the state of the canary rollout.
	CanaryStatus BatchReleaseCanaryStatus `json:"canaryStatus,omitempty"`
	// StableRevision is the pod-template-hash of stable revision pod template.
	StableRevision string `json:"stableRevision,omitempty"`
	// UpdateRevision is the pod-template-hash of update revision pod template.
	UpdateRevision string `json:"updateRevision,omitempty"`
	// ObservedGeneration is the most recent generation observed for this BatchRelease.
	// It corresponds to this BatchRelease's generation, which is updated on mutation
	// by the API Server, and only if BatchRelease Spec was changed, its generation will increase 1.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// ObservedRolloutID is the most recent rollout-id observed for this BatchRelease.
	// If RolloutID was changed, we will restart to roll out from batch 0,
	// to ensure the batch-id and rollout-id labels of Pods are correct.
	ObservedRolloutID string `json:"observedRolloutID,omitempty"`
	// ObservedWorkloadReplicas is observed replicas of target referenced workload.
	// This field is designed to deal with scaling event during rollout, if this field changed,
	// it means that the workload is scaling during rollout.
	ObservedWorkloadReplicas int32 `json:"observedWorkloadReplicas,omitempty"`
	// Count of hash collisions for creating canary Deployment. The controller uses this
	// field as a collision avoidance mechanism when it needs to create the name for the
	// newest canary Deployment.
	// +optional
	CollisionCount *int32 `json:"collisionCount,omitempty"`
	// ObservedReleasePlanHash is a hash code of observed itself spec.releasePlan.
	ObservedReleasePlanHash string `json:"observedReleasePlanHash,omitempty"`
	// Phase is the release plan phase, which indicates the current state of release
	// plan state machine in BatchRelease controller.
	Phase RolloutPhase `json:"phase,omitempty"`
}

type BatchReleaseCanaryStatus struct {
	// CurrentBatchState indicates the release state of the current batch.
	CurrentBatchState BatchReleaseBatchStateType `json:"batchState,omitempty"`
	// The current batch the rollout is working on/blocked, it starts from 0
	CurrentBatch int32 `json:"currentBatch"`
	// BatchReadyTime is the ready timestamp of the current batch or the last batch.
	// This field is updated once a batch ready, and the batches[x].pausedSeconds
	// relies on this field to calculate the real-time duration.
	BatchReadyTime *metav1.Time `json:"batchReadyTime,omitempty"`
	// UpdatedReplicas is the number of upgraded Pods.
	UpdatedReplicas int32 `json:"updatedReplicas,omitempty"`
	// UpdatedReadyReplicas is the number upgraded Pods that have a Ready Condition.
	UpdatedReadyReplicas int32 `json:"updatedReadyReplicas,omitempty"`
	// the number of pods that no need to rollback in rollback scene.
	NoNeedUpdateReplicas *int32 `json:"noNeedUpdateReplicas,omitempty"`
}

type BatchReleaseBatchStateType string

const (
	// UpgradingBatchState indicates that current batch is at upgrading pod state
	UpgradingBatchState BatchReleaseBatchStateType = "Upgrading"
	// VerifyingBatchState indicates that current batch is at verifying whether it's ready state
	VerifyingBatchState BatchReleaseBatchStateType = "Verifying"
	// ReadyBatchState indicates that current batch is at batch ready state
	ReadyBatchState BatchReleaseBatchStateType = "Ready"
)

const (
	// RolloutPhasePreparing indicates a rollout is preparing for next progress.
	RolloutPhasePreparing RolloutPhase = "Preparing"
	// RolloutPhaseFinalizing indicates a rollout is finalizing
	RolloutPhaseFinalizing RolloutPhase = "Finalizing"
	// RolloutPhaseCompleted indicates a rollout is completed/cancelled/terminated
	RolloutPhaseCompleted RolloutPhase = "Completed"
)
