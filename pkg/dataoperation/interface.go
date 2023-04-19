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

package dataoperation

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// OperationInterface the interface of data operation crd
type OperationInterface interface {
	// GetTargetDatasetNamespacedName get the target dataeset namespace and name of the data operation
	GetTargetDatasetNamespacedName(object client.Object) (*types.NamespacedName, error)

	// GetReleaseNameSpacedName get the installed helm chart name
	GetReleaseNameSpacedName(object client.Object) types.NamespacedName

	// GetChartsDirectory get the helm charts directory of data operation
	GetChartsDirectory() string

	// GetOperationType get the data operation type and also used as a lock key for dataset
	GetOperationType() OperationType

	// UpdateOperationApiStatus update the data operation status, object is the data operation crd instance.
	UpdateOperationApiStatus(object client.Object, opStatus *datav1alpha1.OperationStatus) error

	// Validate check the data operation spec is valid or not, if not valid return error with conditions
	Validate(ctx runtime.ReconcileRequestContext, object client.Object) ([]datav1alpha1.Condition, error)

	// UpdateStatusInfoForCompleted update the status infos field for phase completed, the parameter infos is not nil
	UpdateStatusInfoForCompleted(object client.Object, infos map[string]string) error

	// UpdateStatusByHelmStatus update the operation status according to helm job status
	UpdateStatusByHelmStatus(ctx runtime.ReconcileRequestContext, object client.Object, opStatus *datav1alpha1.OperationStatus) error

	// SetTargetDatasetStatusInProgress set the dataset status for certain field when data operation executing.
	SetTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset)

	// RemoveTargetDatasetStatusInProgress remove the dataset status for certain field when data operation finished.
	RemoveTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset)
}
