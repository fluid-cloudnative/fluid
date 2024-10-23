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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	// RolloutIDLabel is set to workload labels.
	// RolloutIDLabel is designed to distinguish each workload revision publications.
	// The value of RolloutIDLabel corresponds Rollout.Spec.RolloutID.
	RolloutIDLabel = "rollouts.kruise.io/rollout-id"

	// RolloutBatchIDLabel is patched in pod labels.
	// RolloutBatchIDLabel is the label key of batch id that will be patched to pods during rollout.
	// Only when RolloutIDLabel is set, RolloutBatchIDLabel will be patched.
	// Users can use RolloutIDLabel and RolloutBatchIDLabel to select the pods that are upgraded in some certain batch and release.
	RolloutBatchIDLabel = "rollouts.kruise.io/rollout-batch-id"

	// RollbackInBatchAnnotation is set to rollout annotations.
	// RollbackInBatchAnnotation allow use disable quick rollback, and will roll back in batch style.
	RollbackInBatchAnnotation = "rollouts.kruise.io/rollback-in-batch"

	// RolloutStyleAnnotation define the rolling behavior for Deployment.
	// must be "partition" or "canary":
	// * "partition" means rolling in batches just like CloneSet, and will NOT create any extra Workload;
	// * "canary" means rolling in canary way, and will create a canary Workload.
	// Currently, only Deployment support both "partition" and "canary" rolling styles.
	// For other workload types, they only support "partition" styles.
	// Defaults to "canary" to Deployment.
	// Defaults to "partition" to the others.
	RolloutStyleAnnotation = "rollouts.kruise.io/rolling-style"
)

// RolloutSpec defines the desired state of Rollout
type RolloutSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// ObjectRef indicates workload
	ObjectRef ObjectRef `json:"objectRef"`
	// rollout strategy
	Strategy RolloutStrategy `json:"strategy"`
	// DeprecatedRolloutID is the deprecated field.
	// It is recommended that configure RolloutId in workload.annotations[rollouts.kruise.io/rollout-id].
	// RolloutID should be changed before each workload revision publication.
	// It is to distinguish consecutive multiple workload publications and rollout progress.
	DeprecatedRolloutID string `json:"rolloutID,omitempty"`
}

type ObjectRef struct {
	// WorkloadRef contains enough information to let you identify a workload for Rollout
	// Batch release of the bypass
	WorkloadRef *WorkloadRef `json:"workloadRef,omitempty"`

	// revisionRef
	// Fully managed batch publishing capability
	//RevisionRef *ControllerRevisionRef `json:"revisionRef,omitempty"`
}

type ObjectRefType string

const (
	WorkloadRefType = "workloadRef"
	RevisionRefType = "revisionRef"
)

// WorkloadRef holds a references to the Kubernetes object
type WorkloadRef struct {
	// API Version of the referent
	APIVersion string `json:"apiVersion"`
	// Kind of the referent
	Kind string `json:"kind"`
	// Name of the referent
	Name string `json:"name"`
}

// RolloutStrategy defines strategy to apply during next rollout
type RolloutStrategy struct {
	// Paused indicates that the Rollout is paused.
	// Default value is false
	Paused bool `json:"paused,omitempty"`
	// +optional
	Canary *CanaryStrategy `json:"canary,omitempty"`
	// +optional
	// BlueGreen *BlueGreenStrategy `json:"blueGreen,omitempty"`
}

type RolloutStrategyType string

const (
	RolloutStrategyCanary    RolloutStrategyType = "canary"
	RolloutStrategyBlueGreen RolloutStrategyType = "blueGreen"
)

// CanaryStrategy defines parameters for a Replica Based Canary
type CanaryStrategy struct {
	// Steps define the order of phases to execute release in batches(20%, 40%, 60%, 80%, 100%)
	// +optional
	Steps []CanaryStep `json:"steps,omitempty"`
	// TrafficRoutings hosts all the supported service meshes supported to enable more fine-grained traffic routing
	// todo current only support one TrafficRouting
	TrafficRoutings []*TrafficRouting `json:"trafficRoutings,omitempty"`
	// FailureThreshold indicates how many failed pods can be tolerated in all upgraded pods.
	// Only when FailureThreshold are satisfied, Rollout can enter ready state.
	// If FailureThreshold is nil, Rollout will use the MaxUnavailable of workload as its
	// FailureThreshold.
	// Defaults to nil.
	FailureThreshold *intstr.IntOrString `json:"failureThreshold,omitempty"`
	// MetricsAnalysis *MetricsAnalysisBackground `json:"metricsAnalysis,omitempty"`
}

