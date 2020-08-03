package base_test

import (
	"github.com/cloudnativefluid/fluid/pkg/ddc/base"
	enginemock "github.com/cloudnativefluid/fluid/pkg/ddc/base/mock"
	"github.com/cloudnativefluid/fluid/pkg/runtime"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var _ = Describe("TemplateEngine", func() {
	var fakeRuntime = runtime.ReconcileRequestContext{
		Log:           log.NullLogger{},
		RuntimeType:   "test-runtime-type",
		FinalizerName: "test-finalizer-name",
		Runtime:       nil,
	}
	var t = base.TemplateEngine{
		Id:      "test-id",
		Log:     fakeRuntime.Log,
		Context: fakeRuntime,
	}
	var im *enginemock.MockImplement

	BeforeEach(func() {
		ctrl := gomock.NewController(GinkgoT())
		im = enginemock.NewMockImplement(ctrl)
		t.Implement = im
	})

	Context("When everything is set up", func() {
		It("Should return immediately after checking setup", func() {
			im.EXPECT().IsSetupDone().Return(true, nil).Times(1)
		})
		It("Should check all if checking setup failed", func() {
			im.EXPECT().IsSetupDone().Return(false, nil).Times(1)
			im.EXPECT().ShouldSetupMaster().Return(false, nil).Times(1)
			im.EXPECT().CheckMasterReady().Return(true, nil).Times(1)
			im.EXPECT().ShouldCheckUFS().Return(false, nil).Times(1)
			im.EXPECT().ShouldSetupWorkers().Return(false, nil).Times(1)
			im.EXPECT().CheckWorkersReady().Return(true, nil).Times(1)
			im.EXPECT().CheckAndUpdateRuntimeStatus().Return(true, nil).Times(1)
			im.EXPECT().UpdateDatasetStatus(gomock.Any()).Return(nil).Times(1)
			Expect(t.Setup(fakeRuntime)).Should(Equal(true))
		})
	})

	Context("When nothing is set up", func() {
		Context("When everything goes fine", func() {
			It("Should set all up successfully", func() {
				im.EXPECT().IsSetupDone().Return(false, nil).Times(1)
				im.EXPECT().ShouldSetupMaster().Return(true, nil).Times(1)
				im.EXPECT().SetupMaster().Return(nil).Times(1)
				im.EXPECT().CheckMasterReady().Return(true, nil).Times(1)
				im.EXPECT().ShouldCheckUFS().Return(true, nil).Times(1)
				im.EXPECT().PrepareUFS().Return(nil).Times(1)
				im.EXPECT().ShouldSetupWorkers().Return(true, nil).Times(1)
				im.EXPECT().SetupWorkers().Return(nil).Times(1)
				im.EXPECT().CheckWorkersReady().Return(true, nil).Times(1)
				im.EXPECT().CheckAndUpdateRuntimeStatus().Return(true, nil).Times(1)
				im.EXPECT().UpdateDatasetStatus(gomock.Any()).Return(nil).Times(1)
				Expect(t.Setup(fakeRuntime)).Should(Equal(true))
			})
		})
	})
})
