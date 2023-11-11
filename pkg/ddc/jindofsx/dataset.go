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

package jindofsx

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

func (e *JindoFSxEngine) UpdateDatasetStatus(phase datav1alpha1.DatasetPhase) (err error) {
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
				// Stores binding relation between dataset and runtime
				if len(datasetToUpdate.Status.Runtimes) == 0 {
					datasetToUpdate.Status.Runtimes = []datav1alpha1.Runtime{}
				}

				datasetToUpdate.Status.Runtimes = utils.AddRuntimesIfNotExist(datasetToUpdate.Status.Runtimes, utils.NewRuntime(e.name,
					e.namespace,
					common.AccelerateCategory,
					common.JindoRuntime,
					e.runtime.Spec.Master.Replicas))

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

func (e *JindoFSxEngine) UpdateCacheOfDataset() (err error) {
	defer utils.TimeTrack(time.Now(), "JindoFSxEngine.UpdateCacheOfDataset", "name", e.name, "namespace", e.namespace)
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

		e.Log.Info("the dataset status", "status", datasetToUpdate.Status)

		if datasetToUpdate.Status.HCFSStatus == nil {
			if runtime.Spec.Master.Disabled {
				datasetToUpdate.Status.HCFSStatus, err = e.GetHCFSStatusWithoutMaster()
			} else {
				datasetToUpdate.Status.HCFSStatus, err = e.GetHCFSStatus()
			}
			if err != nil {
				return err
			}
		} else {
			e.Log.Info("No need to update HCFS status")
		}

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

func (e *JindoFSxEngine) BindToDataset() (err error) {
	return e.UpdateDatasetStatus(datav1alpha1.BoundDatasetPhase)
}
