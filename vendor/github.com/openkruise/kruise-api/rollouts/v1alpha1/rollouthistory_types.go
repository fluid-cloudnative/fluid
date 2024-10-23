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
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RolloutHistorySpec defines the desired state of RolloutHistory
type RolloutHistorySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Rollout indicates information of the rollout related with rollouthistory
	Rollout RolloutInfo `json:"rollout,omitempty"`
	// Workload indicates information of the workload, such as cloneset, deployment, advanced statefulset
	Workload WorkloadInfo `json:"workload,omitempty"`
	// Service indicates information of the service related with workload
	Service ServiceInfo `json:"service,omitempty"`
	// TrafficRouting indicates information of traffic route related with workload
	TrafficRouting TrafficRoutingInfo `json:"trafficRouting,omitempty"`
}

type NameAndSpecData struct {
	// Name indicates the name of object ref, such as rollout name, workload name, ingress name, etc.
	Name string `json:"name"`
	// Data indecates the spec of object ref
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Data runtime.RawExtension `json:"data,omitempty"`
}

// RolloutInfo indicates information of the rollout related
type RolloutInfo struct {
	// RolloutID indicates the new rollout
	// if there is no new RolloutID this time, ignore it and not execute RolloutHistory
	RolloutID       string `json:"rolloutID"`
	NameAndSpecData `json:",inline"`
}

// ServiceInfo indicates information of the service related
type ServiceInfo struct {
	NameAndSpecData `json:",inline"`
}

// TrafficRoutingInfo indicates information of Gateway API or Ingress
type TrafficRoutingInfo struct {
	// IngressRef indicates information of ingress
	// +optional
	Ingress *IngressInfo `json:"ingress,omitempty"`
	// HTTPRouteRef indacates information of Gateway API
	// +optional
	HTTPRoute *HTTPRouteInfo `json:"httpRoute,omitempty"`
}

// IngressInfo indicates information of the ingress related
type IngressInfo struct {
	NameAndSpecData `json:",inline"`
}

// HTTPRouteInfo indicates information of gateway API
type HTTPRouteInfo struct {
	NameAndSpecData `json:",inline"`
}

// WorkloadInfo indicates information of the workload, such as cloneset, deployment, advanced statefulset
type WorkloadInfo struct {
	metav1.TypeMeta `json:",inline"`
	NameAndSpecData `json:",inline"`
}

// RolloutHistoryStatus defines the observed state of RolloutHistory
type RolloutHistoryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Phase indicates phase of RolloutHistory, just "" or "completed"
	Phase string `json:"phase,omitempty"`
	// CanarySteps indicates the pods released each step
	CanarySteps []CanaryStepInfo `json:"canarySteps,omitempty"`
}

// CanaryStepInfo indicates the pods for a revision
type CanaryStepInfo struct {
	// CanaryStepIndex indicates step this revision
	CanaryStepIndex int32 `json:"canaryStepIndex,omitempty"`
	// Pods indicates the pods information
	Pods []Pod `json:"pods,omitempty"`
}

// Pod indicates the information of a pod, including name, ip, node_name.
type Pod struct {
	// Name indicates the node name
	Name string `json:"name,omitempty"`
	// IP indicates the pod ip
	IP string `json:"ip,omitempty"`
	// NodeName indicates the node which pod is located at
	NodeName string `json:"nodeName,omitempty"`
	// todo
	// State indicates whether the pod is ready or not
	// State string `json:"state, omitempty"`
}

// Phase indicates rollouthistory phase
const (
	PhaseCompleted string = "completed"
)

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RolloutHistory is the Schema for the rollouthistories API
type RolloutHistory struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RolloutHistorySpec   `json:"spec,omitempty"`
	Status RolloutHistoryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RolloutHistoryList contains a list of RolloutHistory
type RolloutHistoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RolloutHistory `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RolloutHistory{}, &RolloutHistoryList{})
}
