package statefulset

import (
	"github.com/fluid-cloudnative/fluid/pkg/types/cacheworkerset/asts/apis"
	"log"
)

type ScaleDownStrategy interface {
	// AllowCreateDuringDeletion returns true if Pods can be created during deletion
	AllowCreateDuringDeletion() bool
	// ShouldWaitForRunningAndReady returns true if Pods should wait for running and ready before deletion
	ShouldWaitForRunningAndReady() bool
}

type ScaleDownStatus struct {
	// CurReplicas is the total number of replicas in the StatefulSet
	CurReplicas int
	// ScaledDown is the number of replicas that were scaled down
	ScaledDown int
	// Errors contains any errors that occurred during the scale down operation
	Errors []error
}

func (s *ScaleDownStatus) AddError(err error) {
	s.Errors = append(s.Errors, err)
}

type DefaultScaleDownOperation struct {
	PodControl  AdvancedStatefulPodControl
	StatefulSet *apis.AdvancedStatefulSet
}

func NewDefaultScaleDownOperation(
	podControl AdvancedStatefulPodControl,
	statefulSet *apis.AdvancedStatefulSet) *DefaultScaleDownOperation {
	return &DefaultScaleDownOperation{
		PodControl:  podControl,
		StatefulSet: statefulSet,
	}
}

type ScaleDownOperation interface {
	// ExecuteScaleDown performs the scale down operation according to the given strategy
	ExecuteScaleDown(set *apis.AdvancedStatefulSet, strategy ScaleDownStrategy) (*ScaleDownStatus, error)
}

func (o *DefaultScaleDownOperation) ExecuteScaleDown(set *apis.AdvancedStatefulSet, strategy ScaleDownStrategy) (ScaleDownStatus, error) {
	r := StatefulSetReconciler{}
	// 假设已经有一个 AdvancedStatefulSet 对象 set
	// 调用 ShrinkPod 方法
	PodScaleInNum, err := r.ScaleInPodFunc(set, 5)
	if PodScaleInNum == -1 && err != nil {
		log.Fatalf("Error shrinking pods: %v", err)
	}
	status := ScaleDownStatus{CurReplicas: 5, ScaledDown: PodScaleInNum}
	return status, nil
}
