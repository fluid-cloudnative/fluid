/*
  Copyright 2022 The Fluid Authors.

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

package thin

import (
	"context"
	"fmt"
	"reflect"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"

	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

func (t ThinEngine) CheckRuntimeHealthy() (err error) {
	if t.isWorkerEnable() {
		// 1. Check the healthy of the workers
		err = t.checkWorkersHealthy()
		if err != nil {
			t.Log.Error(err, "The workers are not healthy")
			updateErr := t.UpdateDatasetStatus(data.FailedDatasetPhase)
			if updateErr != nil {
				t.Log.Error(updateErr, "Failed to update dataset")
			}
			return
		}

		updateErr := t.UpdateDatasetStatus(data.BoundDatasetPhase)
		if updateErr != nil {
			t.Log.Error(updateErr, "Failed to update dataset")
		}
	}

	// Check the healthy of the fuse
	err = t.checkFuseHealthy()
	if err != nil {
		t.Log.Error(err, "checkFuseHealthy failed")
		return
	}

	return
}

// checkWorkersHealthy check workers number changed
func (t *ThinEngine) checkWorkersHealthy() (err error) {
	workerName := t.getWorkerName()

	// Check the status of workers
	workers, err := kubeclient.GetStatefulSet(t.Client, workerName, t.namespace)
	if err != nil {
		return err
	}

	healthy := false
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {

		runtime, err := t.getRuntime()
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

			t.Log.Error(err, "the workers are not ready")
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
			updateErr := t.Client.Status().Update(context.TODO(), runtimeToUpdate)
			if updateErr != nil {
				return updateErr
			}
		}

		return err
	})

	if err != nil {
		t.Log.Error(err, "Failed update runtime")
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
func (t *ThinEngine) checkFuseHealthy() error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		runtime, err := t.getRuntime()
		if err != nil {
			t.Log.Error(err, "Failed to get Runtime", "runtimeName", t.name, "runtimeNamespace", t.namespace)
			return
		}
		err = t.Helper.CheckFuseHealthy(t.Recorder, runtime.DeepCopy(), t.getFuseName())
		if err != nil {
			t.Log.Error(err, "Failed to check runtimeFuse healthy")
		}
		return
	})
}
