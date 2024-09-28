package statefulset

import apps "github.com/fluid-cloudnative/fluid/pkg/types/cacheworkerset/client/v1"

type ScaleDownStrategy interface {
	// AllowCreateDuringDeletion returns true if Pods can be created during deletion
	AllowCreateDuringDeletion() bool
	// ShouldWaitForRunningAndReady returns true if Pods should wait for running and ready before deletion
	ShouldWaitForRunningAndReady() bool
}

type ScaleDownStatus struct {
	// TotalReplicas is the total number of replicas in the StatefulSet
	TotalReplicas int
	// ScaledDown is the number of replicas that were scaled down
	ScaledDown int
	// Errors contains any errors that occurred during the scale down operation
	Errors []error
}

func (s *ScaleDownStatus) AddError(err error) {
	s.Errors = append(s.Errors, err)
}

var _ ScaleDownOperation = &DefaultScaleDownOperation{}

type DefaultScaleDownOperation struct {
	StatefulSetControl StatefulSetControlInterface
	PodControl         PodControlInterface
	StatefulSet        *apps.StatefulSet
}

func NewDefaultScaleDownOperation(
	statefulSetControl StatefulSetControlInterface,
	podControl PodControlInterface,
	statefulSet *apps.StatefulSet) *DefaultScaleDownOperation {
	return &DefaultScaleDownOperation{
		StatefulSetControl: statefulSetControl,
		PodControl:         podControl,
		StatefulSet:        statefulSet,
	}
}
func (o *DefaultScaleDownOperation) ExecuteScaleDown(strategy ScaleDownStrategy) (*ScaleDownStatus, error) {
}

type ScaleDownOperation interface {
	// ExecuteScaleDown performs the scale down operation according to the given strategy
	ExecuteScaleDown(strategy ScaleDownStrategy) (*ScaleDownStatus, error)
}

type DefaultScaleDownOperation struct {
	// StatefulSetControlInterface is the control logic for StatefulSets
	StatefulSetControl StatefulSetControlInterface
	// PodControlInterface is used to create, update, and delete Pods
	PodControl PodControlInterface
	// StatefulSet is the target of the scale down operation
	StatefulSet *apps.StatefulSet
}

func (o *DefaultScaleDownOperation) ExecuteScaleDown(strategy ScaleDownStrategy) (*ScaleDownStatus, error) {
	status := &ScaleDownStatus{
		TotalReplicas: int(*o.StatefulSet.Spec.Replicas),
	}

	// Logic to execute scale down operation according to the strategy
	// This may involve:
	// - Waiting for Pods to be running and ready
	// - Deleting Pods
	// - Handling errors
	// - Updating status

	// Example pseudocode for waiting and deleting
	pods, err := o.StatefulSetControl.ListPods(o.StatefulSet)
	if err != nil {
		return nil, err
	}

	for _, pod := range pods {
		if strategy.ShouldWaitForRunningAndReady() {
			if err := o.PodControl.WaitForRunningAndReady(pod); err != nil {
				status.AddError(err)
				continue
			}
		}

		if !strategy.AllowCreateDuringDeletion() {
			if err := o.PodControl.DeletePod(pod); err != nil {
				status.AddError(err)
			} else {
				status.UpdateScaledDown(1)
			}
		}
	}

	return status, nil
}
