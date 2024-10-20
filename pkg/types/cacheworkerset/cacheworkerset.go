package cacheworkerset

import (
	appspub "github.com/openkruise/kruise/apis/apps/pub"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	MaxMinReadySeconds = 300
)

type CacheWorkerSetUpdateStrategy struct {
	Type apps.StatefulSetUpdateStrategyType

	RollingUpdate *RollingUpdateCacheWorkerSetStrategy
}

type RollingUpdateCacheWorkerSetStrategy struct {
	Partition *int32

	MaxUnavailable *intstr.IntOrString

	PodUpdatePolicy PodUpdateStrategyType

	Paused bool

	UnorderedUpdate *UnorderedUpdateStrategy

	InPlaceUpdateStrategy *appspub.InPlaceUpdateStrategy

	MinReadySeconds *int32
}

type UnorderedUpdateStrategy struct {
	PriorityStrategy *appspub.UpdatePriorityStrategy
}

type PodUpdateStrategyType string

const (
	RecreatePodUpdateStrategyType PodUpdateStrategyType = "ReCreate"

	InPlaceIfPossiblePodUpdateStrategyType PodUpdateStrategyType = "InPlaceIfPossible"

	InPlaceOnlyPodUpdateStrategyType PodUpdateStrategyType = "InPlaceOnly"
)

type PersistentVolumeClaimRetentionPolicyType string

const (
	RetainPersistentVolumeClaimRetentionPolicyType PersistentVolumeClaimRetentionPolicyType = "Retain"

	DeletePersistentVolumeClaimRetentionPolicyType PersistentVolumeClaimRetentionPolicyType = "Delete"
)

type CacheWorkerSetPersistentVolumeClaimRetentionPolicy struct {
	WhenDeleted PersistentVolumeClaimRetentionPolicyType

	WhenScaled PersistentVolumeClaimRetentionPolicyType
}

type CacheWorkerSetSpec struct {
	Replicas *int32

	Selector *metav1.LabelSelector

	Template v1.PodTemplateSpec

	VolumeClaimTemplates []v1.PersistentVolumeClaim

	ServiceName string

	PodManagementPolicy apps.PodManagementPolicyType

	UpdateStrategy CacheWorkerSetUpdateStrategy

	RevisionHistoryLimit *int32

	ReserveOrdinals []int

	Lifecycle *appspub.Lifecycle

	ScaleStrategy *CacheWorkerSetScaleStrategy

	PersistentVolumeClaimRetentionPolicy *CacheWorkerSetPersistentVolumeClaimRetentionPolicy
}

type CacheWorkerSetScaleStrategy struct {
	MaxUnavailable *intstr.IntOrString
}

type CacheWorkerSetStatus struct {
	ObservedGeneration int64

	Replicas int32

	ReadyReplicas int32

	AvailableReplicas int32

	CurrentReplicas int32

	UpdatedReplicas int32

	UpdatedReadyReplicas int32

	CurrentRevision string

	UpdateRevision string

	CollisionCount *int32

	Conditions []apps.StatefulSetCondition

	LabelSelector string
}

const (
	FailedCreatePod apps.StatefulSetConditionType = "FailedCreatePod"
	FailedUpdatePod apps.StatefulSetConditionType = "FailedUpdatePod"
)

type AbstractCacheWorkerSet struct {
	WorkerType WorkerType
	TypeMeta   metav1.TypeMeta
	ObjectMeta metav1.ObjectMeta

	Spec   CacheWorkerSetSpec
	Status CacheWorkerSetStatus
}

type CacheWorkerSetList struct {
	TypeMeta   metav1.TypeMeta
	ObjectMeta metav1.ListMeta
	Items      []AbstractCacheWorkerSet
}
