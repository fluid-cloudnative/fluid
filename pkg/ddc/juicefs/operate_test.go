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
	"reflect"

	"github.com/agiledragon/gomonkey/v2"
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

// mockOperation implements dataoperation.OperationInterface for testing.
// It provides minimal stub implementations to test GetDataOperationValueFile routing logic.
type mockOperation struct {
	operationType dataoperation.OperationType
	object        client.Object
}

var _ dataoperation.OperationInterface = (*mockOperation)(nil)

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
	return m.operationType
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
		engine  *JuiceFSEngine
		ctx     cruntime.ReconcileRequestContext
		patches *gomonkey.Patches
	)

	BeforeEach(func() {
		engine = &JuiceFSEngine{
			Log:       fake.NullLogger(),
			name:      "test-juicefs",
			namespace: "default",
		}
		ctx = cruntime.ReconcileRequestContext{
			Log: fake.NullLogger(),
		}
	})

	AfterEach(func() {
		if patches != nil {
			patches.Reset()
		}
	})

	Context("when operation type is unsupported", func() {
		It("should return NotSupported error for DataBackup type", func() {
			operation := &mockOperation{
				operationType: dataoperation.DataBackupType,
				object: &datav1alpha1.DataBackup{
					TypeMeta: metav1.TypeMeta{
						Kind:       "DataBackup",
						APIVersion: "data.fluid.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-backup",
						Namespace: "default",
					},
				},
			}

			valueFileName, err := engine.GetDataOperationValueFile(ctx, operation)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not supported"))
			Expect(valueFileName).To(BeEmpty())
		})

		It("should return NotSupported error for unknown operation type", func() {
			operation := &mockOperation{
				operationType: dataoperation.OperationType("UnknownType"),
				object: &datav1alpha1.DataBackup{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Unknown",
						APIVersion: "data.fluid.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-unknown",
						Namespace: "default",
					},
				},
			}

			valueFileName, err := engine.GetDataOperationValueFile(ctx, operation)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not supported"))
			Expect(valueFileName).To(BeEmpty())
		})
	})

	Context("when operation type is DataMigrate", func() {
		It("should call generateDataMigrateValueFile and return its result", func() {
			dataMigrate := &datav1alpha1.DataMigrate{
				TypeMeta: metav1.TypeMeta{
					Kind:       "DataMigrate",
					APIVersion: "data.fluid.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-migrate",
					Namespace: "default",
				},
			}

			operation := &mockOperation{
				operationType: dataoperation.DataMigrateType,
				object:        dataMigrate,
			}

			patches = gomonkey.ApplyPrivateMethod(reflect.TypeOf(engine), "generateDataMigrateValueFile",
				func(_ *JuiceFSEngine, _ cruntime.ReconcileRequestContext, _ client.Object) (string, error) {
					return "/tmp/test-migrate-values.yaml", nil
				})

			valueFileName, err := engine.GetDataOperationValueFile(ctx, operation)

			Expect(err).NotTo(HaveOccurred())
			Expect(valueFileName).To(Equal("/tmp/test-migrate-values.yaml"))
		})
	})

	Context("when operation type is DataLoad", func() {
		It("should call generateDataLoadValueFile and return its result", func() {
			dataLoad := &datav1alpha1.DataLoad{
				TypeMeta: metav1.TypeMeta{
					Kind:       "DataLoad",
					APIVersion: "data.fluid.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-load",
					Namespace: "default",
				},
			}

			operation := &mockOperation{
				operationType: dataoperation.DataLoadType,
				object:        dataLoad,
			}

			patches = gomonkey.ApplyPrivateMethod(reflect.TypeOf(engine), "generateDataLoadValueFile",
				func(_ *JuiceFSEngine, _ cruntime.ReconcileRequestContext, _ client.Object) (string, error) {
					return "/tmp/test-load-values.yaml", nil
				})

			valueFileName, err := engine.GetDataOperationValueFile(ctx, operation)

			Expect(err).NotTo(HaveOccurred())
			Expect(valueFileName).To(Equal("/tmp/test-load-values.yaml"))
		})
	})

	Context("when operation type is DataProcess", func() {
		It("should call generateDataProcessValueFile and return its result", func() {
			dataProcess := &datav1alpha1.DataProcess{
				TypeMeta: metav1.TypeMeta{
					Kind:       "DataProcess",
					APIVersion: "data.fluid.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-process",
					Namespace: "default",
				},
			}

			operation := &mockOperation{
				operationType: dataoperation.DataProcessType,
				object:        dataProcess,
			}

			patches = gomonkey.ApplyPrivateMethod(reflect.TypeOf(engine), "generateDataProcessValueFile",
				func(_ *JuiceFSEngine, _ cruntime.ReconcileRequestContext, _ client.Object) (string, error) {
					return "/tmp/test-process-values.yaml", nil
				})

			valueFileName, err := engine.GetDataOperationValueFile(ctx, operation)

			Expect(err).NotTo(HaveOccurred())
			Expect(valueFileName).To(Equal("/tmp/test-process-values.yaml"))
		})
	})
})
