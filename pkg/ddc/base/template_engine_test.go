/*

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
	"context"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	enginemock "github.com/fluid-cloudnative/fluid/pkg/ddc/base/mock"
	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachineryRuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var _ = Describe("TemplateEngine", func() {
	mockDatasetName := "fluid-data-set"
	mockDatasetNamespace := "default"

	fakeDataset := &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mockDatasetName,
			Namespace: mockDatasetNamespace,
		},
	}
	s := apimachineryRuntime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, fakeDataset)
	fakeClient := fake.NewFakeClientWithScheme(s, fakeDataset)

	var fakeCtx = runtime.ReconcileRequestContext{
		Context: context.Background(),
		NamespacedName: types.NamespacedName{
			Namespace: mockDatasetNamespace,
			Name:      mockDatasetName,
		},
		Client:        fakeClient,
		Log:           log.NullLogger{},
		RuntimeType:   "test-runtime-type",
		FinalizerName: "test-finalizer-name",
		Runtime:       nil,
	}
	var t = base.TemplateEngine{
		Id:      "test-id",
		Log:     fakeCtx.Log,
		Context: fakeCtx,
		Client:  fakeClient,
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
			impl.EXPECT().ShouldUpdateUFS().Return(&utils.UFSToUpdate{}).Times(1)
			impl.EXPECT().UpdateOnUFSChange(&utils.UFSToUpdate{}).Return(true, nil).Times(1)
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
