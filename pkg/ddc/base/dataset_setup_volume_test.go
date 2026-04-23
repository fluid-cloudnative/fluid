/*
Copyright 2026 The Fluid Authors.

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
	"errors"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	testDatasetName      = "spark"
	testDatasetNamespace = "fluid"
	testDatasetUID       = "dataset-uid"
)

type stubImplement struct {
	shouldSetupMasterFn           func() (bool, error)
	setupMasterFn                 func() error
	checkMasterReadyFn            func() (bool, error)
	shouldCheckUFSFn              func() (bool, error)
	prepareUFSFn                  func() error
	shouldSetupWorkersFn          func() (bool, error)
	setupWorkersFn                func() error
	checkWorkersReadyFn           func() (bool, error)
	checkAndUpdateRuntimeStatusFn func() (bool, error)
	bindToDatasetFn               func() error
	createVolumeFn                func(context.Context) error
	deleteVolumeFn                func(context.Context) error
}

func (s *stubImplement) GetDataOperationValueFile(cruntime.ReconcileRequestContext, dataoperation.OperationInterface) (string, error) {
	return "", nil
}

func (s *stubImplement) CheckMasterReady() (bool, error) {
	if s.checkMasterReadyFn != nil {
		return s.checkMasterReadyFn()
	}
	return false, nil
}

func (s *stubImplement) CheckWorkersReady() (bool, error) {
	if s.checkWorkersReadyFn != nil {
		return s.checkWorkersReadyFn()
	}
	return false, nil
}

func (s *stubImplement) ShouldSetupMaster() (bool, error) {
	if s.shouldSetupMasterFn != nil {
		return s.shouldSetupMasterFn()
	}
	return false, nil
}

func (s *stubImplement) ShouldSetupWorkers() (bool, error) {
	if s.shouldSetupWorkersFn != nil {
		return s.shouldSetupWorkersFn()
	}
	return false, nil
}

func (s *stubImplement) ShouldCheckUFS() (bool, error) {
	if s.shouldCheckUFSFn != nil {
		return s.shouldCheckUFSFn()
	}
	return false, nil
}

func (s *stubImplement) SetupMaster() error {
	if s.setupMasterFn != nil {
		return s.setupMasterFn()
	}
	return nil
}

func (s *stubImplement) SetupWorkers() error {
	if s.setupWorkersFn != nil {
		return s.setupWorkersFn()
	}
	return nil
}

func (s *stubImplement) UpdateDatasetStatus(datav1alpha1.DatasetPhase) error { return nil }

func (s *stubImplement) PrepareUFS() error {
	if s.prepareUFSFn != nil {
		return s.prepareUFSFn()
	}
	return nil
}

func (s *stubImplement) ShouldSyncDatasetMounts() (bool, error) { return false, nil }

func (s *stubImplement) SyncDatasetMounts() error { return nil }

func (s *stubImplement) ShouldUpdateUFS() *utils.UFSToUpdate { return nil }

func (s *stubImplement) UpdateOnUFSChange(*utils.UFSToUpdate) (bool, error) { return false, nil }

func (s *stubImplement) Shutdown() error { return nil }

func (s *stubImplement) CheckRuntimeHealthy() error { return nil }

func (s *stubImplement) UpdateCacheOfDataset() error { return nil }

func (s *stubImplement) CheckAndUpdateRuntimeStatus() (bool, error) {
	if s.checkAndUpdateRuntimeStatusFn != nil {
		return s.checkAndUpdateRuntimeStatusFn()
	}
	return false, nil
}

func (s *stubImplement) CreateVolume(ctx context.Context) error {
	if s.createVolumeFn != nil {
		return s.createVolumeFn(ctx)
	}
	return nil
}

func (s *stubImplement) SyncReplicas(cruntime.ReconcileRequestContext) error { return nil }

func (s *stubImplement) SyncMetadata() error { return nil }

func (s *stubImplement) DeleteVolume(ctx context.Context) error {
	if s.deleteVolumeFn != nil {
		return s.deleteVolumeFn(ctx)
	}
	return nil
}

func (s *stubImplement) BindToDataset() error {
	if s.bindToDatasetFn != nil {
		return s.bindToDatasetFn()
	}
	return nil
}

func (s *stubImplement) CheckRuntimeReady() bool { return false }

func (s *stubImplement) SyncRuntime(cruntime.ReconcileRequestContext) (bool, error) {
	return false, nil
}

func (s *stubImplement) SyncScheduleInfoToCacheNodes() error { return nil }

func (s *stubImplement) Validate(cruntime.ReconcileRequestContext) error { return nil }

func (s *stubImplement) UsedStorageBytes() (int64, error) { return 0, nil }

func (s *stubImplement) FreeStorageBytes() (int64, error) { return 0, nil }

func (s *stubImplement) TotalStorageBytes() (int64, error) { return 0, nil }

func (s *stubImplement) TotalFileNums() (int64, error) { return 0, nil }

func (s *stubImplement) Operate(cruntime.ReconcileRequestContext, *datav1alpha1.OperationStatus, dataoperation.OperationInterface) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func newTemplateEngineForTest(impl Implement) (*TemplateEngine, cruntime.ReconcileRequestContext) {
	ctx := cruntime.ReconcileRequestContext{
		Context: context.Background(),
		NamespacedName: types.NamespacedName{
			Namespace: testDatasetNamespace,
			Name:      testDatasetName,
		},
		Log:         fake.NullLogger(),
		RuntimeType: "alluxio",
		Runtime: &datav1alpha1.AlluxioRuntime{
			TypeMeta: metav1.TypeMeta{Kind: "AlluxioRuntime"},
		},
	}

	return NewTemplateEngine(impl, "test-engine", ctx), ctx
}

var _ = Describe("Dataset helpers", func() {
})

var _ = Describe("TemplateEngine setup", func() {
	var (
		impl   *stubImplement
		engine *TemplateEngine
		ctx    cruntime.ReconcileRequestContext
	)

	BeforeEach(func() {
		impl = &stubImplement{}
		engine, ctx = newTemplateEngineForTest(impl)
	})

	It("returns ready when master, workers, and runtime are all ready", func() {
		impl.shouldSetupMasterFn = func() (bool, error) { return false, nil }
		impl.checkMasterReadyFn = func() (bool, error) { return true, nil }
		impl.shouldCheckUFSFn = func() (bool, error) { return false, nil }
		impl.shouldSetupWorkersFn = func() (bool, error) { return false, nil }
		impl.checkWorkersReadyFn = func() (bool, error) { return true, nil }
		impl.checkAndUpdateRuntimeStatusFn = func() (bool, error) { return true, nil }
		impl.bindToDatasetFn = func() error { return nil }

		ready, err := engine.Setup(ctx)

		Expect(err).NotTo(HaveOccurred())
		Expect(ready).To(BeTrue())
	})

	It("short-circuits when the master is not ready yet", func() {
		impl.shouldSetupMasterFn = func() (bool, error) { return false, nil }
		impl.checkMasterReadyFn = func() (bool, error) { return false, nil }

		ready, err := engine.Setup(ctx)

		Expect(err).NotTo(HaveOccurred())
		Expect(ready).To(BeFalse())
	})

	It("short-circuits when the workers are not ready yet", func() {
		impl.shouldSetupMasterFn = func() (bool, error) { return false, nil }
		impl.checkMasterReadyFn = func() (bool, error) { return true, nil }
		impl.shouldCheckUFSFn = func() (bool, error) { return false, nil }
		impl.shouldSetupWorkersFn = func() (bool, error) { return false, nil }
		impl.checkWorkersReadyFn = func() (bool, error) { return false, nil }

		ready, err := engine.Setup(ctx)

		Expect(err).NotTo(HaveOccurred())
		Expect(ready).To(BeFalse())
	})

	It("returns not ready when runtime status update fails", func() {
		expectedErr := errors.New("runtime status failed")
		impl.shouldSetupMasterFn = func() (bool, error) { return false, nil }
		impl.checkMasterReadyFn = func() (bool, error) { return true, nil }
		impl.shouldCheckUFSFn = func() (bool, error) { return false, nil }
		impl.shouldSetupWorkersFn = func() (bool, error) { return false, nil }
		impl.checkWorkersReadyFn = func() (bool, error) { return true, nil }
		impl.checkAndUpdateRuntimeStatusFn = func() (bool, error) { return true, expectedErr }

		ready, err := engine.Setup(ctx)

		Expect(err).To(MatchError(expectedErr))
		Expect(ready).To(BeFalse())
	})

	It("returns not ready when binding the dataset fails", func() {
		expectedErr := errors.New("bind failed")
		impl.shouldSetupMasterFn = func() (bool, error) { return false, nil }
		impl.checkMasterReadyFn = func() (bool, error) { return true, nil }
		impl.shouldCheckUFSFn = func() (bool, error) { return false, nil }
		impl.shouldSetupWorkersFn = func() (bool, error) { return false, nil }
		impl.checkWorkersReadyFn = func() (bool, error) { return true, nil }
		impl.checkAndUpdateRuntimeStatusFn = func() (bool, error) { return true, nil }
		impl.bindToDatasetFn = func() error { return expectedErr }

		ready, err := engine.Setup(ctx)

		Expect(err).To(MatchError(expectedErr))
		Expect(ready).To(BeFalse())
	})
})

var _ = Describe("TemplateEngine volume operations", func() {
	var (
		impl   *stubImplement
		engine *TemplateEngine
	)

	BeforeEach(func() {
		impl = &stubImplement{}
		engine, _ = newTemplateEngineForTest(impl)
	})

	It("delegates CreateVolume to the implementation", func() {
		volumeCtx := context.WithValue(context.Background(), "volume", "create")
		called := false
		impl.createVolumeFn = func(ctx context.Context) error {
			called = true
			Expect(ctx).To(BeIdenticalTo(volumeCtx))
			return nil
		}

		Expect(engine.CreateVolume(volumeCtx)).To(Succeed())
		Expect(called).To(BeTrue())
	})

	It("propagates CreateVolume errors", func() {
		volumeCtx := context.WithValue(context.Background(), "volume", "create-error")
		expectedErr := errors.New("create failed")
		impl.createVolumeFn = func(ctx context.Context) error {
			Expect(ctx).To(BeIdenticalTo(volumeCtx))
			return expectedErr
		}

		Expect(engine.CreateVolume(volumeCtx)).To(MatchError(expectedErr))
	})

	It("delegates DeleteVolume to the implementation", func() {
		volumeCtx := context.WithValue(context.Background(), "volume", "delete")
		called := false
		impl.deleteVolumeFn = func(ctx context.Context) error {
			called = true
			Expect(ctx).To(BeIdenticalTo(volumeCtx))
			return nil
		}

		Expect(engine.DeleteVolume(volumeCtx)).To(Succeed())
		Expect(called).To(BeTrue())
	})

	It("propagates DeleteVolume errors", func() {
		volumeCtx := context.WithValue(context.Background(), "volume", "delete-error")
		expectedErr := errors.New("delete failed")
		impl.deleteVolumeFn = func(ctx context.Context) error {
			Expect(ctx).To(BeIdenticalTo(volumeCtx))
			return expectedErr
		}

		Expect(engine.DeleteVolume(volumeCtx)).To(MatchError(expectedErr))
	})
})
