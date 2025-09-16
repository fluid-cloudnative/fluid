package engine

import (
	"context"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

func (e *CacheEngine) SetupWorkerComponent(workerValue *common.CacheRuntimeComponentValue) (bool, error) {
	shouldSetupWorker, err := e.ShouldSetupWorker()
	if err != nil {
		return false, err
	}
	if shouldSetupWorker {
		if err = e.SetupWorkerInternal(workerValue); err != nil {
			e.Log.Error(err, "failed to setup worker")
			return false, err
		}
	}

	return true, nil
}
func (e *CacheEngine) ShouldSetupWorker() (bool, error) {
	runtime, err := e.getRuntime()
	if err != nil {
		return false, err
	}

	if runtime.Status.Worker.Phase == datav1alpha1.RuntimePhaseNone {
		return true, nil
	}

	return false, nil
}

func (e *CacheEngine) SetupWorkerInternal(workerValue *common.CacheRuntimeComponentValue) error {
	// 1. reconcile to create worker workload
	if err := e.masterHelper.Reconciler(context.TODO(), workerValue); err != nil {
		return err
	}

	// 2. Update the status of the runtime
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		workerStatus, err := e.workerHelper.ConstructComponentStatus(context.TODO(), workerValue)
		if err != nil {
			return err
		}

		workerStatus.Phase = datav1alpha1.RuntimePhaseNotReady
		runtimeToUpdate := runtime.DeepCopy()
		runtimeToUpdate.Status.Worker = workerStatus
		if runtime.Status.Worker.Phase == datav1alpha1.RuntimePhaseNone {
			if len(runtimeToUpdate.Status.Conditions) == 0 {
				runtimeToUpdate.Status.Conditions = []datav1alpha1.RuntimeCondition{}
			}
			cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeMasterInitialized, datav1alpha1.RuntimeMasterInitializedReason,
				"The worker is initialized.", corev1.ConditionTrue)
			runtimeToUpdate.Status.Conditions =
				utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
					cond)
		}

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			return e.Client.Status().Update(context.TODO(), runtimeToUpdate)
		}

		return nil
	}); err != nil {
		e.Log.Error(err, "Update runtime status")
		return err
	}

	return nil
}

func (e *CacheEngine) CheckWorkerReady(workerValue *common.CacheRuntimeComponentValue) (bool, error) {
	runtime, err := e.getRuntime()
	if err != nil {
		return false, err
	}
	runtimeToUpdate := runtime.DeepCopy()

	exist, err := e.workerHelper.CheckComponentExist(context.TODO(), workerValue)
	if err != nil || !exist {
		return false, err
	}

	workerStatus, err := e.workerHelper.ConstructComponentStatus(context.TODO(), workerValue)
	if err != nil {
		return false, err
	}

	phase := datav1alpha1.RuntimePhasePartialReady
	cond := datav1alpha1.RuntimeCondition{}

	if workerStatus.ReadyReplicas == workerValue.Replicas {
		phase = datav1alpha1.RuntimePhaseReady
	}

	ready := false
	switch phase {
	case datav1alpha1.RuntimePhaseReady, datav1alpha1.RuntimePhasePartialReady:
		ready = true
	default:
		e.Log.Info("workers are not ready", "phase", phase)
	}
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		if phase != runtimeToUpdate.Status.Worker.Phase {
			statusToUpdate := runtimeToUpdate.Status.DeepCopy()
			statusToUpdate.Worker.Phase = phase

			if len(statusToUpdate.Conditions) == 0 {
				statusToUpdate.Conditions = []datav1alpha1.RuntimeCondition{}
			}

			switch phase {
			case datav1alpha1.RuntimePhaseReady:
				cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, datav1alpha1.RuntimeWorkersReadyReason,
					"The workers are ready.", corev1.ConditionTrue)
			case datav1alpha1.RuntimePhasePartialReady:
				cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, datav1alpha1.RuntimeWorkersReadyReason,
					"The workers are partially ready.", corev1.ConditionTrue)
			case datav1alpha1.RuntimePhaseNotReady:
				cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, datav1alpha1.RuntimeWorkersReadyReason,
					"The workers are not ready.", corev1.ConditionFalse)
			}

			if cond.Type != "" {
				statusToUpdate.Conditions =
					utils.UpdateRuntimeCondition(statusToUpdate.Conditions,
						cond)
			}

			if !reflect.DeepEqual(runtimeToUpdate.Status, statusToUpdate) {
				return e.Client.Status().Update(context.TODO(), runtime)
			}

		} else {
			e.Log.V(1).Info("No need to update runtime status for checking healthy")
		}
		return nil

	}); err != nil {
		e.Log.Error(err, "failed to update runtime status")
		return false, err
	}

	return ready, nil
}
