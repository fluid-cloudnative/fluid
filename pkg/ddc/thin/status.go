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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"reflect"
	"time"
)

func (t *ThinEngine) CheckAndUpdateRuntimeStatus() (bool, error) {
	dataset, err := utils.GetDataset(t.Client, t.name, t.namespace)
	if err != nil {
		return false, err
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := t.getRuntime()
		if err != nil {
			return err
		}
		runtimeToUpdate := runtime.DeepCopy()

		if runtimeToUpdate.Status.FusePhase == datav1alpha1.RuntimePhaseReady {
			return nil
		}

		runtimeToUpdate.Status.FusePhase = datav1alpha1.RuntimePhaseReady

		if len(runtime.Status.CacheStates) == 0 {
			runtimeToUpdate.Status.CacheStates = map[common.CacheStateName]string{
				common.CacheCapacity:        "N/A",
				common.CachedPercentage:     "N/A",
				common.Cached:               "N/A",
				common.CacheHitRatio:        "N/A",
				common.CacheThroughputRatio: "N/A",
			}
		}

		runtimeToUpdate.Status.ValueFileConfigmap = "N/A"
		if t.ifRuntimeHelmValueEnable() {
			runtimeToUpdate.Status.ValueFileConfigmap = t.getHelmValuesConfigMapName()
		}

		// Update the setup time of thinFS runtime
		if runtimeToUpdate.Status.SetupDuration == "" {
			runtimeToUpdate.Status.SetupDuration = utils.CalculateDuration(runtimeToUpdate.CreationTimestamp.Time, time.Now())
		}

		// update mount status
		var statusMountsToUpdate []datav1alpha1.Mount
		for _, mount := range dataset.Status.Mounts {
			optionExcludedMount := mount.DeepCopy()
			optionExcludedMount.EncryptOptions = nil
			optionExcludedMount.Options = nil
			statusMountsToUpdate = append(statusMountsToUpdate, *optionExcludedMount)
		}
		runtimeToUpdate.Status.Mounts = statusMountsToUpdate

		// update condition
		if len(runtimeToUpdate.Status.Conditions) == 0 {
			runtimeToUpdate.Status.Conditions = []datav1alpha1.RuntimeCondition{}
		}
		cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersInitialized, datav1alpha1.RuntimeWorkersInitializedReason,
			"The fuse is initialized.", corev1.ConditionTrue)
		runtimeToUpdate.Status.Conditions =
			utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
				cond)

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			return t.Client.Status().Update(context.TODO(), runtimeToUpdate)
		}

		return nil
	})

	if err != nil {
		t.Log.Error(err, "Update runtime status")
		return false, err
	}
	return true, nil
}
