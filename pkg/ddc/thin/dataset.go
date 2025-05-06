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

package thin

import (
	"context"
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

func (t *ThinEngine) UpdateDatasetStatus(phase datav1alpha1.DatasetPhase) (err error) {
	runtime, err := t.getRuntime()
	if err != nil {
		return err
	}

	// 2. update the dataset status
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		dataset, err := utils.GetDataset(t.Client, t.name, t.namespace)
		if err != nil {
			return err
		}

		datasetToUpdate := dataset.DeepCopy()
		var cond datav1alpha1.DatasetCondition

		if phase == datav1alpha1.BoundDatasetPhase {
			// Stores dataset mount info
			if len(datasetToUpdate.Status.Mounts) == 0 {
				datasetToUpdate.Status.Mounts = datasetToUpdate.Spec.Mounts
			}

			// Stores binding relation between dataset and runtime
			if len(datasetToUpdate.Status.Runtimes) == 0 {
				datasetToUpdate.Status.Runtimes = []datav1alpha1.Runtime{}
			}

			datasetToUpdate.Status.Runtimes = utils.AddRuntimesIfNotExist(datasetToUpdate.Status.Runtimes, utils.NewRuntime(t.name,
				t.namespace,
				common.AccelerateCategory,
				common.ThinRuntime,
				t.runtime.Spec.Replicas))
		}

		if datasetToUpdate.Status.Phase != phase {
			datasetToUpdate.Status.Phase = phase

			switch phase {
			case datav1alpha1.BoundDatasetPhase:
				cond = utils.NewDatasetCondition(datav1alpha1.DatasetReady, datav1alpha1.DatasetReadyReason,
					"The ddc runtime is ready.",
					corev1.ConditionTrue)
			case datav1alpha1.FailedDatasetPhase:
				cond = utils.NewDatasetCondition(datav1alpha1.DatasetReady, datav1alpha1.DatasetReadyReason,
					"The ddc runtime is not ready.",
					corev1.ConditionFalse)
			default:
				cond = utils.NewDatasetCondition(datav1alpha1.DatasetReady, datav1alpha1.DatasetReadyReason,
					"The ddc runtime is unknown.",
					corev1.ConditionFalse)
			}
			datasetToUpdate.Status.Conditions = utils.UpdateDatasetCondition(datasetToUpdate.Status.Conditions,
				cond)
		}

		datasetToUpdate.Status.CacheStates = runtime.Status.CacheStates
		if !reflect.DeepEqual(dataset.Status, datasetToUpdate.Status) {
			t.Log.V(1).Info("Update DatasetStatus", "dataset", fmt.Sprintf("%s/%s", datasetToUpdate.GetNamespace(), datasetToUpdate.GetName()))
			err = t.Client.Status().Update(context.TODO(), datasetToUpdate)
			if err != nil {
				return err
			} else {
				t.Log.Info("No need to update the cache of the data")
			}
		}

		return nil
	})

	if err != nil {
		return utils.LoggingErrorExceptConflict(t.Log, err, "Failed to Update dataset",
			types.NamespacedName{Namespace: t.namespace, Name: t.name})
	}

	return
}

func (t *ThinEngine) UpdateCacheOfDataset() (err error) {
	return
}

func (t *ThinEngine) BindToDataset() (err error) {
	return t.UpdateDatasetStatus(datav1alpha1.BoundDatasetPhase)
}
