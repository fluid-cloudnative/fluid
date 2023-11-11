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

package thin

import (
	"context"
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
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

			datasetToUpdate.Status.Runtimes = utils.AddRuntimesIfNotExist(datasetToUpdate.Status.Runtimes, utils.NewRuntime(t.name,
				t.namespace,
				common.AccelerateCategory,
				common.ThinRuntime,
				t.runtime.Spec.Replicas))

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

		datasetToUpdate.Status.CacheStates = runtime.Status.CacheStates

		if !reflect.DeepEqual(dataset.Status, datasetToUpdate.Status) {
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
	runtime, err := t.getRuntime()
	if err != nil {
		return err
	}

	// 2.update the dataset status
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		dataset, err := utils.GetDataset(t.Client, t.name, t.namespace)
		if err != nil {
			return err
		}
		datasetToUpdate := dataset.DeepCopy()

		datasetToUpdate.Status.CacheStates = runtime.Status.CacheStates

		t.Log.Info("the dataset status", "status", datasetToUpdate.Status)

		if !reflect.DeepEqual(dataset.Status, datasetToUpdate.Status) {
			err = t.Client.Status().Update(context.TODO(), datasetToUpdate)
			return err
		} else {
			t.Log.Info("No need to update the cache of the data")
		}

		return nil
	})

	if err != nil {
		return utils.LoggingErrorExceptConflict(t.Log, err, "Failed to Update dataset",
			types.NamespacedName{Namespace: t.namespace, Name: t.name})
	}

	return
}

func (t *ThinEngine) BindToDataset() (err error) {
	return t.UpdateDatasetStatus(datav1alpha1.BoundDatasetPhase)
}
