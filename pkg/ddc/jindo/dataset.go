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

package jindo

import (
	"context"
	"reflect"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
)

// UpdateDatasetStatus updates the status of a dataset in the JindoEngine.
// This function is primarily responsible for updating the phase and conditions of the dataset based on the given phase,
// and updating the cache state based on the runtime status.
//
// Parameters:
//   - phase (datav1alpha1.DatasetPhase): The new phase to update the dataset to.
//
// Returns:
//   - err (error): Returns an error if the update process fails, otherwise returns nil.
func (e *JindoEngine) UpdateDatasetStatus(phase datav1alpha1.DatasetPhase) (err error) {
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
		var cond datav1alpha1.DatasetCondition

		if phase != dataset.Status.Phase {
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

		return nil
	})

	if err != nil {
		e.Log.Error(err, "Update dataset")
		return err
	}

	return
}

func (e *JindoEngine) UpdateCacheOfDataset() (err error) {
	defer utils.TimeTrack(time.Now(), "JindoEngine.UpdateCacheOfDataset", "name", e.name, "namespace", e.namespace)
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
		// datasetToUpdate.Status.CacheStates =

		if len(datasetToUpdate.Status.Runtimes) == 0 {
			datasetToUpdate.Status.Runtimes = []datav1alpha1.Runtime{}
		}

		datasetToUpdate.Status.Runtimes = utils.AddRuntimesIfNotExist(datasetToUpdate.Status.Runtimes, utils.NewRuntime(e.name,
			e.namespace,
			common.AccelerateCategory,
			common.JindoRuntime,
			e.runtime.Spec.Master.Replicas))

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

	return
}

func (e *JindoEngine) BindToDataset() (err error) {
	return e.UpdateDatasetStatus(datav1alpha1.BoundDatasetPhase)
}
