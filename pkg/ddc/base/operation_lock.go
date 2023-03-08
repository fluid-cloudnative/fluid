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

package base

import (
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetDataOperationRef(name, namespace string) string {
	// namespace may contain '-', use '/' as separator
	return fmt.Sprintf("%s/%s", namespace, name)
}

// LockTargetDataset lock target dataset if not locked by this data operation.
func LockTargetDataset(ctx cruntime.ReconcileRequestContext, object client.Object,
	operation dataoperation.OperationInterface, engine Engine) error {
	targetDataset := ctx.Dataset

	operationTypeName := string(operation.GetOperationType())

	// 1. Check if there's any conflict
	conflictDataOpRef := targetDataset.GetLockedNameForOperation(operationTypeName)
	dataOpRef := GetDataOperationRef(object.GetName(), object.GetNamespace())

	// already locked by self, return
	if conflictDataOpRef == dataOpRef {
		return nil
	}

	// conflict lock, return error and requeue
	if len(conflictDataOpRef) != 0 && conflictDataOpRef != dataOpRef {
		ctx.Log.Info(fmt.Sprintf("Found other %s that is in Executing phase, will backoff", operationTypeName), "other", conflictDataOpRef)
		ctx.Recorder.Eventf(object, v1.EventTypeNormal, common.DataOperationCollision,
			"Found other %s(%s) that is in Executing phase, will backoff",
			operationTypeName, conflictDataOpRef)
		return fmt.Errorf("found other %s that is in Executing phase, will backoff", operationTypeName)
	}

	// 2. can lock, check if the bounded runtime is ready
	ready := engine.CheckRuntimeReady()
	if !ready {
		ctx.Log.V(1).Info("Bounded accelerate runtime not ready", "targetDataset", targetDataset)
		ctx.Recorder.Eventf(object,
			v1.EventTypeNormal,
			common.RuntimeNotReady,
			"Bounded accelerate runtime not ready")
		return fmt.Errorf("bounded accelerate runtime not ready")
	}

	ctx.Log.Info("No conflicts detected, try to lock the target dataset")

	// 3. Try lock target dataset
	datasetToUpdate := targetDataset.DeepCopy()
	datasetToUpdate.LockOperation(operationTypeName, dataOpRef)

	// different operation may set other fields
	operation.LockTargetDatasetStatus(datasetToUpdate)

	if !reflect.DeepEqual(targetDataset.Status, datasetToUpdate.Status) {
		if err := ctx.Client.Status().Update(context.TODO(), datasetToUpdate); err != nil {
			ctx.Log.Info("fail to update target dataset's lock, will requeue", "targetDatasetName", targetDataset.Name)
			return err
		}
	}

	return nil
}

// ReleaseTargetDataset release target dataset if locked by this data operation.
func ReleaseTargetDataset(ctx cruntime.ReconcileRequestContext, object client.Object,
	operation dataoperation.OperationInterface) error {
	// Note: ctx.Dataset may be nil, so use the `GetTargetDatasetNamespacedName`
	targetDatasetNamespacedName, err := operation.GetTargetDatasetNamespacedName(object)
	if err != nil {
		return err
	}

	operationTypeName := string(operation.GetOperationType())

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		dataset, err := utils.GetDataset(ctx.Client, targetDatasetNamespacedName.Name, targetDatasetNamespacedName.Namespace)
		if err != nil {
			if utils.IgnoreNotFound(err) == nil {
				ctx.Log.Info("can't find target dataset, won't release lock", "targetDataset", targetDatasetNamespacedName.Name)
				return nil
			}
			// other error
			return err
		}
		currentRef := dataset.GetLockedNameForOperation(operationTypeName)

		if currentRef != GetDataOperationRef(object.GetName(), object.GetNamespace()) {
			ctx.Log.Info("Found Ref inconsistent with the reconciling DataBack, won't release this lock, ignore it", "Operation", operationTypeName, "ref", currentRef)
			return nil
		}
		datasetToUpdate := dataset.DeepCopy()
		datasetToUpdate.ReleaseOperation(operationTypeName)

		// different operation may set other fields
		operation.ReleaseTargetDatasetStatus(datasetToUpdate)
		if !reflect.DeepEqual(datasetToUpdate.Status, dataset) {
			if err := ctx.Client.Status().Update(context.TODO(), datasetToUpdate); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		ctx.Log.Error(err, "can't release lock on target dataset", "targetDataset", targetDatasetNamespacedName)
	}
	return err
}
