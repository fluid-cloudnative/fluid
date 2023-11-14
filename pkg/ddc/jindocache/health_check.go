/*
Copyright 2023 The Fluid Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package jindocache

import (
	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

func (e *JindoCacheEngine) CheckRuntimeHealthy() (err error) {

	// 1. Check the healthy of the master
	if !e.runtime.Spec.Master.Disabled {
		err = e.checkMasterHealthy()
		if err != nil {
			e.Log.Error(err, "The master is not healthy")
			updateErr := e.UpdateDatasetStatus(data.FailedDatasetPhase)
			if updateErr != nil {
				e.Log.Error(updateErr, "Failed to update dataset")
			}
			return
		}
	}

	// 2. Check the healthy of the workers
	if !e.runtime.Spec.Worker.Disabled {
		err = e.checkWorkersHealthy()
		if err != nil {
			e.Log.Error(err, "The worker is not healthy")
			updateErr := e.UpdateDatasetStatus(data.FailedDatasetPhase)
			if updateErr != nil {
				e.Log.Error(updateErr, "Failed to update dataset")
			}
			return
		}
	}

	// 3. Check the healthy of the fuse
	if !e.runtime.Spec.Fuse.Disabled {
		err = e.checkFuseHealthy()
		if err != nil {
			e.Log.Error(err, "The fuse is not healthy")
			updateErr := e.UpdateDatasetStatus(data.FailedDatasetPhase)
			if updateErr != nil {
				e.Log.Error(updateErr, "Failed to update dataset")
			}
			return
		}
	}

	// 4. Update the dataset as Bounded
	return e.UpdateDatasetStatus(data.BoundDatasetPhase)
}

// checkMasterHealthy checks the master healthy
func (e *JindoCacheEngine) checkMasterHealthy() (err error) {
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
func (e *JindoCacheEngine) checkWorkersHealthy() (err error) {
	workers, err := ctrl.GetWorkersAsStatefulset(e.Client,
		types.NamespacedName{Namespace: e.namespace, Name: e.getWorkerName()})
	if err != nil {
		if fluiderrs.IsDeprecated(err) {
			e.Log.Info("Warning: the current runtime is created by runtime controller before v0.7.0, checking worker health state is not supported. To support these features, please create a new dataset", "details", err)
			e.Recorder.Event(e.runtime, corev1.EventTypeWarning, common.RuntimeDeprecated, "The runtime is created by controllers before v0.7.0, to fully enable latest capabilities, please delete the runtime and create a new one")
			return nil
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
func (e *JindoCacheEngine) checkFuseHealthy() (err error) {
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
