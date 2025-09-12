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

package efc

import (
	"context"
	"reflect"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

// CheckAndUpdateRuntimeStatus checks the related runtime status and updates it.
func (e *EFCEngine) CheckAndUpdateRuntimeStatus() (ready bool, err error) {
	var (
		masterReady, workerReady bool
		masterName               string = e.getMasterName()
		workerName               string = e.getWorkerName()
		namespace                string = e.namespace
	)

	// 1. Master should be ready
	master, err := kubeclient.GetStatefulSet(e.Client, masterName, namespace)
	if err != nil {
		return ready, err
	}

	// 2. Worker should be ready
	workers, err := ctrl.GetWorkersAsStatefulset(e.Client,
		types.NamespacedName{Namespace: e.namespace, Name: workerName})
	if err != nil {
		return ready, err
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		runtimeToUpdate := runtime.DeepCopy()

		states, err := e.queryCacheStatus()
		if err != nil {
			return err
		}

		if len(runtime.Status.CacheStates) == 0 {
			runtimeToUpdate.Status.CacheStates = map[common.CacheStateName]string{}
		}

		runtimeToUpdate.Status.CacheStates[common.CacheCapacity] = states.cacheCapacity
		runtimeToUpdate.Status.CacheStates[common.CachedPercentage] = states.cachedPercentage
		runtimeToUpdate.Status.CacheStates[common.Cached] = states.cached
		// update cache hit ratio
		runtimeToUpdate.Status.CacheStates[common.CacheHitRatio] = states.cacheHitStates.cacheHitRatio
		runtimeToUpdate.Status.CacheStates[common.LocalHitRatio] = states.cacheHitStates.localHitRatio
		runtimeToUpdate.Status.CacheStates[common.RemoteHitRatio] = states.cacheHitStates.remoteHitRatio
		// update cache throughput ratio
		runtimeToUpdate.Status.CacheStates[common.LocalThroughputRatio] = states.cacheHitStates.localThroughputRatio
		runtimeToUpdate.Status.CacheStates[common.RemoteThroughputRatio] = states.cacheHitStates.remoteThroughputRatio
		runtimeToUpdate.Status.CacheStates[common.CacheThroughputRatio] = states.cacheHitStates.cacheThroughputRatio

		if *master.Spec.Replicas == master.Status.ReadyReplicas {
			masterReady = true
		}

		if runtime.Replicas() == 0 || workers.Status.ReadyReplicas > 0 {
			workerReady = true
		}

		if masterReady && workerReady {
			ready = true
		}

		// Update the setup time of EFC runtime
		if ready && runtimeToUpdate.Status.SetupDuration == "" {
			runtimeToUpdate.Status.SetupDuration = utils.CalculateDuration(runtimeToUpdate.CreationTimestamp.Time, time.Now())
		}

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			err = e.Client.Status().Update(context.TODO(), runtimeToUpdate)
		} else {
			e.Log.Info("Do nothing because the runtime status is not changed.")
		}

		return err
	})

	if err != nil {
		_ = utils.LoggingErrorExceptConflict(e.Log, err, "Failed to update runtime status", types.NamespacedName{Namespace: e.namespace, Name: e.name})
	}

	return
}
