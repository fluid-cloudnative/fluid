/*
Copyright 2024 The Fluid Authors.
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

package vineyard

import (
	"context"
	"reflect"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

// CheckAndUpdateRuntimeStatus checks the related runtime status and updates it.
func (e *VineyardEngine) CheckAndUpdateRuntimeStatus() (ready bool, err error) {

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
		return ready, err
	}

	//var workerNodeAffinity = kubeclient.MergeNodeSelectorAndNodeAffinity(workers.Spec.Template.Spec.NodeSelector, workers.Spec.Template.Spec.Affinity)

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		runtimeToUpdate := runtime.DeepCopy()

		/* TODO:(caoye) query the cache status from vineyard metrics exporter
		states, err := e.queryCacheStatus(runtime)
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
		*/

		if *master.Spec.Replicas == master.Status.ReadyReplicas {
			masterReady = true
		}

		if runtime.Replicas() == 0 || workers.Status.ReadyReplicas > 0 {
			workerReady = true
		}

		if masterReady && workerReady {
			ready = true
		}

		// Update the setup time of Vineyard runtime
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
