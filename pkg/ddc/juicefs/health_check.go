/*
Copyright 2021 The Fluid Authors.

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

package juicefs

import (
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"reflect"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"

	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

func (j *JuiceFSEngine) CheckRuntimeHealthy() (err error) {
	// 1. Check the healthy of the workers
	err = j.checkWorkersHealthy()
	if err != nil {
		j.Log.Error(err, "The workers are not healthy")
		updateErr := j.UpdateDatasetStatus(data.FailedDatasetPhase)
		if updateErr != nil {
			j.Log.Error(updateErr, "Failed to update dataset")
		}
		return
	}

	// 2. Check the healthy of the fuse
	err = j.checkFuseHealthy()
	if err != nil {
		j.Log.Error(err, "The fuse is not healthy")
		updateErr := j.UpdateDatasetStatus(data.FailedDatasetPhase)
		if updateErr != nil {
			j.Log.Error(updateErr, "Failed to update dataset")
		}
		return
	}

	updateErr := j.UpdateDatasetStatus(data.BoundDatasetPhase)
	if updateErr != nil {
		j.Log.Error(updateErr, "Failed to update dataset")
	}

	return
}

// checkWorkersHealthy check workers number changed
func (j *JuiceFSEngine) checkWorkersHealthy() (err error) {
	workerName := j.getWorkerName()

	// Check the status of workers
	workers, err := kubeclient.GetStatefulSet(j.Client, workerName, j.namespace)
	if err != nil {
		return err
	}

	healthy := false
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {

		runtime, err := j.getRuntime()
		if err != nil {
			return err
		}

		runtimeToUpdate := runtime.DeepCopy()
		if workers.Status.ReadyReplicas == 0 && *workers.Spec.Replicas > 0 {
			if len(runtimeToUpdate.Status.Conditions) == 0 {
				runtimeToUpdate.Status.Conditions = []data.RuntimeCondition{}
			}
			cond := utils.NewRuntimeCondition(data.RuntimeWorkersReady, "The workers are not ready.",
				fmt.Sprintf("The statefulset %s in %s are not ready, the Unavailable number is %d, please fix it.",
					workers.Name,
					workers.Namespace,
					*workers.Spec.Replicas-workers.Status.ReadyReplicas), v1.ConditionFalse)

			_, oldCond := utils.GetRuntimeCondition(runtimeToUpdate.Status.Conditions, cond.Type)

			if oldCond == nil || oldCond.Type != cond.Type {
				runtimeToUpdate.Status.Conditions =
					utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
						cond)
			}

			runtimeToUpdate.Status.WorkerPhase = data.RuntimePhaseNotReady

			j.Log.Error(err, "the workers are not ready")
		} else {
			healthy = true
			cond := utils.NewRuntimeCondition(data.RuntimeWorkersReady, "The workers are ready.",
				"The workers are ready", v1.ConditionTrue)

			_, oldCond := utils.GetRuntimeCondition(runtimeToUpdate.Status.Conditions, cond.Type)

			if oldCond == nil || oldCond.Type != cond.Type {
				runtimeToUpdate.Status.Conditions =
					utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
						cond)
			}
		}
		runtimeToUpdate.Status.WorkerNumberReady = workers.Status.ReadyReplicas
		runtimeToUpdate.Status.WorkerNumberAvailable = workers.Status.CurrentReplicas
		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			updateErr := j.Client.Status().Update(context.TODO(), runtimeToUpdate)
			if updateErr != nil {
				return updateErr
			}
		}

		return err
	})

	if err != nil {
		j.Log.Error(err, "Failed update runtime")
		return err
	}

	if !healthy {
		err = fmt.Errorf("the workers %s in %s are not ready, the unhealthy number %d",
			workers.Name,
			workers.Namespace,
			*workers.Spec.Replicas-workers.Status.ReadyReplicas)
	}

	return err
}

// checkFuseHealthy check fuses number changed
func (j *JuiceFSEngine) checkFuseHealthy() (err error) {
	fuseName := j.getFuseDaemonsetName()

	fuses, err := j.getDaemonset(fuseName, j.namespace)
	if err != nil {
		return err
	}

	healthy := false
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {

		runtime, err := j.getRuntime()
		if err != nil {
			return err
		}

		runtimeToUpdate := runtime.DeepCopy()

		if fuses.Status.NumberUnavailable > 0 ||
			(fuses.Status.DesiredNumberScheduled > 0 && fuses.Status.NumberAvailable == 0) {
			if len(runtimeToUpdate.Status.Conditions) == 0 {
				runtimeToUpdate.Status.Conditions = []data.RuntimeCondition{}
			}
			cond := utils.NewRuntimeCondition(data.RuntimeFusesReady, "The Fuses are not ready.",
				fmt.Sprintf("The daemonset %s in %s are not ready, the unhealthy number %d",
					fuses.Name,
					fuses.Namespace,
					fuses.Status.UpdatedNumberScheduled), v1.ConditionFalse)
			_, oldCond := utils.GetRuntimeCondition(runtimeToUpdate.Status.Conditions, cond.Type)

			if oldCond == nil || oldCond.Type != cond.Type {
				runtimeToUpdate.Status.Conditions =
					utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
						cond)
			}

			runtimeToUpdate.Status.FusePhase = data.RuntimePhaseNotReady
			j.Log.Error(err, "Failed to check the fuse healthy")
		} else {
			healthy = true
			runtimeToUpdate.Status.FusePhase = data.RuntimePhaseReady
			cond := utils.NewRuntimeCondition(data.RuntimeFusesReady, "The Fuses are ready.",
				"The fuses are ready", v1.ConditionFalse)
			_, oldCond := utils.GetRuntimeCondition(runtimeToUpdate.Status.Conditions, cond.Type)

			if oldCond == nil || oldCond.Type != cond.Type {
				runtimeToUpdate.Status.Conditions =
					utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
						cond)
			}
		}

		runtimeToUpdate.Status.FuseNumberReady = int32(fuses.Status.NumberReady)
		runtimeToUpdate.Status.FuseNumberAvailable = int32(fuses.Status.NumberAvailable)
		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			updateErr := j.Client.Status().Update(context.TODO(), runtimeToUpdate)
			if updateErr != nil {
				return updateErr
			}
		}

		return err
	})

	if err != nil {
		j.Log.Error(err, "Failed update runtime")
		return err
	}

	if !healthy {
		err = fmt.Errorf("the daemonset %s in %s are not ready, the unhealthy number %d",
			fuses.Name,
			fuses.Namespace,
			fuses.Status.UpdatedNumberScheduled)
	}
	return err
}

func (j JuiceFSEngine) CheckRuntimeReady() (ready bool) {
	// todo
	return
}
