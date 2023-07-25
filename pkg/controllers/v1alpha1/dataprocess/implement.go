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

package dataprocess

import (
	"context"
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *DataProcessReconciler) GetTargetDataset(object client.Object) (*datav1alpha1.Dataset, error) {
	dataProcess, ok := object.(*datav1alpha1.DataProcess)
	if !ok {
		return nil, fmt.Errorf("object %v is not of type DataProcess", object)
	}

	return utils.GetDataset(r.Client, dataProcess.Spec.Dataset.Name, dataProcess.Spec.Dataset.Namespace)
}

// GetReleaseNameSpacedName get the installed helm chart name
func (r *DataProcessReconciler) GetReleaseNameSpacedName(object client.Object) types.NamespacedName {
	return types.NamespacedName{
		Namespace: object.GetNamespace(),
		Name:      utils.GetDataProcessReleaseName(object.GetName()),
	}
}

// GetChartsDirectory get the helm charts directory of data operation
func (r *DataProcessReconciler) GetChartsDirectory() string {
	panic("not implemented") // TODO: Implement
}

// GetOperationType get the data operation type
func (r *DataProcessReconciler) GetOperationType() dataoperation.OperationType {
	return dataoperation.DataProcess
}

// UpdateOperationApiStatus update the data operation status, object is the data operation crd instance.
func (r *DataProcessReconciler) UpdateOperationApiStatus(object client.Object, opStatus *datav1alpha1.OperationStatus) error {
	dataProcess, ok := object.(*datav1alpha1.DataProcess)
	if !ok {
		return fmt.Errorf("%+v is not of type DataProcess", object)
	}

	var dataProcessCopy = dataProcess.DeepCopy()
	dataProcessCopy.Status = *opStatus.DeepCopy()
	return r.Status().Update(context.TODO(), dataProcessCopy)
}

// Validate check the data operation spec is valid or not, if not valid return error with conditions
func (r *DataProcessReconciler) Validate(ctx runtime.ReconcileRequestContext, object client.Object) ([]datav1alpha1.Condition, error) {
	panic("not implemented") // TODO: Implement
}

// UpdateStatusInfoForCompleted update the status infos field for phase completed, the parameter infos is not nil
func (r *DataProcessReconciler) UpdateStatusInfoForCompleted(object client.Object, infos map[string]string) error {
	return nil
}

// SetTargetDatasetStatusInProgress set the dataset status for certain field when data operation executing.
func (r *DataProcessReconciler) SetTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	// DataProcess does not need to update Dataset status before execution.
}

// RemoveTargetDatasetStatusInProgress remove the dataset status for certain field when data operation finished.
func (r *DataProcessReconciler) RemoveTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	// DataProcess does not need to recover Dataset status after execution.
}

func (r *DataProcessReconciler) GetStatusHandler(object client.Object) dataoperation.StatusHandler {
	panic("not implemented") // TODO: Implement
}
