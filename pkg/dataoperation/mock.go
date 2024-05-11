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

package dataoperation

import (
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func BuildMockDataloadOperationReconcilerInterface(expectType datav1alpha1.OperationType, ttlSecondsAfterFinished *int32) (operation OperationInterface) {

	return &mockDataloadOperationReconciler{
		expectType:              expectType,
		TTLSecondsAfterFinished: ttlSecondsAfterFinished,
	}
}

type mockDataloadOperationReconciler struct {
	expectType              datav1alpha1.OperationType
	TTLSecondsAfterFinished *int32
}

func (m mockDataloadOperationReconciler) HasPrecedingOperation() bool {
	panic("unimplemented")
}

func (m mockDataloadOperationReconciler) GetOperationObject() client.Object {
	panic("unimplemented")
}

// GetChartsDirectory implements OperationInterface.
func (mockDataloadOperationReconciler) GetChartsDirectory() string {
	panic("unimplemented")
}

// GetOperationType implements OperationInterface.
func (m mockDataloadOperationReconciler) GetOperationType() datav1alpha1.OperationType {
	return datav1alpha1.DataLoadType
}

// GetReleaseNameSpacedName implements OperationInterface.
func (mockDataloadOperationReconciler) GetReleaseNameSpacedName() types.NamespacedName {
	panic("unimplemented")
}

// GetStatusHandler implements OperationInterface.
func (mockDataloadOperationReconciler) GetStatusHandler() StatusHandler {
	panic("unimplemented")
}

// GetTTL implements OperationInterface.
func (m mockDataloadOperationReconciler) GetTTL() (ttl *int32, err error) {
	if m.expectType != datav1alpha1.DataLoadType {
		err = fmt.Errorf("the dataoperation type is %s, not DataloadType", m.expectType)
	}
	return m.TTLSecondsAfterFinished, err
}

func (r *mockDataloadOperationReconciler) GetParallelTaskNumber() int32 {
	return 1
}

// GetTargetDataset implements OperationInterface.
func (m mockDataloadOperationReconciler) GetTargetDataset() (*datav1alpha1.Dataset, error) {
	panic("unimplemented")
}

// RemoveTargetDatasetStatusInProgress implements OperationInterface.
func (mockDataloadOperationReconciler) RemoveTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	panic("unimplemented")
}

// SetTargetDatasetStatusInProgress implements OperationInterface.
func (mockDataloadOperationReconciler) SetTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	panic("unimplemented")
}

// UpdateOperationApiStatus implements OperationInterface.
func (mockDataloadOperationReconciler) UpdateOperationApiStatus(opStatus *datav1alpha1.OperationStatus) error {
	panic("unimplemented")
}

// UpdateStatusInfoForCompleted implements OperationInterface.
func (mockDataloadOperationReconciler) UpdateStatusInfoForCompleted(infos map[string]string) error {
	panic("unimplemented")
}

// Validate implements OperationInterface.
func (mockDataloadOperationReconciler) Validate(ctx runtime.ReconcileRequestContext) ([]datav1alpha1.Condition, error) {
	panic("unimplemented")
}
