package cacheworker

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type AdvancedStatefulSetUpdateStrategyType string

// AdvancedStatefulSetUpdateStrategy indicates the strategy that the StatefulSet
// controller will use to perform updates. It includes any additional parameters
// necessary to perform the update for the indicated strategy.
// RollingUpdateStatefulSetStrategy is used to communicate parameter for RollingUpdateStatefulSetStrategyType.
type RollingUpdateStatefulSetStrategy struct {
	// Partition indicates the ordinal at which the StatefulSet should be partitioned
	// for updates. During a rolling update, all pods from ordinal Replicas-1 to
	// Partition are updated. All pods from ordinal Partition-1 to 0 remain untouched.
	// This is helpful in being able to do a canary based deployment. The default value is 0.
	Partition      *int32              `json:"partition,omitempty" protobuf:"varint,1,opt,name=partition"`
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty" protobuf:"varint,2,opt,name=maxUnavailable"`
}
type AdvancedStatefulSetUpdateStrategy struct {
	// Type indicates the type of the AdvancedStatefulSetUpdateStrategy.
	// Default is RollingUpdate.
	// +optional
	Type AdvancedStatefulSetUpdateStrategyType `json:"type,omitempty" protobuf:"bytes,1,opt,name=type,casttype=StatefulSetStrategyType"`
	// RollingUpdate is used to communicate parameters when Type is RollingUpdateStatefulSetStrategyType.
	// +optional
	RollingUpdate *RollingUpdateStatefulSetStrategy `json:"rollingUpdate,omitempty" protobuf:"bytes,2,opt,name=rollingUpdate"`
}
type PodManagementPolicyType string

// A StatefulSetSpec is the specification of a StatefulSet.
// protobuf:"<type>,<number>,<label>,name=<fieldName>"
type AdvancedStatefulSetSpec struct {
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	Selector *metav1.LabelSelector `json:"selector" protobuf:"bytes,2,opt,name=selector"`

	Template v1.PodTemplateSpec `json:"template" protobuf:"bytes,3,opt,name=template"`

	VolumeClaimTemplates []v1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty" protobuf:"bytes,4,rep,name=volumeClaimTemplates"`

	ServiceName string `json:"serviceName" protobuf:"bytes,5,opt,name=serviceName"`
	// +optional
	PodManagementPolicy PodManagementPolicyType `json:"podManagementPolicy,omitempty" protobuf:"bytes,6,opt,name=podManagementPolicy,casttype=PodManagementPolicyType"`

	// Template.
	UpdateStrategy AdvancedStatefulSetUpdateStrategy `json:"updateStrategy,omitempty" protobuf:"bytes,7,opt,name=updateStrategy"`
	// StatefulSetSpec version. The default value is 10.
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty" protobuf:"varint,8,opt,name=revisionHistoryLimit"`

	// +optional
	MinReadySeconds int32 `json:"minReadySeconds,omitempty" protobuf:"varint,9,opt,name=minReadySeconds"`

	// which is alpha.  +optional
	PersistentVolumeClaimRetentionPolicy *StatefulSetPersistentVolumeClaimRetentionPolicy `json:"persistentVolumeClaimRetentionPolicy,omitempty" protobuf:"bytes,10,opt,name=persistentVolumeClaimRetentionPolicy"`
	// +optional
	Ordinals *AdvancedStatefulSetOrdinals `json:"ordinals,omitempty" protobuf:"bytes,11,opt,name=ordinals"`
}

type AdvancedStatefulSetStatus struct {
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`
	// replicas is the number of Pods created by the StatefulSet controller.
	Replicas int32 `json:"replicas" protobuf:"varint,2,opt,name=replicas"`

	// readyReplicas is the number of pods created for this StatefulSet with a Ready Condition.
	ReadyReplicas int32 `json:"readyReplicas,omitempty" protobuf:"varint,3,opt,name=readyReplicas"`

	CurrentReplicas int32 `json:"currentReplicas,omitempty" protobuf:"varint,4,opt,name=currentReplicas"`

	UpdatedReplicas int32 `json:"updatedReplicas,omitempty" protobuf:"varint,5,opt,name=updatedReplicas"`

	CurrentRevision string `json:"currentRevision,omitempty" protobuf:"bytes,6,opt,name=currentRevision"`
	// updateRevision, if not empty, indicates the version of the StatefulSet used to generate Pods in the sequence
	// [replicas-updatedReplicas,replicas)
	UpdateRevision string `json:"updateRevision,omitempty" protobuf:"bytes,7,opt,name=updateRevision"`

	// +optional
	CollisionCount *int32 `json:"collisionCount,omitempty" protobuf:"varint,9,opt,name=collisionCount"`

	// +patchStrategy=merge
	Conditions []AdvancedStatefulSetCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,10,rep,name=conditions"`

	// +optional
	AvailableReplicas int32 `json:"availableReplicas" protobuf:"varint,11,opt,name=availableReplicas"`

	ScaleInConfig scaleInConfig `json:"scaleInConfig" protobuf:"varint,12,opt,name=scaleInConfig"`
}

type scaleInConfig struct {
	//scaleDownIndices []int32
}

type AdvancedStatefulSetConditionType string

// AdvancedStatefulSetCondition describes the state of a statefulset at a certain point.
type AdvancedStatefulSetCondition struct {
	// Type of statefulset condition.
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

// AdvancedStatefulSetList is a collection of StatefulSets.
type AdvancedStatefulSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Items is the list of stateful sets.
	Items []AdvancedStatefulSet `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// AdvancedStatefulSetOrdinals describes the policy used for replica ordinal assignment
// in this StatefulSet.
type AdvancedStatefulSetOrdinals struct {
	Start int32 `json:"start" protobuf:"varint,1,opt,name=start"`
}
type PersistentVolumeClaimRetentionPolicyType string

const (
	RetainPersistentVolumeClaimRetentionPolicyType PersistentVolumeClaimRetentionPolicyType = "Retain"
	DeletePersistentVolumeClaimRetentionPolicyType PersistentVolumeClaimRetentionPolicyType = "Delete"
)

// StatefulSetPersistentVolumeClaimRetentionPolicy describes the policy used for PVCs
// created from the StatefulSet VolumeClaimTemplates.
type StatefulSetPersistentVolumeClaimRetentionPolicy struct {
	WhenDeleted PersistentVolumeClaimRetentionPolicyType `json:"whenDeleted,omitempty" protobuf:"bytes,1,opt,name=whenDeleted,casttype=PersistentVolumeClaimRetentionPolicyType"`

	WhenScaled PersistentVolumeClaimRetentionPolicyType `json:"whenScaled,omitempty" protobuf:"bytes,2,opt,name=whenScaled,casttype=PersistentVolumeClaimRetentionPolicyType"`
}

type AdvancedStatefulSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              AdvancedStatefulSetSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            AdvancedStatefulSetStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

func (a AdvancedStatefulSet) DeepCopyObject() runtime.Object {
	//TODO implement me
	panic("implement me")
}

func (a AdvancedStatefulSetStatus) GetReplicas() int {
	//TODO implement me
	panic("implement me")
}

func (a AdvancedStatefulSetStatus) GetSpec() interface{} {
	//TODO implement me
	panic("implement me")
}
