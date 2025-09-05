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

package goosefs

import (
	"context"
	"reflect"
	"time"

	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

// CheckAndUpdateRuntimeStatus checks the related runtime status and updates it.
func (e *GooseFSEngine) CheckAndUpdateRuntimeStatus() (ready bool, err error) {

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
		if fluiderrs.IsDeprecated(err) {
			e.Log.Info("Warning: Deprecated mode is not support, so skip handling", "details", err)
			return ready, nil
		}
	}
	var workerNodeAffinity = kubeclient.MergeNodeSelectorAndNodeAffinity(workers.Spec.Template.Spec.NodeSelector, workers.Spec.Template.Spec.Affinity)

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

		// 0. Update the cache status
		if len(runtime.Status.CacheStates) == 0 {
			runtimeToUpdate.Status.CacheStates = map[common.CacheStateName]string{}
		}

		// set node affinity
		runtimeToUpdate.Status.CacheAffinity = workerNodeAffinity

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

		runtimeToUpdate.Status.CurrentMasterNumberScheduled = int32(master.Status.Replicas)
		runtimeToUpdate.Status.MasterNumberReady = int32(master.Status.ReadyReplicas)

		if *master.Spec.Replicas == master.Status.ReadyReplicas {
			runtimeToUpdate.Status.MasterPhase = data.RuntimePhaseReady
			masterReady = true
		} else {
			runtimeToUpdate.Status.MasterPhase = data.RuntimePhaseNotReady
		}

		runtimeToUpdate.Status.WorkerNumberReady = int32(workers.Status.ReadyReplicas)
		runtimeToUpdate.Status.WorkerNumberUnavailable = int32(*workers.Spec.Replicas - workers.Status.ReadyReplicas)
		runtimeToUpdate.Status.WorkerNumberAvailable = int32(workers.Status.CurrentReplicas)
		if runtime.Replicas() == 0 {
			runtimeToUpdate.Status.WorkerPhase = data.RuntimePhaseReady
			workerReady = true
		} else if workers.Status.ReadyReplicas > 0 {
			if runtime.Replicas() == workers.Status.ReadyReplicas {
				runtimeToUpdate.Status.WorkerPhase = data.RuntimePhaseReady
				workerReady = true
			} else if workers.Status.ReadyReplicas >= 1 {
				runtimeToUpdate.Status.WorkerPhase = data.RuntimePhasePartialReady
				workerReady = true
			}
		} else {
			runtimeToUpdate.Status.WorkerPhase = data.RuntimePhaseNotReady
		}

		if masterReady && workerReady {
			ready = true
		}

		// Update the setup time of GooseFS runtime
		if ready && runtimeToUpdate.Status.SetupDuration == "" {
			runtimeToUpdate.Status.SetupDuration = utils.CalculateDuration(runtimeToUpdate.CreationTimestamp.Time, time.Now())
		}

		// Update API Gateway address if it is enabled
		if runtimeToUpdate.Spec.APIGateway.Enabled {
			if runtimeToUpdate.Status.APIGatewayStatus == nil {
				runtimeToUpdate.Status.APIGatewayStatus, err = e.GetAPIGatewayStatus()
				if err != nil {
					return err
				}
			} else {
				e.Log.V(1).Info("No need to update APIGateway status")
			}
		} else {
			e.Log.V(1).Info("No need to update APIGateway status")
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
