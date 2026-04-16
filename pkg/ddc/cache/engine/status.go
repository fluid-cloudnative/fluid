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
	"fmt"
	fluidapi "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/cache/component"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"reflect"
	"time"
)

func (e *CacheEngine) setMasterComponentStatus(componentValue *common.CacheRuntimeComponentValue, status *fluidapi.CacheRuntimeStatus) (ready bool, err error) {
	manager := component.NewComponentHelper(componentValue.WorkloadType, e.Scheme, e.Client)

	masterStatus, err := manager.ConstructComponentStatus(context.TODO(), componentValue)
	if err != nil {
		return false, err
	}
	if masterStatus.ReadyReplicas == masterStatus.DesiredReplicas {
		masterStatus.Phase = fluidapi.RuntimePhaseReady
		ready = true
	} else {
		masterStatus.Phase = fluidapi.RuntimePhaseNotReady
	}
	status.Master = masterStatus

	return ready, err
}
func (e *CacheEngine) setWorkerComponentStatus(componentValue *common.CacheRuntimeComponentValue, status *fluidapi.CacheRuntimeStatus) (ready bool, err error) {
	manager := component.NewComponentHelper(componentValue.WorkloadType, e.Scheme, e.Client)

	workerStatus, err := manager.ConstructComponentStatus(context.TODO(), componentValue)
	if err != nil {
		return false, err
	}

	if workerStatus.DesiredReplicas == 0 {
		workerStatus.Phase = fluidapi.RuntimePhaseReady
		ready = true
	} else if workerStatus.ReadyReplicas > 0 {
		if workerStatus.DesiredReplicas == workerStatus.ReadyReplicas {
			workerStatus.Phase = fluidapi.RuntimePhaseReady
			ready = true
		} else if workerStatus.ReadyReplicas >= 1 {
			workerStatus.Phase = fluidapi.RuntimePhasePartialReady
			ready = true
		}
	} else {
		workerStatus.Phase = fluidapi.RuntimePhaseNotReady
	}
	status.Worker = workerStatus

	return ready, err
}
func (e *CacheEngine) setClientComponentStatus(componentValue *common.CacheRuntimeComponentValue, status *fluidapi.CacheRuntimeStatus) (err error) {
	manager := component.NewComponentHelper(componentValue.WorkloadType, e.Scheme, e.Client)

	clientStatus, err := manager.ConstructComponentStatus(context.TODO(), componentValue)
	if err != nil {
		return err
	}
	if clientStatus.DesiredReplicas > 0 {
		if clientStatus.DesiredReplicas == clientStatus.ReadyReplicas {
			clientStatus.Phase = fluidapi.RuntimePhaseReady
		} else if clientStatus.ReadyReplicas >= 1 {
			clientStatus.Phase = fluidapi.RuntimePhasePartialReady
		}
	}
	status.Client = clientStatus

	return nil
}
func (e *CacheEngine) CheckAndUpdateRuntimeStatus(value *common.CacheRuntimeValue) (bool, error) {
	var masterReady, workerReady, runtimeReady = true, true, false

	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		runtimeToUpdate := runtime.DeepCopy()
		if value.Master.Enabled {
			masterReady, err = e.setMasterComponentStatus(value.Master, &runtimeToUpdate.Status)
			if err != nil {
				return err
			}
		}

		if value.Worker.Enabled {
			workerReady, err = e.setWorkerComponentStatus(value.Worker, &runtimeToUpdate.Status)
			if err != nil {
				return err
			}
		}

		if value.Client.Enabled {
			err = e.setClientComponentStatus(value.Client, &runtimeToUpdate.Status)
			if err != nil {
				return err
			}
		}

		if masterReady && workerReady {
			runtimeReady = true
		} else {
			e.Log.Info(fmt.Sprintf("MasterReady: %v, workerReady: %v", masterReady, workerReady))
		}

		// Update the setup time
		if runtimeReady && runtimeToUpdate.Status.SetupDuration == "" {
			runtimeToUpdate.Status.SetupDuration = utils.CalculateDuration(runtimeToUpdate.CreationTimestamp.Time, time.Now())
		}

		// TODO(cache runtime): set the CacheRuntime Status left fields，should add CacheStates field ?

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			err = e.Client.Status().Update(context.TODO(), runtimeToUpdate)
		} else {
			e.Log.Info("Do nothing because the runtime status is not changed.")
		}

		return err
	})

	if err != nil {
		_ = utils.LoggingErrorExceptConflict(e.Log, err, "Failed to update runtime status", types.NamespacedName{Namespace: e.namespace, Name: e.name})
		return false, err
	}

	return runtimeReady, nil
}
