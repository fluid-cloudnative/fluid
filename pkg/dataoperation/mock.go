package dataoperation

import (
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func BuildMockDataloadOperationReconcilerInterface(expectType datav1alpha1.OperationType, ttlSecondsAfterFinished *int32) (operation OperationReconcilerInterface) {

	return &mockDataloadOperationReconciler{
		expectType:              expectType,
		TTLSecondsAfterFinished: ttlSecondsAfterFinished,
	}
}

type mockDataloadOperationReconciler struct {
	expectType              datav1alpha1.OperationType
	TTLSecondsAfterFinished *int32
}

// GetChartsDirectory implements OperationReconcilerInterface.
func (mockDataloadOperationReconciler) GetChartsDirectory() string {
	panic("unimplemented")
}

// GetOperationType implements OperationReconcilerInterface.
func (m mockDataloadOperationReconciler) GetOperationType() datav1alpha1.OperationType {
	return datav1alpha1.DataLoadType
}

// GetReleaseNameSpacedName implements OperationReconcilerInterface.
func (mockDataloadOperationReconciler) GetReleaseNameSpacedName(object client.Object) types.NamespacedName {
	panic("unimplemented")
}

// GetStatusHandler implements OperationReconcilerInterface.
func (mockDataloadOperationReconciler) GetStatusHandler(object client.Object) StatusHandler {
	panic("unimplemented")
}

// GetTTL implements OperationReconcilerInterface.
func (m mockDataloadOperationReconciler) GetTTL(object client.Object) (ttl *int32, err error) {
	if m.expectType != datav1alpha1.DataLoadType {
		err = fmt.Errorf("the dataoperation type is %s, not DataloadType", m.expectType)
	}
	return m.TTLSecondsAfterFinished, err
}

// GetTargetDataset implements OperationReconcilerInterface.
func (m mockDataloadOperationReconciler) GetTargetDataset(object client.Object) (*datav1alpha1.Dataset, error) {
	panic("unimplemented")
}

// RemoveTargetDatasetStatusInProgress implements OperationReconcilerInterface.
func (mockDataloadOperationReconciler) RemoveTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	panic("unimplemented")
}

// SetTargetDatasetStatusInProgress implements OperationReconcilerInterface.
func (mockDataloadOperationReconciler) SetTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	panic("unimplemented")
}

// UpdateOperationApiStatus implements OperationReconcilerInterface.
func (mockDataloadOperationReconciler) UpdateOperationApiStatus(object client.Object, opStatus *datav1alpha1.OperationStatus) error {
	panic("unimplemented")
}

// UpdateStatusInfoForCompleted implements OperationReconcilerInterface.
func (mockDataloadOperationReconciler) UpdateStatusInfoForCompleted(object client.Object, infos map[string]string) error {
	panic("unimplemented")
}

// Validate implements OperationReconcilerInterface.
func (mockDataloadOperationReconciler) Validate(ctx runtime.ReconcileRequestContext, object client.Object) ([]datav1alpha1.Condition, error) {
	panic("unimplemented")
}
