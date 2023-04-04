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
	"fmt"
	"reflect"

	"github.com/fluid-cloudnative/fluid/pkg/common"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

func (e *EFCEngine) BindToDataset() (err error) {
	return e.UpdateDatasetStatus(datav1alpha1.BoundDatasetPhase)
}

// UpdateDatasetStatus updates the status of the dataset
func (e *EFCEngine) UpdateDatasetStatus(phase datav1alpha1.DatasetPhase) (err error) {
	// 1. update the runtime status
	runtime, err := e.getRuntime()
	if err != nil {
		return err
	}

	// 2.update the dataset status
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
		if err != nil {
			return err
		}
		datasetToUpdate := dataset.DeepCopy()

		if phase != dataset.Status.Phase {
			var cond datav1alpha1.DatasetCondition

			switch phase {
			case datav1alpha1.BoundDatasetPhase:
				if len(datasetToUpdate.Status.Mounts) == 0 {
					datasetToUpdate.Status.Mounts = datasetToUpdate.Spec.Mounts
				}
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

			datasetToUpdate.Status.Phase = phase
			datasetToUpdate.Status.Conditions = utils.UpdateDatasetCondition(datasetToUpdate.Status.Conditions,
				cond)
		}

		datasetToUpdate.Status.CacheStates = runtime.Status.CacheStates

		e.Log.Info("the dataset status", "status", datasetToUpdate.Status)

		if !reflect.DeepEqual(dataset.Status, datasetToUpdate.Status) {
			err = e.Client.Status().Update(context.TODO(), datasetToUpdate)
			if err != nil {
				e.Log.Error(err, "Update dataset")
				return err
			}
		}

		return err
	})

	if err != nil {
		_ = utils.LoggingErrorExceptConflict(e.Log, err, "Failed to update dataset status", types.NamespacedName{Namespace: e.namespace, Name: e.name})
		return err
	}

	return
}

// UpdateCacheOfDataset updates the CacheStates and Runtimes of the dataset.
func (e *EFCEngine) UpdateCacheOfDataset() (err error) {
	// 1. update the runtime status
	runtime, err := e.getRuntime()
	if err != nil {
		return err
	}

	// 2.update the dataset status
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
		if err != nil {
			return err
		}
		datasetToUpdate := dataset.DeepCopy()

		datasetToUpdate.Status.CacheStates = runtime.Status.CacheStates

		if len(datasetToUpdate.Status.Runtimes) == 0 {
			datasetToUpdate.Status.Runtimes = []datav1alpha1.Runtime{}
		}

		datasetToUpdate.Status.Runtimes = utils.AddRuntimesIfNotExist(datasetToUpdate.Status.Runtimes, utils.NewRuntime(e.name,
			e.namespace,
			common.AccelerateCategory,
			common.EFCRuntime,
			e.runtime.MasterReplicas()))

		e.Log.Info("the dataset status", "status", datasetToUpdate.Status)

		if !reflect.DeepEqual(dataset.Status, datasetToUpdate.Status) {
			err = e.Client.Status().Update(context.TODO(), datasetToUpdate)
			if err != nil {
				e.Log.Error(err, "Update dataset")
				return err
			}
		} else {
			e.Log.Info("No need to update the cache of the data")
		}

		return nil
	})

	if err != nil {
		e.Log.Error(err, "Update dataset")
		return err
	}

	// TODO: remove the following logic if supports asynchrounous labeling cache node
	if runtime.Status.WorkerPhase == datav1alpha1.RuntimePhaseNotReady || runtime.Status.WorkerPhase == datav1alpha1.RuntimePhasePartialReady {
		return fmt.Errorf("EFCRuntime \"%s/%s\" scaling not ready", runtime.Namespace, runtime.Name)
	}

	return

}
