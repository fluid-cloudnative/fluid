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

package thin

import (
	"context"
	"reflect"
	"time"

	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"k8s.io/client-go/util/retry"
)

func (t *ThinEngine) CheckAndUpdateRuntimeStatus() (ready bool, err error) {
	var (
		workerReady bool
		workerName  string = t.getWorkerName()
		namespace   string = t.namespace
	)

	dataset, err := utils.GetDataset(t.Client, t.name, t.namespace)
	if err != nil {
		return ready, err
	}

	// 1. Worker should be ready
	workers, err := kubeclient.GetStatefulSet(t.Client, workerName, namespace)
	if err != nil {
		return ready, err
	}

	var workerNodeAffinity = kubeclient.MergeNodeSelectorAndNodeAffinity(workers.Spec.Template.Spec.NodeSelector, workers.Spec.Template.Spec.Affinity)

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := t.getRuntime()
		if err != nil {
			return err
		}

		runtimeToUpdate := runtime.DeepCopy()
		if reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			t.Log.V(1).Info("The runtime is equal after deepcopy")
		}

		// todo: maybe set query shell in runtime
		// 0. Update the cache status
		if len(runtime.Status.CacheStates) == 0 {
			runtimeToUpdate.Status.CacheStates = map[common.CacheStateName]string{}
		}

		// set node affinity
		runtimeToUpdate.Status.CacheAffinity = workerNodeAffinity

		runtimeToUpdate.Status.CacheStates[common.CacheCapacity] = "N/A"
		runtimeToUpdate.Status.CacheStates[common.CachedPercentage] = "N/A"
		runtimeToUpdate.Status.CacheStates[common.Cached] = "N/A"
		runtimeToUpdate.Status.CacheStates[common.CacheHitRatio] = "N/A"
		runtimeToUpdate.Status.CacheStates[common.CacheThroughputRatio] = "N/A"

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

		if workerReady {
			ready = true
		}

		// Update the setup time of thinFS runtime
		if ready && runtimeToUpdate.Status.SetupDuration == "" {
			runtimeToUpdate.Status.SetupDuration = utils.CalculateDuration(runtimeToUpdate.CreationTimestamp.Time, time.Now())
		}

		var statusMountsToUpdate []data.Mount
		for _, mount := range dataset.Status.Mounts {
			optionExcludedMount := mount.DeepCopy()
			optionExcludedMount.EncryptOptions = nil
			optionExcludedMount.Options = nil
			statusMountsToUpdate = append(statusMountsToUpdate, *optionExcludedMount)
		}
		runtimeToUpdate.Status.Mounts = statusMountsToUpdate

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			err = t.Client.Status().Update(context.TODO(), runtimeToUpdate)
			if err != nil {
				t.Log.Error(err, "Failed to update the runtime")
			}
		} else {
			t.Log.Info("Do nothing because the runtime status is not changed.")
		}

		return err
	})

	return
}

func (t *ThinEngine) UpdateRuntimeSetConfigIfNeeded() (updated bool, err error) {
	fuseAddresses, err := t.Helper.GetIpAddressesOfFuse()
	if err != nil {
		return
	}

	workerAddresses, err := t.Helper.GetIpAddressesOfWorker()
	if err != nil {
		return
	}

	configMapName := t.runtimeInfo.GetName() + "-runtimeset"
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		cm, err := kubeclient.GetConfigmapByName(t.Client, configMapName, t.namespace)
		if err != nil {
			return err
		}

		if cm == nil {
			t.Log.Info("configmap is not found", "key", configMapName)
			return nil
		}

		cmToUpdate := cm.DeepCopy()
		result, err := t.toRuntimeSetConfig(workerAddresses,
			fuseAddresses)
		if err != nil {
			return err
		}
		cmToUpdate.Data["runtime.json"] = result

		if !reflect.DeepEqual(cm, cmToUpdate) {
			err = t.Client.Update(context.TODO(), cmToUpdate)
			if err != nil {
				t.Log.Error(err, "Failed to update the ip addresses of runtime")
			}
			updated = true
		} else {
			t.Log.Info("Do nothing because the ip addresses of runtime are not changed.")
		}

		return nil

	})

	return
}
