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
	"fmt"
	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/util/retry"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// CheckAndUpdateRuntimeStatus checks the related runtime status and updates it.
func (e *AlluxioEngine) CheckAndUpdateRuntimeStatus() (ready bool, err error) {
	var (
		masterReady, workerReady bool
		masterName               string = e.getMasterStatefulsetName()
		workerName               string = e.getWorkerDaemonsetName()
		namespace                string = e.namespace
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

		// Master
		if *master.Spec.Replicas == master.Status.ReadyReplicas {
			runtimeToUpdate.Status.MasterPhase = data.RuntimePhaseReady
			masterReady = true
		} else {
			runtimeToUpdate.Status.MasterPhase = data.RuntimePhaseNotReady
		}

		// Worker
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

		runtimeInfo, err := e.getRuntimeInfo()
		if err != nil {
			return err
		}

		fuseLabel, err := labels.Parse(fmt.Sprintf("%s=true", runtimeInfo.GetFuseLabelName()))
		if err != nil {
			return err
		}

		var nodeList = &corev1.NodeList{}
		err = e.Client.List(context.TODO(), nodeList, &client.ListOptions{
			LabelSelector: fuseLabel,
		})

		if err != nil {
			if !errors.IsNotFound(err) {
				return err
			}
		}

		runtimeToUpdate.Status.FuseNumberReady = int32(len(nodeList.Items))

		if masterReady && workerReady {
			ready = true
		}

		// Update the setup time of Alluxio runtime
		if ready && runtimeToUpdate.Status.SetupDuration == "" {
			runtimeToUpdate.Status.SetupDuration = utils.CalculateDuration(runtimeToUpdate.CreationTimestamp.Time, time.Now())
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
