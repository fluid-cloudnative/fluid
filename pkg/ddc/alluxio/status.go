/*

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

package alluxio

import (
	"context"
	"reflect"

	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"k8s.io/client-go/util/retry"
)

// CheckAndUpdateRuntimeStatus checks the related runtime status and update it.
func (e *AlluxioEngine) CheckAndUpdateRuntimeStatus() (ready bool, err error) {

	var (
		masterReady, workerReady, fuseReady bool
		// workerPartialReady, fusePartialReady bool
		masterName string = e.getMasterStatefulsetName()
		workerName string = e.getWorkerDaemonsetName()
		fuseName   string = e.getFuseDaemonsetName()
		namespace  string = e.namespace
	)

	// 1. Master should be ready
	master, err := e.getMasterStatefulset(masterName, namespace)
	if err != nil {
		return ready, err
	}

	// 2. Worker should be ready
	workers, err := e.getDaemonset(workerName, namespace)
	if err != nil {
		return ready, err
	}

	// 3. fuse shoulde be ready
	// runtimeToUpdate.Status.DesiredFuseNumberScheduled = int32(fuses.Status.DesiredNumberScheduled)
	fuses, err := e.getDaemonset(fuseName, namespace)
	if err != nil {
		return ready, err
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		runtimeToUpdate := runtime.DeepCopy()
		if reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			e.Log.V(1).Info("The runtime is equal after deepcopy")
		}

		states, err := e.queryCacheStatus()
		if err != nil {
			return err
		}

		// 0. Update the cache status
		// runtimeToUpdate.Status.CacheStates[data.Cacheable] = states.cacheable
		if len(runtime.Status.CacheStates) == 0 {
			runtimeToUpdate.Status.CacheStates = map[common.CacheStateName]string{}
		}

		runtimeToUpdate.Status.CacheStates[common.CacheCapacity] = states.cacheCapacity
		runtimeToUpdate.Status.CacheStates[common.CachedPercentage] = states.cachedPercentage
		runtimeToUpdate.Status.CacheStates[common.Cached] = states.cached

		runtimeToUpdate.Status.CurrentMasterNumberScheduled = int32(master.Status.Replicas)
		runtimeToUpdate.Status.MasterNumberReady = int32(master.Status.ReadyReplicas)

		if *master.Spec.Replicas == master.Status.ReadyReplicas {
			runtimeToUpdate.Status.MasterPhase = data.RuntimePhaseReady
			masterReady = true
		} else {
			runtimeToUpdate.Status.MasterPhase = data.RuntimePhaseNotReady
		}

		runtimeToUpdate.Status.WorkerNumberReady = int32(workers.Status.NumberReady)
		runtimeToUpdate.Status.WorkerNumberUnavailable = int32(workers.Status.NumberUnavailable)
		runtimeToUpdate.Status.WorkerNumberAvailable = int32(workers.Status.NumberAvailable)
		if runtime.Replicas() == workers.Status.NumberReady {
			runtimeToUpdate.Status.WorkerPhase = data.RuntimePhaseReady
			// runtimeToUpdate.Status.CacheStates[data.Cacheable] = runtime.Status.CacheStates[data.CacheCapacity]
			workerReady = true
		} else if workers.Status.NumberAvailable == workers.Status.NumberReady {
			runtimeToUpdate.Status.WorkerPhase = data.RuntimePhasePartialReady
			workerReady = true
		} else {
			runtimeToUpdate.Status.WorkerPhase = data.RuntimePhaseNotReady
		}

		runtimeToUpdate.Status.FuseNumberReady = int32(fuses.Status.NumberReady)
		runtimeToUpdate.Status.FuseNumberUnavailable = int32(fuses.Status.NumberUnavailable)
		runtimeToUpdate.Status.FuseNumberAvailable = int32(fuses.Status.NumberAvailable)
		if fuses.Status.DesiredNumberScheduled == fuses.Status.NumberReady {
			runtimeToUpdate.Status.FusePhase = data.RuntimePhaseReady
			fuseReady = true
		} else if fuses.Status.NumberAvailable == fuses.Status.NumberReady {
			runtimeToUpdate.Status.FusePhase = data.RuntimePhasePartialReady
			fuseReady = true
		} else {
			runtimeToUpdate.Status.FusePhase = data.RuntimePhaseNotReady
		}

		if masterReady && workerReady && fuseReady {
			ready = true
		}

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			err = e.Client.Status().Update(context.TODO(), runtimeToUpdate)
			if err != nil {
				e.Log.Error(err, "Failed to update the runtime")
			}
		} else {
			e.Log.Info("Do nothing because the runtime status is not changed.")
		}

		return err
	})

	return
}
