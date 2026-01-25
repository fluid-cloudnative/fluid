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

package juicefs

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

// mockOperation is a minimal mock implementation of the dataoperation.OperationInterface
// used to test GetDataOperationValueFile error paths. Most methods return zero/nil values
// since only GetOperationType and GetOperationObject are exercised in these tests.
type mockOperation struct {
	opType dataoperation.OperationType
	object client.Object
}

func (m *mockOperation) HasPrecedingOperation() bool {
	return false
}

func (m *mockOperation) GetOperationObject() client.Object {
	return m.object
}

func (m *mockOperation) GetPossibleTargetDatasetNamespacedNames() []types.NamespacedName {
	return nil
}

func (m *mockOperation) GetTargetDataset() (*datav1alpha1.Dataset, error) {
	return nil, nil
}

func (m *mockOperation) GetReleaseNameSpacedName() types.NamespacedName {
	return types.NamespacedName{}
}

func (m *mockOperation) GetChartsDirectory() string {
	return ""
}

func (m *mockOperation) GetOperationType() dataoperation.OperationType {
	return m.opType
}

func (m *mockOperation) UpdateOperationApiStatus(opStatus *datav1alpha1.OperationStatus) error {
	return nil
}

func (m *mockOperation) Validate(ctx cruntime.ReconcileRequestContext) ([]datav1alpha1.Condition, error) {
	return nil, nil
}

func (m *mockOperation) UpdateStatusInfoForCompleted(infos map[string]string) error {
	return nil
}

func (m *mockOperation) SetTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	// No-op for test mock
}

func (m *mockOperation) RemoveTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	// No-op for test mock
}

func (m *mockOperation) GetStatusHandler() dataoperation.StatusHandler {
	return nil
}

func (m *mockOperation) GetTTL() (ttl *int32, err error) {
	return nil, nil
}

func (m *mockOperation) GetParallelTaskNumber() int32 {
	return 1
}

var _ = Describe("GetDataOperationValueFile", func() {
	var (
		engine     *JuiceFSEngine
		fakeClient client.Client
		ctx        cruntime.ReconcileRequestContext
	)

	BeforeEach(func() {
		fakeClient = fake.NewFakeClientWithScheme(testScheme)
		engine = &JuiceFSEngine{
			name:      "test",
			namespace: "default",
			Client:    fakeClient,
			Log:       fake.NullLogger(),
		}
		ctx = cruntime.ReconcileRequestContext{
			Client: fakeClient,
		}
	})

	DescribeTable("should return error for various operation types",
		func(opType dataoperation.OperationType, object client.Object) {
			op := &mockOperation{
				opType: opType,
				object: object,
			}
			_, err := engine.GetDataOperationValueFile(ctx, op)
			Expect(err).To(HaveOccurred())
		},
		Entry("DataMigrate - missing target dataset",
			dataoperation.DataMigrateType,
			&datav1alpha1.DataMigrate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-migrate",
					Namespace: "default",
				},
			},
		),
		Entry("DataLoad - missing target dataset",
			dataoperation.DataLoadType,
			&datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-load",
					Namespace: "default",
				},
			},
		),
		Entry("DataProcess - missing target dataset",
			dataoperation.DataProcessType,
			&datav1alpha1.DataProcess{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-process",
					Namespace: "default",
				},
			},
		),
		Entry("DataBackup - unsupported operation type",
			dataoperation.DataBackupType,
			&datav1alpha1.DataBackup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-backup",
					Namespace: "default",
				},
			},
		),
	)
})
