package apis

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
import (
	v1 "k8s.io/api/core/v1"
)

const (
	AdvancedStatefulSetPodNameLabel = "AdvancedStatefulSet.kubernetes.io/pod-name"
)

// TODO: add `+genclient:method=ApplyScale,verb=apply,subresource=scale,input=k8s.io/api/autoscaling/v1.Scale,result=k8s.io/api/autoscaling/v1.Scale`
// ref: https://github.com/kubernetes/kubernetes/issues/119360

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:method=GetScale,verb=get,subresource=scale,result=k8s.io/api/autoscaling/v1.Scale
// +genclient:method=UpdateScale,verb=update,subresource=scale,input=k8s.io/api/autoscaling/v1.Scale,result=k8s.io/api/autoscaling/v1.Scale

type AdvancedStatefulSet struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec defines the desired identities of pods in this set.
	// +optional
	Spec AdvancedStatefulSetSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Status is the current status of Pods in this AdvancedStatefulSet. This data
	// may be out of date by some window of time.
	// +optional
	Status AdvancedStatefulSetStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// PodManagementPolicyType defines the policy for creating pods under a stateful set.
type PodManagementPolicyType string

const (
	// OrderedReadyPodManagement will create pods in strictly increasing order on
	// scale up and strictly decreasing order on scale down, progressing only when
	// the previous pod is ready or terminated. At most one pod will be changed
	// at any time.
	OrderedReadyPodManagement PodManagementPolicyType = "OrderedReady"
	// ParallelPodManagement will create and delete pods as soon as the stateful set
	// replica count is changed, and will not wait for pods to be ready or complete
	// termination.
	ParallelPodManagement PodManagementPolicyType = "Parallel"
)

// AdvancedStatefulSetUpdateStrategy indicates the strategy that the AdvancedStatefulSet
// controller will use to perform updates. It includes any additional parameters
// necessary to perform the update for the indicated strategy.
type AdvancedStatefulSetUpdateStrategy struct {
	// Type indicates the type of the AdvancedStatefulSetUpdateStrategy.
	// Default is RollingUpdate.
	// +optional
	Type AdvancedStatefulSetUpdateStrategyType `json:"type,omitempty" protobuf:"bytes,1,opt,name=type,casttype=AdvancedStatefulSetStrategyType"`
	// RollingUpdate is used to communicate parameters when Type is RollingUpdateAdvancedStatefulSetStrategyType.
	// +optional
	RollingUpdate *RollingUpdateAdvancedStatefulSetStrategy `json:"rollingUpdate,omitempty" protobuf:"bytes,2,opt,name=rollingUpdate"`
}

// AdvancedStatefulSetUpdateStrategyType is a string enumeration type that enumerates
// all possible update strategies for the AdvancedStatefulSet controller.
type AdvancedStatefulSetUpdateStrategyType string

const (
	// RollingUpdateAdvancedStatefulSetStrategyType indicates that update will be
	// applied to all Pods in the AdvancedStatefulSet with respect to the AdvancedStatefulSet
	// ordering constraints. When a scale operation is performed with this
	// strategy, new Pods will be created from the specification version indicated
	// by the AdvancedStatefulSet's updateRevision.
	RollingUpdateAdvancedStatefulSetStrategyType AdvancedStatefulSetUpdateStrategyType = "RollingUpdate"
	// OnDeleteAdvancedStatefulSetStrategyType triggers the legacy behavior. Version
	// tracking and ordered rolling restarts are disabled. Pods are recreated
	// from the AdvancedStatefulSetSpec when they are manually deleted. When a scale
	// operation is performed with this strategy,specification version indicated
	// by the AdvancedStatefulSet's currentRevision.
	OnDeleteAdvancedStatefulSetStrategyType AdvancedStatefulSetUpdateStrategyType = "OnDelete"
)

// RollingUpdateAdvancedStatefulSetStrategy is used to communicate parameter for RollingUpdateAdvancedStatefulSetStrategyType.
type RollingUpdateAdvancedStatefulSetStrategy struct {
	// Partition indicates the ordinal at which the AdvancedStatefulSet should be
	// partitioned.
	// Default value is 0.
	// +optional
	Partition *int32 `json:"partition,omitempty" protobuf:"varint,1,opt,name=partition"`
}

