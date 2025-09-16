package engine

import (
	"context"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

func (e *CacheEngine) SetupClientComponent(clientValue *common.CacheRuntimeComponentValue) (bool, error) {
	shouldSetupClient, err := e.ShouldSetupClient()
	if err != nil {
		return false, err
	}
	if shouldSetupClient {
		if err = e.SetupClientInternal(clientValue); err != nil {
			e.Log.Error(err, "failed to setup client")
			return false, err
		}
	}

	return true, nil
}

func (e *CacheEngine) ShouldSetupClient() (bool, error) {
	runtime, err := e.getRuntime()
	if err != nil {
		return false, err
	}

	if runtime.Status.Client.Phase == datav1alpha1.RuntimePhaseNone {
		return true, nil
	}

	return false, nil
}

func (e *CacheEngine) SetupClientInternal(clientValue *common.CacheRuntimeComponentValue) error {
	// 1. reconcile to create client workload
	if err := e.clientHelper.Reconciler(context.TODO(), clientValue); err != nil {
		return err
	}

	// 2. Update the status of the runtime
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		clientStatus, err := e.clientHelper.ConstructComponentStatus(context.TODO(), clientValue)
		if err != nil {
			return err
		}
		clientStatus.Phase = datav1alpha1.RuntimePhaseNotReady

		runtimeToUpdate := runtime.DeepCopy()
		runtimeToUpdate.Status.Client = clientStatus
		if runtime.Status.Client.Phase == datav1alpha1.RuntimePhaseNone && clientStatus.Phase != datav1alpha1.RuntimePhaseNone {
			if len(runtimeToUpdate.Status.Conditions) == 0 {
				runtimeToUpdate.Status.Conditions = []datav1alpha1.RuntimeCondition{}
			}
			cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeMasterInitialized, datav1alpha1.RuntimeMasterInitializedReason,
				"The client setup finished.", corev1.ConditionTrue)
			runtimeToUpdate.Status.Conditions =
				utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
					cond)
		}

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			return e.Client.Status().Update(context.TODO(), runtimeToUpdate)
		}

		return nil
	}); err != nil {
		e.Log.Error(err, "update runtime status")
		return err
	}

	return nil
}

func (e *CacheEngine) CheckClientReady(clientValue *common.CacheRuntimeComponentValue) (bool, error) {
	clientStatus, err := e.clientHelper.ConstructComponentStatus(context.TODO(), clientValue)
	if err != nil {
		return false, err
	}

	ready := false
	if clientStatus.Phase != datav1alpha1.RuntimePhaseReady {
		e.Log.Info("The client is not ready.", "replicas", clientValue.Replicas,
			"readyReplicas", clientStatus.ReadyReplicas)
	} else {
		e.Log.Info("The client is ready.", "replicas", clientValue.Replicas,
			"readyReplicas", clientStatus.ReadyReplicas)
		ready = true
	}

	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}
		runtimeToUpdate := runtime.DeepCopy()
		runtimeToUpdate.Status.Worker = clientStatus
		if len(runtimeToUpdate.Status.Conditions) == 0 {
			runtimeToUpdate.Status.Conditions = []datav1alpha1.RuntimeCondition{}
		}
		if ready {
			cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, datav1alpha1.RuntimeWorkersReadyReason,
				"The worker is ready.", corev1.ConditionTrue)
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
		return false, err
	}

	return ready, nil
}
