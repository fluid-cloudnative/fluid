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
	"reflect"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

func GetDataOperationKey(object client.Object) string {
	return object.GetName()
}

// SetDataOperationInTargetDataset set status of target dataset to mark the data operation being performed.
func SetDataOperationInTargetDataset(ctx cruntime.ReconcileRequestContext, operation dataoperation.OperationReconcilerInterface, engine Engine) error {
	targetDataset := ctx.Dataset
	object := operation.GetReconciledObject()

	// check if the bounded runtime is ready
	ready := engine.CheckRuntimeReady()
	if !ready {
		ctx.Log.V(1).Info("Bounded accelerate runtime not ready", "targetDataset", targetDataset)
		ctx.Recorder.Eventf(object,
			v1.EventTypeNormal,
			common.RuntimeNotReady,
			"Bounded accelerate runtime not ready")
		return fmt.Errorf("bounded accelerate runtime not ready")
	}

	operationTypeName := string(operation.GetOperationType())
	dataOpKey := GetDataOperationKey(object)

	// set current data operation in target dataset
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		dataset, err := utils.GetDataset(ctx.Client, targetDataset.Name, targetDataset.Namespace)
		if err != nil {
			return err
		}

		// set current data operation in the target dataset
		datasetToUpdate := dataset.DeepCopy()
		datasetToUpdate.SetDataOperationInProgress(operationTypeName, dataOpKey)
		// different operation may set other fields
		operation.SetTargetDatasetStatusInProgress(datasetToUpdate)

		if !reflect.DeepEqual(dataset.Status, datasetToUpdate.Status) {
			if err := ctx.Client.Status().Update(context.TODO(), datasetToUpdate); err != nil {
				ctx.Log.Info("fail to update target dataset's lock, will requeue", "targetDatasetName", targetDataset.Name)
				return err
			}
		}
		return nil
	})
	if err != nil {
		ctx.Log.Error(err, "can't set lock on target dataset", "targetDataset", targetDataset.Name)
	}
	return err
}

// ReleaseTargetDataset release target dataset OperationRef field which marks the data operation being performed.
func ReleaseTargetDataset(ctx cruntime.ReconcileRequestContext, operation dataoperation.OperationReconcilerInterface) error {

	dataOpKey := GetDataOperationKey(operation.GetReconciledObject())
	operationTypeName := string(operation.GetOperationType())

	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		dataset, err := operation.GetTargetDataset()
		if err != nil {
			if utils.IgnoreNotFound(err) == nil {
				statusError := err.(*apierrors.StatusError)
				ctx.Log.Info("can't find target dataset, won't release lock", "dataset", statusError.Status().Details.Name)
				return nil
			}
			// other error
			return err
		}

		datasetToUpdate := dataset.DeepCopy()
		dataOpRef := datasetToUpdate.RemoveDataOperationInProgress(operationTypeName, dataOpKey)

		if dataOpRef == "" {
			// different operation may set other fields
			operation.RemoveTargetDatasetStatusInProgress(datasetToUpdate)
		}
		if !reflect.DeepEqual(datasetToUpdate.Status, dataset.Status) {
			if err = ctx.Client.Status().Update(context.TODO(), datasetToUpdate); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		ctx.Log.Error(err, "can't release lock on target dataset")
	}
	return err
}