// CanaryStep defines a step of a canary workload.
type CanaryStep struct {
	// Weight indicate how many percentage of traffic the canary pods should receive
	// +optional
	Weight *int32 `json:"weight,omitempty"`
	// Replicas is the number of expected canary pods in this batch
	// it can be an absolute number (ex: 5) or a percentage of total pods.
	Replicas *intstr.IntOrString `json:"replicas,omitempty"`
	// Pause defines a pause stage for a rollout, manual or auto
	// +optional
	Pause RolloutPause `json:"pause,omitempty"`
	// Matches define conditions used for matching the incoming HTTP requests to canary service.
	// Each match is independent, i.e. this rule will be matched if **any** one of the matches is satisfied.
	// If Gateway API, current only support one match.
	// And cannot support both weight and matches, if both are configured, then matches takes precedence.
	Matches []HttpRouteMatch `json:"matches,omitempty"`
}

type HttpRouteMatch struct {
	// Headers specifies HTTP request header matchers. Multiple match values are
	// ANDed together, meaning, a request must match all the specified headers
	// to select the route.
	// +kubebuilder:validation:MaxItems=16
	Headers []gatewayv1alpha2.HTTPHeaderMatch `json:"headers,omitempty"`
}

// RolloutPause defines a pause stage for a rollout
type RolloutPause struct {
	// Duration the amount of time to wait before moving to the next step.
	// +optional
	Duration *int32 `json:"duration,omitempty"`
}

// TrafficRouting hosts all the different configuration for supported service meshes to enable more fine-grained traffic routing
type TrafficRouting struct {
	// Service holds the name of a service which selects pods with stable version and don't select any pods with canary version.
	Service string `json:"service"`
	// Optional duration in seconds the traffic provider(e.g. nginx ingress controller) consumes the service, ingress configuration changes gracefully.
	GracePeriodSeconds int32 `json:"gracePeriodSeconds,omitempty"`
	// Ingress holds Ingress specific configuration to route traffic, e.g. Nginx, Alb.
	Ingress *IngressTrafficRouting `json:"ingress,omitempty"`
	// Gateway holds Gateway specific configuration to route traffic
	// Gateway configuration only supports >= v0.4.0 (v1alpha2).
	Gateway *GatewayTrafficRouting `json:"gateway,omitempty"`
}

// IngressTrafficRouting configuration for ingress controller to control traffic routing
type IngressTrafficRouting struct {
	// ClassType refers to the type of `Ingress`.
	// current support nginx, aliyun-alb. default is nginx.
	// +optional
	ClassType string `json:"classType,omitempty"`
	// Name refers to the name of an `Ingress` resource in the same namespace as the `Rollout`
	Name string `json:"name"`
}

// GatewayTrafficRouting configuration for gateway api
type GatewayTrafficRouting struct {
	// HTTPRouteName refers to the name of an `HTTPRoute` resource in the same namespace as the `Rollout`
	HTTPRouteName *string `json:"httpRouteName,omitempty"`
	// TCPRouteName *string `json:"tcpRouteName,omitempty"`
	// UDPRouteName *string `json:"udpRouteName,omitempty"`
}

// RolloutStatus defines the observed state of Rollout
type RolloutStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// observedGeneration is the most recent generation observed for this Rollout.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Canary describes the state of the canary rollout
	// +optional
	CanaryStatus *CanaryStatus `json:"canaryStatus,omitempty"`
	// Conditions a list of conditions a rollout can have.
	// +optional
	Conditions []RolloutCondition `json:"conditions,omitempty"`
	// +optional
	//BlueGreenStatus *BlueGreenStatus `json:"blueGreenStatus,omitempty"`
	// Phase is the rollout phase.
	Phase RolloutPhase `json:"phase,omitempty"`
	// Message provides details on why the rollout is in its current phase
	Message string `json:"message,omitempty"`
}

// RolloutCondition describes the state of a rollout at a certain point.
type RolloutCondition struct {
	// Type of rollout condition.
	Type RolloutConditionType `json:"type"`
	// Phase of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason"`
	// A human readable message indicating details about the transition.
	Message string `json:"message"`
}

// RolloutConditionType defines the conditions of Rollout
type RolloutConditionType string

