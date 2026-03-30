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
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"reflect"
	"time"
)

func (e *CacheEngine) setMasterComponentStatus(status *fluidapi.RuntimeComponentStatus) (ready bool, err error) {
	return true, newNotImplementError("setMasterComponentStatus")
}
func (e *CacheEngine) setWorkerComponentStatus(status *fluidapi.RuntimeComponentStatus) (ready bool, err error) {
	return true, newNotImplementError("setWorkerComponentStatus")
}
func (e *CacheEngine) setClientComponentStatus(status *fluidapi.RuntimeComponentStatus) (err error) {
	return newNotImplementError("setClientComponentStatus")
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
			masterReady, err = e.setMasterComponentStatus(&runtimeToUpdate.Status.Master)
			if err != nil {
				return err
			}
		}

		if value.Worker.Enabled {
			workerReady, err = e.setWorkerComponentStatus(&runtimeToUpdate.Status.Worker)
			if err != nil {
				return err
			}
		}

		if value.Client.Enabled {
			err = e.setClientComponentStatus(&runtimeToUpdate.Status.Client)
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
