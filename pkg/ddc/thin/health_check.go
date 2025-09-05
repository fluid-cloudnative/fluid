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
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

func (t ThinEngine) CheckRuntimeHealthy() (err error) {
	if t.isWorkerEnable() {
		// 1. Check the healthy of the workers
		var workerReady bool
		workerReady, err = t.CheckWorkersReady()
		if err != nil {
			t.Log.Error(err, "fail to check if worker is ready")
			updateErr := t.UpdateDatasetStatus(data.FailedDatasetPhase)
			if updateErr != nil {
				t.Log.Error(updateErr, "fail to update dataset status to \"Failed\"")
			}
			return
		}

		if !workerReady {
			return fmt.Errorf("the worker \"%s\" is not healthy, expect at least one replica is ready", t.getWorkerName())
		}
	}

	// Check the healthy of the fuse
	var fuseReady bool
	fuseReady, err = t.checkFuseHealthy()
	if err != nil {
		t.Log.Error(err, "fail to check if fuse is ready")
		return
	}

	if !fuseReady {
		return fmt.Errorf("the fuse \"%s\" is not healthy", t.getFuseName())
	}

	err = t.UpdateDatasetStatus(data.BoundDatasetPhase)
	if err != nil {
		t.Log.Error(err, "fail to update dataset status to \"Bound\"")
	}

	return
}

// checkFuseHealthy check fuses number changed
func (t *ThinEngine) checkFuseHealthy() (ready bool, err error) {
	getRuntimeFn := func(client client.Client) (base.RuntimeInterface, error) {
		return utils.GetThinRuntime(client, t.name, t.namespace)
	}

	ready, err = t.Helper.CheckAndSyncFuseStatus(getRuntimeFn, types.NamespacedName{Namespace: t.namespace, Name: t.getFuseName()})
	if err != nil {
		t.Log.Error(err, "fail to check and update fuse status")
		return
	}

	if !ready {
		t.Log.Info("fuses are not ready")
	}

	return
}