// These are valid conditions of a rollout.
const (
	// RolloutConditionProgressing means the rollout is progressing. Progress for a rollout is
	// considered when a new replica set is created or adopted, when pods scale
	// up or old pods scale down, or when the services are updated. Progress is not estimated
	// for paused rollouts.
	RolloutConditionProgressing RolloutConditionType = "Progressing"
	// Progressing Reason
	ProgressingReasonInitializing = "Initializing"
	ProgressingReasonInRolling    = "InRolling"
	ProgressingReasonFinalising   = "Finalising"
	ProgressingReasonCompleted    = "Completed"
	ProgressingReasonCancelling   = "Cancelling"
	ProgressingReasonPaused       = "Paused"

	// RolloutConditionSucceeded indicates whether rollout is succeeded or failed.
	RolloutConditionSucceeded RolloutConditionType = "Succeeded"

	// Terminating condition
	RolloutConditionTerminating RolloutConditionType = "Terminating"
	// Terminating Reason
	TerminatingReasonInTerminating = "InTerminating"
	TerminatingReasonCompleted     = "Completed"
)

// CanaryStatus status fields that only pertain to the canary rollout
type CanaryStatus struct {
	// observedWorkloadGeneration is the most recent generation observed for this Rollout ref workload generation.
	ObservedWorkloadGeneration int64 `json:"observedWorkloadGeneration,omitempty"`
	// ObservedRolloutID will record the newest spec.RolloutID if status.canaryRevision equals to workload.updateRevision
	ObservedRolloutID string `json:"observedRolloutID,omitempty"`
	// RolloutHash from rollout.spec object
	RolloutHash string `json:"rolloutHash,omitempty"`
	// StableRevision indicates the revision of stable pods
	StableRevision string `json:"stableRevision,omitempty"`
	// CanaryRevision is calculated by rollout based on podTemplateHash, and the internal logic flow uses
	// It may be different from rs podTemplateHash in different k8s versions, so it cannot be used as service selector label
	CanaryRevision string `json:"canaryRevision"`
	// pod template hash is used as service selector label
	PodTemplateHash string `json:"podTemplateHash"`
	// CanaryReplicas the numbers of canary revision pods
	CanaryReplicas int32 `json:"canaryReplicas"`
	// CanaryReadyReplicas the numbers of ready canary revision pods
	CanaryReadyReplicas int32 `json:"canaryReadyReplicas"`
	// CurrentStepIndex defines the current step of the rollout is on. If the current step index is null, the
	// controller will execute the rollout.
	// +optional
	CurrentStepIndex int32           `json:"currentStepIndex"`
	CurrentStepState CanaryStepState `json:"currentStepState"`
	Message          string          `json:"message,omitempty"`
	LastUpdateTime   *metav1.Time    `json:"lastUpdateTime,omitempty"`
}

type CanaryStepState string

const (
	CanaryStepStateUpgrade         CanaryStepState = "StepUpgrade"
	CanaryStepStateTrafficRouting  CanaryStepState = "StepTrafficRouting"
	CanaryStepStateMetricsAnalysis CanaryStepState = "StepMetricsAnalysis"
	CanaryStepStatePaused          CanaryStepState = "StepPaused"
	CanaryStepStateReady           CanaryStepState = "StepReady"
	CanaryStepStateCompleted       CanaryStepState = "Completed"
)

// RolloutPhase are a set of phases that this rollout
type RolloutPhase string

const (
	// RolloutPhaseInitial indicates a rollout is Initial
	RolloutPhaseInitial RolloutPhase = "Initial"
	// RolloutPhaseHealthy indicates a rollout is healthy
	RolloutPhaseHealthy RolloutPhase = "Healthy"
	// RolloutPhaseProgressing indicates a rollout is not yet healthy but still making progress towards a healthy state
	RolloutPhaseProgressing RolloutPhase = "Progressing"
	// RolloutPhaseTerminating indicates a rollout is terminated
	RolloutPhaseTerminating RolloutPhase = "Terminating"
)

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.phase",description="The rollout status phase"
// +kubebuilder:printcolumn:name="CANARY_STEP",type="integer",JSONPath=".status.canaryStatus.currentStepIndex",description="The rollout canary status step"
// +kubebuilder:printcolumn:name="CANARY_STATE",type="string",JSONPath=".status.canaryStatus.currentStepState",description="The rollout canary status step state"
// +kubebuilder:printcolumn:name="MESSAGE",type="string",JSONPath=".status.message",description="The rollout canary status message"
// +kubebuilder:printcolumn:name="AGE",type=date,JSONPath=".metadata.creationTimestamp"

// Rollout is the Schema for the rollouts API
type Rollout struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RolloutSpec   `json:"spec,omitempty"`
	Status RolloutStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RolloutList contains a list of Rollout
type RolloutList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Rollout `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Rollout{}, &RolloutList{})
}
