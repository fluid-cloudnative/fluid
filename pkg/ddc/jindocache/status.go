/*
Copyright 2023 The Fluid Authors.

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

package jindocache

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
func (e *JindoCacheEngine) CheckAndUpdateRuntimeStatus() (ready bool, err error) {
	defer utils.TimeTrack(time.Now(), "JindoCacheEngine.CheckAndUpdateRuntimeStatus", "name", e.name, "namespace", e.namespace)

	if e.runtime.Spec.Master.Disabled && e.runtime.Spec.Worker.Disabled {
		return e.syncFuseOnlyModeRuntimeStatus()
	}

	return e.syncCacheModeRuntimeStatus()
}

func (e *JindoCacheEngine) syncFuseOnlyModeRuntimeStatus() (ready bool, err error) {
	// This is for backward compatibility. In Fuse-only mode, runtime may not have a proper HelmValuesConfigMap
	getHelmValuesConfigMapBackCompatFn := func() (cmName string, fnErr error) {
		cm, fnErr := kubeclient.GetConfigmapByName(e.Client, e.getHelmValuesConfigMapName(), e.namespace)
		if fnErr != nil {
			return "", fnErr
		}
		if cm != nil {
			return e.getHelmValuesConfigMapName(), nil
		}
		// fallback to use older version of JindoRuntime's engineImpl
		return e.name + "-" + common.JindoFSxEngineImpl + "-values", nil
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		runtimeToUpdate := runtime.DeepCopy()
		if runtimeToUpdate.Status.ValueFileConfigmap, err = getHelmValuesConfigMapBackCompatFn(); err != nil {
			return err
		}
		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			err = e.Client.Status().Update(context.TODO(), runtimeToUpdate)
		} else {
			e.Log.V(1).Info("Do nothing because the runtime status is not changed.")
		}
		return err
	})

	if err != nil {
		_ = utils.LoggingErrorExceptConflict(e.Log,
			err,
			"Failed to update the runtime",
			types.NamespacedName{
				Namespace: e.namespace,
				Name:      e.name,
			})
	}

	ready = true
	err = nil
	return
}

func (e *JindoCacheEngine) syncCacheModeRuntimeStatus() (ready bool, err error) {
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

	var workerNodeAffinity = kubeclient.MergeNodeSelectorAndNodeAffinity(workers.Spec.Template.Spec.NodeSelector, workers.Spec.Template.Spec.Affinity)

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		runtimeToUpdate := runtime.DeepCopy()
		// if reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
		// 	e.Log.V(1).Info("The runtime is equal after deepcopy")
		// }

		states, err := e.queryCacheStatus()
		if err != nil {
			return err
		}

		// 0. Update the cache status
		// runtimeToUpdate.Status.CacheStates[data.Cacheable] = states.cacheable
		if len(runtime.Status.CacheStates) == 0 {
			runtimeToUpdate.Status.CacheStates = map[common.CacheStateName]string{}
		}

		// set node affinity
		runtimeToUpdate.Status.CacheAffinity = workerNodeAffinity

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
		runtimeToUpdate.Status.ValueFileConfigmap = e.getHelmValuesConfigMapName()

		if masterReady && workerReady {
			ready = true
		}

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			err = e.Client.Status().Update(context.TODO(), runtimeToUpdate)
		} else {
			e.Log.Info("Do nothing because the runtime status is not changed.")
		}

		return err
	})

	if err != nil {
		_ = utils.LoggingErrorExceptConflict(e.Log,
			err,
			"Failed to update the runtime",
			types.NamespacedName{
				Namespace: e.namespace,
				Name:      e.name,
			})
	}

	return
}
