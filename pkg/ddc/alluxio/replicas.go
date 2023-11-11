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

package alluxio

import (
	"context"
	"fmt"
	"reflect"

	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

// SyncReplicas syncs the replicas
func (e *AlluxioEngine) SyncReplicas(ctx cruntime.ReconcileRequestContext) (err error) {
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		workers, err := ctrl.GetWorkersAsStatefulset(e.Client,
			types.NamespacedName{Namespace: e.namespace, Name: e.getWorkerName()})
		if err != nil {
			if fluiderrs.IsDeprecated(err) {
				e.Log.Info("Warning: the current runtime is created by runtime controller before v0.7.0, scale out/in are not supported. To support these features, please create a new dataset", "details", err)
				return nil
			}
			if errors.IsNotFound(err) {
				cond := utils.NewRuntimeCondition(data.RuntimeWorkersReady, "The workers are not ready.",
					fmt.Sprintf("The statefulset %s in %s is not found, please fix it.",
						e.getWorkerName(),
						e.namespace), corev1.ConditionFalse)

				updateErr := retry.RetryOnConflict(retry.DefaultBackoff, func() error {

					runtime, err := e.getRuntime()
					if err != nil {
						return err
					}

					runtimeToUpdate := runtime.DeepCopy()

					_, oldCond := utils.GetRuntimeCondition(runtimeToUpdate.Status.Conditions, cond.Type)

					if oldCond == nil || oldCond.Type != cond.Type {
						runtimeToUpdate.Status.Conditions =
							utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
								cond)
					}

					runtimeToUpdate.Status.WorkerPhase = data.RuntimePhaseNotReady
					e.Log.Error(err, "the worker are not ready")

					if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
						updateErr := e.Client.Status().Update(context.TODO(), runtimeToUpdate)
						if updateErr != nil {
							return updateErr
						}

						updateErr = e.UpdateDatasetStatus(data.FailedDatasetPhase)
						if updateErr != nil {
							e.Log.Error(updateErr, "Failed to update dataset")
							return updateErr
						}
					}

					return err
				})
				totalErr := fmt.Errorf("the master engine is not existed %v", updateErr)
				return totalErr
			}
			return err
		}
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}
		runtimeToUpdate := runtime.DeepCopy()
		err = e.Helper.SyncReplicas(ctx, runtimeToUpdate, runtimeToUpdate.Status, workers)
		return err
	})
	if err != nil {
		_ = utils.LoggingErrorExceptConflict(e.Log, err, "Failed to sync replicas", types.NamespacedName{Namespace: e.namespace, Name: e.name})
	}

	return
}
