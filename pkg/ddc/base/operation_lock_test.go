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

package base

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachineryRuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("getDataOperationKey", func() {
	It("should return the object name", func() {
		obj := &datav1alpha1.DataBackup{
			ObjectMeta: metav1.ObjectMeta{Name: "my-backup", Namespace: "default"},
		}
		Expect(getDataOperationKey(obj)).To(Equal("my-backup"))
	})

	It("should return empty string for object with no name", func() {
		obj := &datav1alpha1.DataBackup{
			ObjectMeta: metav1.ObjectMeta{Name: "", Namespace: "default"},
		}
		Expect(getDataOperationKey(obj)).To(Equal(""))
	})

	It("should return empty string for nil object", func() {
		var obj *datav1alpha1.DataBackup = nil
		Expect(getDataOperationKey(obj)).To(Equal(""))
	})
})

// lockTestOperation is a minimal OperationInterface for operation_lock tests.
type lockTestOperation struct {
	parallelTasks int32
}

func (o *lockTestOperation) GetOperationType() dataoperation.OperationType {
	return dataoperation.DataLoadType
}
func (o *lockTestOperation) GetParallelTaskNumber() int32                                { return o.parallelTasks }
func (o *lockTestOperation) SetTargetDatasetStatusInProgress(_ *datav1alpha1.Dataset)    {}
func (o *lockTestOperation) RemoveTargetDatasetStatusInProgress(_ *datav1alpha1.Dataset) {}
func (o *lockTestOperation) GetOperationObject() client.Object {
	return &datav1alpha1.DataLoad{ObjectMeta: metav1.ObjectMeta{Name: "new-load", Namespace: "default"}}
}
func (o *lockTestOperation) GetTargetDataset() (*datav1alpha1.Dataset, error) { return nil, nil }
func (o *lockTestOperation) GetPossibleTargetDatasetNamespacedNames() []types.NamespacedName {
	return nil
}
func (o *lockTestOperation) HasPrecedingOperation() bool { return false }
func (o *lockTestOperation) GetReleaseNameSpacedName() types.NamespacedName {
	return types.NamespacedName{}
}
func (o *lockTestOperation) GetChartsDirectory() string                    { return "" }
func (o *lockTestOperation) GetStatusHandler() dataoperation.StatusHandler { return nil }
func (o *lockTestOperation) GetTTL() (*int32, error)                       { return nil, nil }
func (o *lockTestOperation) UpdateOperationApiStatus(_ *datav1alpha1.OperationStatus) error {
	return nil
}
func (o *lockTestOperation) UpdateStatusInfoForCompleted(_ map[string]string) error { return nil }
func (o *lockTestOperation) Validate(_ cruntime.ReconcileRequestContext) ([]datav1alpha1.Condition, error) {
	return nil, nil
}

func newLockTestClient(dataset *datav1alpha1.Dataset) client.Client {
	s := apimachineryRuntime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, dataset, &datav1alpha1.DatasetList{})
	return fake.NewFakeClientWithScheme(s, dataset)
}

var _ = Describe("updateDatasetDataOperation", func() {
	It("should return error when parallel limit is already reached", func() {
		dataset := &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{Name: "test-dataset", Namespace: "default"},
			Status: datav1alpha1.DatasetStatus{
				OperationRef: map[string]string{"DataLoad": "existing-load"},
			},
		}
		ctx := cruntime.ReconcileRequestContext{
			Client: newLockTestClient(dataset),
			Log:    fake.NullLogger(),
		}

		// parallelTasks=1, "existing-load" already registered — "new-load" must be blocked
		err := updateDatasetDataOperation(ctx, "test-dataset", "default", "DataLoad", "new-load", &lockTestOperation{parallelTasks: 1})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("reached the maximum number of parallel"))
	})

	It("should succeed when no operations are in progress", func() {
		dataset := &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{Name: "test-dataset", Namespace: "default"},
		}
		ctx := cruntime.ReconcileRequestContext{
			Client: newLockTestClient(dataset),
			Log:    fake.NullLogger(),
		}

		err := updateDatasetDataOperation(ctx, "test-dataset", "default", "DataLoad", "new-load", &lockTestOperation{parallelTasks: 1})
		Expect(err).NotTo(HaveOccurred())
	})
})
