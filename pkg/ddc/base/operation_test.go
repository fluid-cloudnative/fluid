/*
Copyright 2024 The Fluid Authors.

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

package base_test

import (
	"errors"
	"os"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	enginemock "github.com/fluid-cloudnative/fluid/pkg/ddc/base/mock"
	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachineryRuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	mockDatasetName = "test-dataset"
	mockNamespace   = "default"
)

var _ = Describe("Operate", func() {

	var (
		fakeCtx   runtime.ReconcileRequestContext
		t         *base.TemplateEngine
		impl      *enginemock.MockImplement
		ctrl      *gomock.Controller
		operation *mockOperation
		opStatus  *datav1alpha1.OperationStatus
		oldSyncRetryDuration string
		syncRetryDurationSet bool
	)

	BeforeEach(func() {
		fakeDataset := &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      mockDatasetName,
				Namespace: mockNamespace,
			},
		}
		s := apimachineryRuntime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, fakeDataset)
		fakeClient := fake.NewFakeClientWithScheme(s, fakeDataset)

		fakeCtx = runtime.ReconcileRequestContext{
			NamespacedName: types.NamespacedName{
				Namespace: mockNamespace,
				Name:      mockDatasetName,
			},
			Client:      fakeClient,
			Log:         fake.NullLogger(),
			RuntimeType: "test-runtime",
			Recorder:    record.NewFakeRecorder(10),
		}

		ctrl = gomock.NewController(GinkgoT())
		impl = enginemock.NewMockImplement(ctrl)

		oldSyncRetryDuration, syncRetryDurationSet = os.LookupEnv("FLUID_SYNC_RETRY_DURATION")
		_ = os.Setenv("FLUID_SYNC_RETRY_DURATION", "0s")

		t = base.NewTemplateEngine(impl, "test-engine", fakeCtx)

		operation = newMockOperation()
		opStatus = &datav1alpha1.OperationStatus{}
	})

	AfterEach(func() {
		if syncRetryDurationSet {
			_ = os.Setenv("FLUID_SYNC_RETRY_DURATION", oldSyncRetryDuration)
		} else {
			_ = os.Unsetenv("FLUID_SYNC_RETRY_DURATION")
		}
		ctrl.Finish()
	})

	Describe("Operate phase routing", func() {
		Context("when phase is unknown", func() {
			It("should return no requeue", func() {
				opStatus.Phase = "unknown-phase"

				result, err := t.Operate(fakeCtx, opStatus, operation)

				Expect(err).NotTo(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())
			})
		})

		Context("when phase is None", func() {
			It("should transition to Pending on valid operation", func() {
				opStatus.Phase = common.PhaseNone
				operation.validateErr = nil
				operation.updateStatusErr = nil

				result, err := t.Operate(fakeCtx, opStatus, operation)

				Expect(err).NotTo(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())
				Expect(opStatus.Phase).To(Equal(common.PhasePending))
				Expect(opStatus.Duration).To(Equal("Unfinished"))
				Expect(opStatus.Conditions).To(BeEmpty())
			})

			It("should transition to Failed on validation error", func() {
				opStatus.Phase = common.PhaseNone
				operation.validateErr = errMockValidate
				operation.updateStatusErr = nil

				result, err := t.Operate(fakeCtx, opStatus, operation)

				Expect(err).NotTo(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())
				Expect(opStatus.Phase).To(Equal(common.PhaseFailed))
				Expect(opStatus.Conditions).NotTo(BeEmpty())
			})
		})

		Context("when phase is Pending", func() {
			It("should return no requeue when waiting for preceding operation", func() {
				opStatus.Phase = common.PhasePending
				waitingTrue := true
				opStatus.WaitingFor.OperationComplete = &waitingTrue

				result, err := t.Operate(fakeCtx, opStatus, operation)

				Expect(err).NotTo(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())
			})

		})
	})
})

// mockOperation implements dataoperation.OperationInterface for testing
type mockOperation struct {
	validateErr     error
	updateStatusErr error
	hasPreceding    bool
	parallelTasks   int32
}

func newMockOperation() *mockOperation {
	return &mockOperation{
		parallelTasks: 1,
	}
}

var errMockValidate = errors.New("mock validation error")

func (m *mockOperation) HasPrecedingOperation() bool {
	return m.hasPreceding
}

func (m *mockOperation) GetOperationObject() client.Object {
	return &datav1alpha1.DataLoad{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mockDatasetName,
			Namespace: "default",
		},
	}
}

func (m *mockOperation) GetPossibleTargetDatasetNamespacedNames() []types.NamespacedName {
	return []types.NamespacedName{{Name: mockDatasetName, Namespace: "default"}}
}

func (m *mockOperation) GetTargetDataset() (*datav1alpha1.Dataset, error) {
	return &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mockDatasetName,
			Namespace: "default",
		},
	}, nil
}

func (m *mockOperation) GetReleaseNameSpacedName() types.NamespacedName {
	return types.NamespacedName{Name: "test-release", Namespace: "default"}
}

func (m *mockOperation) GetChartsDirectory() string {
	return "/charts/dataload"
}

func (m *mockOperation) GetOperationType() dataoperation.OperationType {
	return dataoperation.DataLoadType
}

func (m *mockOperation) UpdateOperationApiStatus(opStatus *datav1alpha1.OperationStatus) error {
	return m.updateStatusErr
}

func (m *mockOperation) Validate(ctx runtime.ReconcileRequestContext) ([]datav1alpha1.Condition, error) {
	if m.validateErr != nil {
		return []datav1alpha1.Condition{{Type: "Error"}}, m.validateErr
	}
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
	return m.parallelTasks
}
