package jindo

import (
	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

func (e *JindoEngine) CheckRuntimeHealthy() (err error) {
	// 1. Check the healthy of the master
	err = e.checkMasterHealthy()
	if err != nil {
		e.Log.Error(err, "The master is not healthy")
		updateErr := e.UpdateDatasetStatus(data.FailedDatasetPhase)
		if updateErr != nil {
			e.Log.Error(updateErr, "Failed to update dataset")
		}
		return
	}

	// 2. Check the healthy of the workers
	err = e.checkWorkersHealthy()
	if err != nil {
		e.Log.Error(err, "The worker is not healthy")
		updateErr := e.UpdateDatasetStatus(data.FailedDatasetPhase)
		if updateErr != nil {
			e.Log.Error(updateErr, "Failed to update dataset")
		}
		return
	}

	// 3. Check the healthy of the fuse
	err = e.checkFuseHealthy()
	if err != nil {
		e.Log.Error(err, "The fuse is not healthy")
		updateErr := e.UpdateDatasetStatus(data.FailedDatasetPhase)
		if updateErr != nil {
			e.Log.Error(updateErr, "Failed to update dataset")
		}
		return
	}

	// 4. Update the dataset as Bounded
	return e.UpdateDatasetStatus(data.BoundDatasetPhase)
}

// checkMasterHealthy checks the master healthy
func (e *JindoEngine) checkMasterHealthy() (err error) {
	master, err := kubeclient.GetStatefulSet(e.Client, e.getMasterName(), e.namespace)
	if err != nil {
		return err
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		runtime, err := e.getRuntime()
		if err != nil {
			return
		}
		runtimeToUpdate := runtime.DeepCopy()
		err = e.Helper.CheckMasterHealthy(e.Recorder, runtimeToUpdate, runtimeToUpdate.Status, master)
		if err != nil {
			e.Log.Error(err, "Failed to check master healthy")
		}
		return
	})

	if err != nil {
		e.Log.Error(err, "Failed to check master healthy")
	}

	return
}

// checkWorkerHealthy checks the Worker healthy
func (e *JindoEngine) checkWorkersHealthy() (err error) {
	workers, err := ctrl.GetWorkersAsStatefulset(e.Client,
		types.NamespacedName{Namespace: e.namespace, Name: e.getWorkerName()})
	if err != nil {
		if fluiderrs.IsDeprecated(err) {
			e.Log.Info("Warning: Deprecated mode is not support, so skip handling", "details", err)
			return
		}
		return
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		runtime, err := e.getRuntime()
		if err != nil {
			return
		}
		runtimeToUpdate := runtime.DeepCopy()
		err = e.Helper.CheckWorkersHealthy(e.Recorder, runtimeToUpdate, runtimeToUpdate.Status, workers)
		if err != nil {
			e.Log.Error(err, "Failed to check Worker healthy")
		}
		return
	})

	if err != nil {
		e.Log.Error(err, "Failed to check Worker healthy")
	}

	return
}

// checkFuseHealthy checks the Fuse healthy
func (e *JindoEngine) checkFuseHealthy() (err error) {
	Fuse, err := kubeclient.GetDaemonset(e.Client, e.getFuseName(), e.namespace)
	if err != nil {
		return err
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		runtime, err := e.getRuntime()
		if err != nil {
			return
		}
		runtimeToUpdate := runtime.DeepCopy()
		err = e.Helper.CheckFuseHealthy(e.Recorder, runtimeToUpdate, runtimeToUpdate.Status, Fuse)
		if err != nil {
			e.Log.Error(err, "Failed to check Fuse healthy")
		}
		return
	})

	if err != nil {
		e.Log.Error(err, "Failed to check Fuse healthy")
	}

	return
}
