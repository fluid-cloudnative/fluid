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
	"reflect"
	"time"

	"k8s.io/client-go/util/retry"

	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// CheckAndUpdateRuntimeStatus checks the related runtime status and updates it.
func (j *JuiceFSEngine) CheckAndUpdateRuntimeStatus() (ready bool, err error) {

	var (
		workerReady bool
		workerName  string = j.getWorkerDaemonsetName()
		namespace   string = j.namespace
	)

	// 1. Worker should be ready
	workers, err := j.getDaemonset(workerName, namespace)
	if err != nil {
		return ready, err
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := j.getRuntime()
		if err != nil {
			return err
		}

		runtimeToUpdate := runtime.DeepCopy()
		if reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			j.Log.V(1).Info("The runtime is equal after deepcopy")
		}

		states, err := j.queryCacheStatus()
		if err != nil {
			return err
		}

		// 0. Update the cache status
		if len(runtime.Status.CacheStates) == 0 {
			runtimeToUpdate.Status.CacheStates = map[common.CacheStateName]string{}
		}

		runtimeToUpdate.Status.CacheStates[common.CacheCapacity] = states.cacheCapacity

		// Todo:show this infomathion when complete dataload function
		//runtimeToUpdate.Status.CacheStates[common.CachedPercentage] = states.cachedPercentage
		//runtimeToUpdate.Status.CacheStates[common.Cached] = states.cached
		runtimeToUpdate.Status.CacheStates[common.CachedPercentage] = "-"
		runtimeToUpdate.Status.CacheStates[common.Cached] = "-"
		// 1. Update cache hit ratio
		runtimeToUpdate.Status.CacheStates[common.CacheHitRatio] = states.cacheHitRatio

		// 2. Update cache throughput ratio
		runtimeToUpdate.Status.CacheStates[common.CacheThroughputRatio] = states.cacheThroughputRatio

		runtimeToUpdate.Status.CurrentWorkerNumberScheduled = int32(workers.Status.CurrentNumberScheduled)
		runtimeToUpdate.Status.DesiredWorkerNumberScheduled = int32(workers.Status.DesiredNumberScheduled)

		runtimeToUpdate.Status.WorkerNumberReady = int32(workers.Status.NumberReady)
		runtimeToUpdate.Status.WorkerNumberUnavailable = int32(workers.Status.NumberUnavailable)
		runtimeToUpdate.Status.WorkerNumberAvailable = int32(workers.Status.NumberAvailable)
		if runtime.Replicas() == workers.Status.NumberReady {
			runtimeToUpdate.Status.WorkerPhase = data.RuntimePhaseReady
			workerReady = true
		} else if workers.Status.NumberAvailable == workers.Status.NumberReady {
			runtimeToUpdate.Status.WorkerPhase = data.RuntimePhasePartialReady
			workerReady = true
		} else {
			runtimeToUpdate.Status.WorkerPhase = data.RuntimePhaseNotReady
		}

		if workerReady {
			ready = true
		}

		// Update the setup time of JuiceFS runtime
		if ready && runtimeToUpdate.Status.SetupDuration == "" {
			runtimeToUpdate.Status.SetupDuration = utils.CalculateDuration(runtimeToUpdate.CreationTimestamp.Time, time.Now())
		}

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			err = j.Client.Status().Update(context.TODO(), runtimeToUpdate)
			if err != nil {
				j.Log.Error(err, "Failed to update the runtime")
			}
		} else {
			j.Log.Info("Do nothing because the runtime status is not changed.")
		}

		return err
	})

	return
}
