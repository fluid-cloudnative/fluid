package engine

import (
	"context"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
)

func (e *CacheEngine) SetupMasterComponent(masterValue *common.CacheRuntimeComponentValue) (bool, error) {
	shouldSetupMaster, err := e.ShouldSetupMaster(masterValue)
	if err != nil {
		return false, err
	}
	if shouldSetupMaster {
		if err = e.SetupMasterInternal(masterValue); err != nil {
			e.Log.Error(err, "failed to setup master")
			return false, err
		}
	}

	return true, nil
}

func (e *CacheEngine) ShouldSetupMaster(masterValue *common.CacheRuntimeComponentValue) (bool, error) {
	runtime, err := e.getRuntime()
	if err != nil {
		return false, err
	}

	if runtime.Status.Master.Phase == datav1alpha1.RuntimePhaseNone {
		return true, nil
	}

	return false, nil
}

func (e *CacheEngine) SetupMasterInternal(masterValue *common.CacheRuntimeComponentValue) error {
	// 1. reconcile to create master workload
	if err := e.masterHelper.Reconciler(context.TODO(), masterValue); err != nil {
		return err
	}

	// 2. Update the status of the runtime
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		masterStatus, err := e.masterHelper.ConstructComponentStatus(context.TODO(), masterValue)
		if err != nil {
			return err
		}
		masterStatus.Phase = datav1alpha1.RuntimePhaseNotReady
		runtimeToUpdate := runtime.DeepCopy()
		runtimeToUpdate.Status.Master = masterStatus

		if len(runtimeToUpdate.Status.Conditions) == 0 {
			runtimeToUpdate.Status.Conditions = []datav1alpha1.RuntimeCondition{}
		}
		cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeMasterInitialized, datav1alpha1.RuntimeMasterInitializedReason,
			"The master is initialized.", corev1.ConditionTrue)
		runtimeToUpdate.Status.Conditions =
			utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
				cond)

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			return e.Client.Status().Update(context.TODO(), runtimeToUpdate)
		}

		return nil
	}); err != nil {
		e.Log.Error(err, "failed to update runtime status")
		return err
	}

	return nil
}

func (e *CacheEngine) CheckMasterReady(masterValue *common.CacheRuntimeComponentValue) (bool, error) {
	exist, err := e.masterHelper.CheckComponentExist(context.TODO(), masterValue)
	if err != nil || !exist {
		return exist, err
	}

	masterStatus, err := e.masterHelper.ConstructComponentStatus(context.TODO(), masterValue)
	if err != nil {
		return false, err
	}

	// TODO: support checkReadyPolicy, vendor can define if partialReady can be tolerated in runtimeClass
	if masterStatus.ReadyReplicas != masterValue.Replicas {
		e.Log.Info("The master is not ready", "masterName", masterValue.Name,
			"masterWorkloadType", masterValue.WorkloadType)

		return false, nil
	}
	masterStatus.Phase = datav1alpha1.RuntimePhaseReady
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}
		runtimeToUpdate := runtime.DeepCopy()

		if len(runtimeToUpdate.Status.Conditions) == 0 {
			runtimeToUpdate.Status.Conditions = []datav1alpha1.RuntimeCondition{}
		}

		cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeMasterReady, datav1alpha1.RuntimeMasterReadyReason,
			"The master is ready.", corev1.ConditionTrue)
		runtimeToUpdate.Status.Conditions =
			utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
				cond)
		runtimeToUpdate.Status.Master = masterStatus

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			return e.Client.Status().Update(context.TODO(), runtimeToUpdate)
		}
		return nil
	}); err != nil {
		e.Log.Error(err, "Update runtime status")
		return false, err
	}

	return true, nil
}
