package base_test

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	enginemock "github.com/fluid-cloudnative/fluid/pkg/ddc/base/mock"
	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var _ = Describe("TemplateEngine", func() {
	var fakeCtx = runtime.ReconcileRequestContext{
		Log:           log.NullLogger{},
		RuntimeType:   "test-runtime-type",
		FinalizerName: "test-finalizer-name",
		Runtime:       nil,
	}
	var t = base.TemplateEngine{
		Id:      "test-id",
		Log:     fakeCtx.Log,
		Context: fakeCtx,
	}
	var impl *enginemock.MockImplement

	BeforeEach(func() {
		ctrl := gomock.NewController(GinkgoT())
		impl = enginemock.NewMockImplement(ctrl)
		t.Implement = impl
	})

	Describe("Setup", func() {
		Context("When everything is set up", func() {
			It("Should return immediately after checking setup", func() {
				impl.EXPECT().IsSetupDone().Return(true, nil).Times(1)
			})
			It("Should check all if checking setup failed", func() {
				impl.EXPECT().IsSetupDone().Return(false, nil).Times(1)
				impl.EXPECT().ShouldSetupMaster().Return(false, nil).Times(1)
				impl.EXPECT().CheckMasterReady().Return(true, nil).Times(1)
				impl.EXPECT().ShouldCheckUFS().Return(false, nil).Times(1)
				impl.EXPECT().ShouldSetupWorkers().Return(false, nil).Times(1)
				impl.EXPECT().CheckWorkersReady().Return(true, nil).Times(1)
				impl.EXPECT().CheckAndUpdateRuntimeStatus().Return(true, nil).Times(1)
				impl.EXPECT().UpdateDatasetStatus(gomock.Any()).Return(nil).Times(1)
				impl.EXPECT().BindToDataset().Return(nil).Times(1)
				Expect(t.Setup(fakeCtx)).Should(Equal(true))
			})
		})

		Context("When nothing is set up", func() {
			Context("When everything goes fine", func() {
				It("Should set all up successfully", func() {
					impl.EXPECT().IsSetupDone().Return(false, nil).Times(1)
					impl.EXPECT().ShouldSetupMaster().Return(true, nil).Times(1)
					impl.EXPECT().SetupMaster().Return(nil).Times(1)
					impl.EXPECT().CheckMasterReady().Return(true, nil).Times(1)
					impl.EXPECT().ShouldCheckUFS().Return(true, nil).Times(1)
					impl.EXPECT().PrepareUFS().Return(nil).Times(1)
					impl.EXPECT().ShouldSetupWorkers().Return(true, nil).Times(1)
					impl.EXPECT().SetupWorkers().Return(nil).Times(1)
					impl.EXPECT().CheckWorkersReady().Return(true, nil).Times(1)
					impl.EXPECT().CheckAndUpdateRuntimeStatus().Return(true, nil).Times(1)
					impl.EXPECT().UpdateDatasetStatus(gomock.Any()).Return(nil).Times(1)
					impl.EXPECT().BindToDataset().Return(nil).Times(1)
					Expect(t.Setup(fakeCtx)).Should(Equal(true))
				})
			})
		})
	})

	Describe("Sync", func() {
		It("Should sync successfully", func() {
			impl.EXPECT().SyncMetadata().Return(nil).Times(1)
			impl.EXPECT().CheckAndUpdateRuntimeStatus().Return(true, nil).Times(1)
			impl.EXPECT().UpdateCacheOfDataset().Return(nil).Times(1)
			impl.EXPECT().CheckRuntimeHealthy().Return(nil).Times(1)
			impl.EXPECT().SyncReplicas(gomock.Eq(fakeCtx)).Return(nil).Times(1)
			impl.EXPECT().CheckAndUpdateRuntimeStatus().Return(true, nil).Times(1)
			Expect(t.Sync(fakeCtx)).To(BeNil())
		})
	})

	Describe("CreateVolume", func() {
		It("Should create volume successfully", func() {
			impl.EXPECT().CreateVolume().Return(nil).Times(1)
			Expect(t.CreateVolume()).To(BeNil())
		})
	})

	Describe("DeleteVolume", func() {
		It("Should delete  volume successfully", func() {
			impl.EXPECT().DeleteVolume().Return(nil).Times(1)
			Expect(t.DeleteVolume()).To(BeNil())
		})
	})

	Describe("ID", func() {
		It("Should return correct id", func() {
			Expect(t.ID()).Should(Equal(t.Id))
		})
	})

	Describe("Shutdown", func() {
		It("Should shutdown successfully", func() {
			impl.EXPECT().Shutdown().Return(nil).Times(1)
			Expect(t.Shutdown()).To(BeNil())
		})
	})
})
