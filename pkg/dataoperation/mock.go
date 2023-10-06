package dataoperation

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func BuildMockOperationReconcilerInterface(operationType datav1alpha1.OperationType) (operation OperationReconcilerInterface) {

	return &mockOperationReconciler{
		operationType: operationType,
	}
}

type mockOperationReconciler struct {
	operationType datav1alpha1.OperationType
}

// GetChartsDirectory implements OperationReconcilerInterface.
func (mockOperationReconciler) GetChartsDirectory() string {
	panic("unimplemented")
}

// GetOperationType implements OperationReconcilerInterface.
func (m mockOperationReconciler) GetOperationType() datav1alpha1.OperationType {
	return m.operationType
}

// GetReleaseNameSpacedName implements OperationReconcilerInterface.
func (mockOperationReconciler) GetReleaseNameSpacedName(object client.Object) types.NamespacedName {
	panic("unimplemented")
}

// GetStatusHandler implements OperationReconcilerInterface.
func (mockOperationReconciler) GetStatusHandler(object client.Object) StatusHandler {
	panic("unimplemented")
}

// GetTTL implements OperationReconcilerInterface.
func (mockOperationReconciler) GetTTL(object client.Object) (ttl *int32, err error) {
	panic("unimplemented")
}

// GetTargetDataset implements OperationReconcilerInterface.
func (mockOperationReconciler) GetTargetDataset(object client.Object) (*datav1alpha1.Dataset, error) {
	panic("unimplemented")
}

// RemoveTargetDatasetStatusInProgress implements OperationReconcilerInterface.
func (mockOperationReconciler) RemoveTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	panic("unimplemented")
}

// SetTargetDatasetStatusInProgress implements OperationReconcilerInterface.
func (mockOperationReconciler) SetTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	panic("unimplemented")
}

// UpdateOperationApiStatus implements OperationReconcilerInterface.
func (mockOperationReconciler) UpdateOperationApiStatus(object client.Object, opStatus *datav1alpha1.OperationStatus) error {
	panic("unimplemented")
}

// UpdateStatusInfoForCompleted implements OperationReconcilerInterface.
func (mockOperationReconciler) UpdateStatusInfoForCompleted(object client.Object, infos map[string]string) error {
	panic("unimplemented")
}

// Validate implements OperationReconcilerInterface.
func (mockOperationReconciler) Validate(ctx runtime.ReconcileRequestContext, object client.Object) ([]datav1alpha1.Condition, error) {
	panic("unimplemented")
}
