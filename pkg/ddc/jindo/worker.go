package jindo

import (
	"context"
	"reflect"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
)

// SetupWorkers checks the desired and current replicas of workers and makes an update
// over the status by setting phases and conditions. The function
// calls for a status update and finally returns error if anything unexpected happens.
func (e *JindoEngine) SetupWorkers() (err error) {
	var (
		workerName        string = e.getWorkertName()
		namespace         string = e.namespace
		needRuntimeUpdate bool   = false
	)
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		workers, err := e.getStatefulset(workerName, namespace)
		if err != nil {
			return err
		}

		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		desireReplicas := runtime.Replicas()
		if *workers.Spec.Replicas != desireReplicas {
			workerToUpdate := workers.DeepCopy()
			workerToUpdate.Spec.Replicas = &desireReplicas
			err = e.Client.Update(context.TODO(), workerToUpdate)
			if err != nil {
				return err
			}
		} else {
			e.Log.V(1).Info("Nothing to do for syncing")
		}

		needRuntimeUpdate = true
		return nil
	})

	if err != nil {
		e.Log.Error(err, "Failed to setup worker")
		return err
	}

	if !needRuntimeUpdate {
		return nil
	}

	// 2. Update the runtime status
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			e.Log.Error(err, "setupWorker")
			return err
		}

		workers, err := e.getStatefulset(workerName, namespace)
		if err != nil {
			return err
		}

		runtimeToUpdate := runtime.DeepCopy()

		runtimeToUpdate.Status.WorkerPhase = datav1alpha1.RuntimePhaseNotReady
		runtimeToUpdate.Status.DesiredWorkerNumberScheduled = runtime.Replicas()
		runtimeToUpdate.Status.CurrentWorkerNumberScheduled = workers.Status.Replicas

		if len(runtimeToUpdate.Status.Conditions) == 0 {
			runtimeToUpdate.Status.Conditions = []datav1alpha1.RuntimeCondition{}
		}
		cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersInitialized, datav1alpha1.RuntimeWorkersInitializedReason,
			"The workers are initialized.", corev1.ConditionTrue)
		runtimeToUpdate.Status.Conditions =
			utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
				cond)

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			return e.Client.Status().Update(context.TODO(), runtimeToUpdate)
		}

		return nil
	})

	return
}

// ShouldSetupWorkers checks if we need setup the workers
func (e *JindoEngine) ShouldSetupWorkers() (should bool, err error) {
	runtime, err := e.getRuntime()
	if err != nil {
		return
	}

	switch runtime.Status.WorkerPhase {
	case datav1alpha1.RuntimePhaseNone:
		should = true
	default:
		should = false
	}

	return
}

// are the workers ready
func (e *JindoEngine) CheckWorkersReady() (ready bool, err error) {
	var (
		workerReady, workerPartialReady bool
		workerName                      string = e.getWorkertName()
		namespace                       string = e.namespace
	)

	runtime, err := e.getRuntime()
	if err != nil {
		return ready, err
	}

	workers, err := e.getStatefulset(workerName, namespace)
	if err != nil {
		return ready, err
	}

	if workers.Status.ReadyReplicas > 0 {
		if runtime.Replicas() == workers.Status.ReadyReplicas {
			workerReady = true
		} else if workers.Status.ReadyReplicas >= 1 {
			workerPartialReady = true
		}
	}

	if workerReady || workerPartialReady {
		ready = true
	} else {
		e.Log.Info("workers are not ready", "workerReady", workerReady,
			"workerPartialReady", workerPartialReady)
		return
	}

	// update the status as the workers are ready
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}
		runtimeToUpdate := runtime.DeepCopy()
		if len(runtimeToUpdate.Status.Conditions) == 0 {
			runtimeToUpdate.Status.Conditions = []datav1alpha1.RuntimeCondition{}
		}
		cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, datav1alpha1.RuntimeWorkersReadyReason,
			"The workers are ready.", corev1.ConditionTrue)
		if workerPartialReady {
			cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, datav1alpha1.RuntimeWorkersReadyReason,
				"The workers are partially ready.", corev1.ConditionTrue)

			runtimeToUpdate.Status.WorkerPhase = datav1alpha1.RuntimePhasePartialReady
		}
		runtimeToUpdate.Status.Conditions =
			utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
				cond)

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			return e.Client.Status().Update(context.TODO(), runtimeToUpdate)
		}

		return nil
	})

	return
}

// getWorkerSelectors gets the selector of the worker
func (e *JindoEngine) getWorkerSelectors() string {
	labels := map[string]string{
		"release":     e.name,
		POD_ROLE_TYPE: WOKRER_POD_ROLE,
		"app":         common.JINDO_RUNTIME,
	}
	labelSelector := &metav1.LabelSelector{
		MatchLabels: labels,
	}

	selectorValue := ""
	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		e.Log.Error(err, "Failed to parse the labelSelector of the runtime", "labels", labels)
	} else {
		selectorValue = selector.String()
	}
	return selectorValue
}