// A AdvancedStatefulSetSpec is the specification of a AdvancedStatefulSet.
type AdvancedStatefulSetSpec struct {
	// replicas is the desired number of replicas of the given Template.
	// These are replicas in the sense that they are instantiations of the
	// same Template, but individual replicas also have a consistent identity.
	// If unspecified, defaults to 1.
	// TODO: Consider a rename of this field.
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	// selector is a label query over pods that should match the replica count.
	// It must match the pod template's labels.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	Selector *metav1.LabelSelector `json:"selector" protobuf:"bytes,2,opt,name=selector"`

	// template is the object that describes the pod that will be created if
	// insufficient replicas are detected. Each pod stamped out by the AdvancedStatefulSet
	// will fulfill this Template, but have a unique identity from the rest
	// of the AdvancedStatefulSet.
	Template v1.PodTemplateSpec `json:"template" protobuf:"bytes,3,opt,name=template"`

	// volumeClaimTemplates is a list of claims that pods are allowed to reference.
	// The AdvancedStatefulSet controller is responsible for mapping network identities to
	// claims in a way that maintains the identity of a pod. Every claim in
	// this list must have at least one matching (by name) volumeMount in one
	// container in the template. A claim in this list takes precedence over
	// any volumes in the template, with the same name.
	// TODO: Define the behavior if a claim already exists with the same name.
	// +optional
	VolumeClaimTemplates []v1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty" protobuf:"bytes,4,rep,name=volumeClaimTemplates"`

	// serviceName is the name of the service that governs this AdvancedStatefulSet.
	// This service must exist before the AdvancedStatefulSet, and is responsible for
	// the network identity of the set. Pods get DNS/hostnames that follow the
	// pattern: pod-specific-string.serviceName.default.svc.cluster.local
	// where "pod-specific-string" is managed by the AdvancedStatefulSet controller.
	ServiceName string `json:"serviceName" protobuf:"bytes,5,opt,name=serviceName"`

	// podManagementPolicy controls how pods are created during initial scale up,
	// when replacing pods on nodes, or when scaling down. The default policy is
	// `OrderedReady`, where pods are created in increasing order (pod-0, then
	// pod-1, etc) and the controller will wait until each pod is ready before
	// continuing. When scaling down, the pods are removed in the opposite order.
	// The alternative policy is `Parallel` which will create pods in parallel
	// to match the desired scale without waiting, and on scale down will delete
	// all pods at once.
	// +optional
	PodManagementPolicy PodManagementPolicyType `json:"podManagementPolicy,omitempty" protobuf:"bytes,6,opt,name=podManagementPolicy,casttype=PodManagementPolicyType"`

	// updateStrategy indicates the AdvancedStatefulSetUpdateStrategy that will be
	// employed to update Pods in the AdvancedStatefulSet when a revision is made to
	// Template.
	UpdateStrategy AdvancedStatefulSetUpdateStrategy `json:"updateStrategy,omitempty" protobuf:"bytes,7,opt,name=updateStrategy"`

	// revisionHistoryLimit is the maximum number of revisions that will
	// be maintained in the AdvancedStatefulSet's revision history. The revision history
	// consists of all revisions not represented by a currently applied
	// AdvancedStatefulSetSpec version. The default value is 10.
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty" protobuf:"varint,8,opt,name=revisionHistoryLimit"`
}

// AdvancedStatefulSetStatus represents the current state of a AdvancedStatefulSet.
type AdvancedStatefulSetStatus struct {
	// observedGeneration is the most recent generation observed for this AdvancedStatefulSet. It corresponds to the
	// AdvancedStatefulSet's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`

	// replicas is the number of Pods created by the AdvancedStatefulSet controller.
	Replicas int32 `json:"replicas" protobuf:"varint,2,opt,name=replicas"`

	// readyReplicas is the number of Pods created by the AdvancedStatefulSet controller that have a Ready Condition.
	ReadyReplicas int32 `json:"readyReplicas,omitempty" protobuf:"varint,3,opt,name=readyReplicas"`

	// currentReplicas is the number of Pods created by the AdvancedStatefulSet controller from the AdvancedStatefulSet version
	// indicated by currentRevision.
	CurrentReplicas int32 `json:"currentReplicas,omitempty" protobuf:"varint,4,opt,name=currentReplicas"`

	// updatedReplicas is the number of Pods created by the AdvancedStatefulSet controller from the AdvancedStatefulSet version
	// indicated by updateRevision.
	UpdatedReplicas int32 `json:"updatedReplicas,omitempty" protobuf:"varint,5,opt,name=updatedReplicas"`

	// currentRevision, if not empty, indicates the version of the AdvancedStatefulSet used to generate Pods in the
	// sequence [0,currentReplicas).
	CurrentRevision string `json:"currentRevision,omitempty" protobuf:"bytes,6,opt,name=currentRevision"`

	// updateRevision, if not empty, indicates the version of the AdvancedStatefulSet used to generate Pods in the sequence
	// [replicas-updatedReplicas,replicas)
	UpdateRevision string `json:"updateRevision,omitempty" protobuf:"bytes,7,opt,name=updateRevision"`

	// collisionCount is the count of hash collisions for the AdvancedStatefulSet. The AdvancedStatefulSet controller
	// uses this field as a collision avoidance mechanism when it needs to create the name for the
	// newest ControllerRevision.
	// +optional
	CollisionCount *int32 `json:"collisionCount,omitempty" protobuf:"varint,9,opt,name=collisionCount"`

	// Represents the latest available observations of a AdvancedStatefulSet's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []AdvancedStatefulSetCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,10,rep,name=conditions"`
}

type AdvancedStatefulSetConditionType string

// AdvancedStatefulSetCondition describes the state of a AdvancedStatefulSet at a certain point.
type AdvancedStatefulSetCondition struct {
	// Type of AdvancedStatefulSet condition.
	Type AdvancedStatefulSetConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=AdvancedStatefulSetConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`
	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AdvancedStatefulSetList is a collection of AdvancedStatefulSets.
type AdvancedStatefulSetList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []AdvancedStatefulSet `json:"items" protobuf:"bytes,2,rep,name=items"`
}
