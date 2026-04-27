/*
  Copyright 2026 The Fluid Authors.

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

package engine

import (
	"context"
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/cache/component"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
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
	manager := component.NewComponentHelper(workerValue.WorkloadType, e.Client)
	err := manager.Reconciler(context.TODO(), workerValue)
	if err != nil {
		return err
	}

	// update status of worker
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		workerStatus, err := manager.ConstructComponentStatus(context.TODO(), workerValue)
		if err != nil {
			return err
		}
		// from RuntimePhaseNone to RuntimePhaseNotReady, not reconcile the worker component the next time.
		workerStatus.Phase = datav1alpha1.RuntimePhaseNotReady

		// TODO: support builds workers affinity ? do it in transformer ?
		runtimeToUpdate := runtime.DeepCopy()
		runtimeToUpdate.Status.Worker = workerStatus

		// TODO(cache runtime): why need this line judgement ?
		if runtime.Status.Worker.Phase == datav1alpha1.RuntimePhaseNone {
			if len(runtimeToUpdate.Status.Conditions) == 0 {
				runtimeToUpdate.Status.Conditions = []datav1alpha1.RuntimeCondition{}
			}
			cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, datav1alpha1.RuntimeWorkersInitializedReason,
				"The worker is initialized.", corev1.ConditionTrue)
			runtimeToUpdate.Status.Conditions = utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions, cond)

		}
		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			return e.Client.Status().Update(context.TODO(), runtimeToUpdate)
		}

		return nil
	})
	if err != nil {
		e.Log.Error(err, "failed to update runtime status")
		return err
	}

	return nil
}
