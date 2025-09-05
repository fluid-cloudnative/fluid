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
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/types"

	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

func (j *JuiceFSEngine) CheckRuntimeHealthy() (err error) {
	// 1. Check the healthy of the workers
	workerReady, err := j.CheckWorkersReady()
	if err != nil {
		j.Log.Error(err, "failed to check if worker is ready")
		updateErr := j.UpdateDatasetStatus(data.FailedDatasetPhase)
		if updateErr != nil {
			j.Log.Error(updateErr, "failed to update dataset status to \"Failed\"")
		}
		return
	}

	if !workerReady {
		return fmt.Errorf("the worker \"%s\" is not healthy, expect at least one replica is ready", j.getWorkerName())
	}

	// 2. Check the healthy of the fuse
	fuseReady, err := j.checkFuseHealthy()
	if err != nil {
		j.Log.Error(err, "failed to check if fuse is healthy")
		updateErr := j.UpdateDatasetStatus(data.FailedDatasetPhase)
		if updateErr != nil {
			j.Log.Error(updateErr, "failed to update dataset status to \"Failed\"")
		}
		return
	}

	if !fuseReady {
		// fluid assumes fuse is always ready, so it's a protective branch.
		return fmt.Errorf("the fuse \"%s\" is not healthy", j.getFuseName())
	}

	err = j.UpdateDatasetStatus(data.BoundDatasetPhase)
	if err != nil {
		j.Log.Error(err, "failed to update dataset status to \"Bound\"")
		return
	}

	return
}

// checkFuseHealthy check fuses number changed
func (j *JuiceFSEngine) checkFuseHealthy() (ready bool, err error) {
	getRuntimeFn := func(client client.Client) (base.RuntimeInterface, error) {
		return utils.GetJuiceFSRuntime(client, j.name, j.namespace)
	}

	ready, err = j.Helper.CheckAndSyncFuseStatus(getRuntimeFn, types.NamespacedName{Namespace: j.namespace, Name: j.getFuseName()})
	if err != nil {
		j.Log.Error(err, "failed to check and update fuse status")
		return
	}

	if !ready {
		j.Log.Info("fuses are not ready")
	}

	return
}
