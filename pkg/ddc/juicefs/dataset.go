/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package juicefs

import (
	"context"
	"reflect"

	"k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// UpdateDatasetStatus updates the status of the dataset
func (j *JuiceFSEngine) UpdateDatasetStatus(phase datav1alpha1.DatasetPhase) (err error) {
	// 1. update the runtime status
	runtime, err := j.getRuntime()
	if err != nil {
		return err
	}

	// 2. update the dataset status
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		dataset, err := utils.GetDataset(j.Client, j.name, j.namespace)
		if err != nil {
			return err
		}
		datasetToUpdate := dataset.DeepCopy()
		var cond datav1alpha1.DatasetCondition

		switch phase {
		case datav1alpha1.BoundDatasetPhase:
			// Stores dataset mount info
			if len(datasetToUpdate.Status.Mounts) == 0 {
				datasetToUpdate.Status.Mounts = datasetToUpdate.Spec.Mounts
			}

			// Stores binding relation between dataset and runtime
			if len(datasetToUpdate.Status.Runtimes) == 0 {
				datasetToUpdate.Status.Runtimes = []datav1alpha1.Runtime{}
			}

			datasetToUpdate.Status.Runtimes = utils.AddRuntimesIfNotExist(datasetToUpdate.Status.Runtimes, utils.NewRuntime(j.name,
				j.namespace,
				common.AccelerateCategory,
				common.JuiceFSRuntime,
				j.runtime.Spec.Replicas))

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

		if datasetToUpdate.Status.Phase != datav1alpha1.DataMigrating {
			datasetToUpdate.Status.Phase = phase
		}
		datasetToUpdate.Status.Conditions = utils.UpdateDatasetCondition(datasetToUpdate.Status.Conditions,
			cond)

		datasetToUpdate.Status.CacheStates = runtime.Status.CacheStates

		if !reflect.DeepEqual(dataset.Status, datasetToUpdate.Status) {
			err = j.Client.Status().Update(context.TODO(), datasetToUpdate)
			if err != nil {
				return err
			} else {
				j.Log.Info("No need to update the cache of the data")
			}
		}

		return nil
	})

	if err != nil {
		return utils.LoggingErrorExceptConflict(j.Log, err, "Failed to Update dataset",
			types.NamespacedName{Namespace: j.namespace, Name: j.name})
	}

	return
}

func (j *JuiceFSEngine) UpdateCacheOfDataset() (err error) {
	// 1. update the runtime status
	runtime, err := j.getRuntime()
	if err != nil {
		return err
	}

	// 2.update the dataset status
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		dataset, err := utils.GetDataset(j.Client, j.name, j.namespace)
		if err != nil {
			return err
		}
		datasetToUpdate := dataset.DeepCopy()

		datasetToUpdate.Status.CacheStates = runtime.Status.CacheStates

		j.Log.Info("the dataset status", "status", datasetToUpdate.Status)

		if !reflect.DeepEqual(dataset.Status, datasetToUpdate.Status) {
			err = j.Client.Status().Update(context.TODO(), datasetToUpdate)
			return err
		} else {
			j.Log.Info("No need to update the cache of the data")
		}

		return nil
	})

	if err != nil {
		return utils.LoggingErrorExceptConflict(j.Log, err, "Failed to Update dataset",
			types.NamespacedName{Namespace: j.namespace, Name: j.name})
	}

	return
}

func (j *JuiceFSEngine) BindToDataset() (err error) {

	return j.UpdateDatasetStatus(datav1alpha1.BoundDatasetPhase)
}
