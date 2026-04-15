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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/cache/component"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"reflect"
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
	manager := component.NewComponentHelper(clientValue.WorkloadType, e.Scheme, e.Client)
	err := manager.Reconciler(context.TODO(), clientValue)
	if err != nil {
		return err
	}

	// 2. Update the status of the runtime
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		clientStatus, err := manager.ConstructComponentStatus(context.TODO(), clientValue)
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
			cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeFusesInitialized, datav1alpha1.RuntimeFusesInitializedReason,
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
