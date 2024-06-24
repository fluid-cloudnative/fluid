/*
Copyright 2020 The Fluid Authors.

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
	"os"
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	enginemock "github.com/fluid-cloudnative/fluid/pkg/ddc/base/mock"
	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachineryRuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("TemplateEngine", func() {
	mockDatasetName := "fluid-data-set"
	mockNamespace := "default"

	fakeDataset := &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mockDatasetName,
			Namespace: mockNamespace,
		},
	}
	s := apimachineryRuntime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, fakeDataset)
	fakeClient := fake.NewFakeClientWithScheme(s, fakeDataset)

	var fakeCtx = runtime.ReconcileRequestContext{
		Context: context.Background(),
		NamespacedName: types.NamespacedName{
			Namespace: mockNamespace,
			Name:      mockDatasetName,
		},
		Client:        fakeClient,
		Log:           fake.NullLogger(),
		RuntimeType:   "test-runtime-type",
		FinalizerName: "test-finalizer-name",
		Runtime:       &datav1alpha1.AlluxioRuntime{},
	}
	var t *base.TemplateEngine

	var (
		impl *enginemock.MockImplement
		ctrl *gomock.Controller
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		impl = enginemock.NewMockImplement(ctrl)
		os.Setenv("FLUID_SYNC_RETRY_DURATION", "0s")
		t = base.NewTemplateEngine(impl, "default-test", fakeCtx)
	})

	// Check if all expectations have been met after each It
	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Setup", func() {
		Context("When everything is set up", func() {
			It("Should check all if checking setup failed", func() {
				gomock.InOrder(
					impl.EXPECT().ShouldSetupMaster().Return(false, nil).Times(1),
					impl.EXPECT().CheckMasterReady().Return(true, nil).Times(1),
					impl.EXPECT().ShouldCheckUFS().Return(false, nil).Times(1),
					impl.EXPECT().ShouldSetupWorkers().Return(false, nil).Times(1),
					impl.EXPECT().CheckWorkersReady().Return(true, nil).Times(1),
					impl.EXPECT().CheckAndUpdateRuntimeStatus().Return(true, nil).Times(1),
					impl.EXPECT().BindToDataset().Return(nil).Times(1),
				)

				Expect(t.Setup(fakeCtx)).Should(Equal(true))
			})
		})

		Context("When nothing is set up", func() {
			It("Should set all up successfully", func() {
				gomock.InOrder(
					impl.EXPECT().ShouldSetupMaster().Return(true, nil).Times(1),
					impl.EXPECT().SetupMaster().Return(nil).Times(1),
					impl.EXPECT().CheckMasterReady().Return(true, nil).Times(1),
					impl.EXPECT().ShouldCheckUFS().Return(true, nil).Times(1),
					impl.EXPECT().PrepareUFS().Return(nil).Times(1),
					impl.EXPECT().ShouldSetupWorkers().Return(true, nil).Times(1),
					impl.EXPECT().SetupWorkers().Return(nil).Times(1),
					impl.EXPECT().CheckWorkersReady().Return(true, nil).Times(1),
					impl.EXPECT().CheckAndUpdateRuntimeStatus().Return(true, nil).Times(1),
					impl.EXPECT().BindToDataset().Return(nil).Times(1),
				)

				Expect(t.Setup(fakeCtx)).Should(Equal(true))
			})
		})
	})

	Describe("Sync", func() {
		Context("When all mount points are synced", func() {
			It("Should sync successfully", func() {
				gomock.InOrder(
					impl.EXPECT().SyncMetadata().Return(nil).Times(1),
					// impl.EXPECT().CheckAndUpdateRuntimeStatus().Return(true, nil).Times(1),
					// impl.EXPECT().UpdateCacheOfDataset().Return(nil).Times(1),
					impl.EXPECT().SyncReplicas(gomock.Eq(fakeCtx)).Return(nil).Times(1),
					impl.EXPECT().SyncRuntime(gomock.Eq(fakeCtx)).Return(false, nil).Times(1),
					impl.EXPECT().CheckRuntimeHealthy().Return(nil).Times(1),
					impl.EXPECT().CheckAndUpdateRuntimeStatus().Return(true, nil).Times(1),
					impl.EXPECT().UpdateCacheOfDataset().Return(nil).Times(1),
					impl.EXPECT().ShouldUpdateUFS().Return(&utils.UFSToUpdate{}).Times(1),
					impl.EXPECT().SyncScheduleInfoToCacheNodes().Return(nil).Times(1),
				)

				Expect(t.Sync(fakeCtx)).To(BeNil())
			})
		})

		Context("When some mount points need to be synced", func() {
			It("All mount points should be synced successfully", func() {
				datasetWithNewMountPoints := &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								// newly added mount points
								Name: "spark",
							},
						},
					},
					Status: datav1alpha1.DatasetStatus{
						Mounts: []datav1alpha1.Mount{},
					},
				}
				ufsToUpdate := utils.NewUFSToUpdate(datasetWithNewMountPoints)
				ufsToUpdate.AnalyzePathsDelta()

				gomock.InOrder(
					impl.EXPECT().SyncMetadata().Return(nil).Times(1),
					// impl.EXPECT().CheckAndUpdateRuntimeStatus().Return(true, nil).Times(1),
					// impl.EXPECT().UpdateCacheOfDataset().Return(nil).Times(1),
					impl.EXPECT().SyncReplicas(gomock.Eq(fakeCtx)).Return(nil).Times(1),
					impl.EXPECT().SyncRuntime(gomock.Eq(fakeCtx)).Return(false, nil).Times(1),
					impl.EXPECT().CheckRuntimeHealthy().Return(nil).Times(1),
					impl.EXPECT().CheckAndUpdateRuntimeStatus().Return(true, nil).Times(1),
					impl.EXPECT().UpdateCacheOfDataset().Return(nil).Times(1),
					impl.EXPECT().ShouldUpdateUFS().Return(ufsToUpdate).Times(1),
					impl.EXPECT().UpdateOnUFSChange(ufsToUpdate).Times(1),
					impl.EXPECT().SyncScheduleInfoToCacheNodes().Return(nil).Times(1),
				)
				Expect(t.Sync(fakeCtx)).Should(BeNil())
			})
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

	Describe("CheckRuntimeReady", func() {
		It("Should check runtime is ready", func() {
			impl.EXPECT().CheckRuntimeReady().Return(true).Times(1)
			Expect(t.CheckRuntimeReady()).Should(Equal(true))
		})
	})

	// Describe("CheckExistenceOfPath", func() {
	// 	It("Should check path exists", func() {
	// 		impl.EXPECT().CheckExistenceOfPath(fakeDataLoad).Return(false, nil)
	// 		Expect(t.CheckExistenceOfPath(fakeDataLoad)).Should(Equal(false))
	// 	})
	// })
})

var (
	testScheme *apimachineryRuntime.Scheme
)

func TestNewTemplateEngine(t *testing.T) {
	testObjs := []apimachineryRuntime.Object{}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
	engine := &alluxio.AlluxioEngine{
		Client: client,
	}

	id := "test id"

	ctx := runtime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Client:      client,
		Log:         fake.NullLogger(),
		RuntimeType: "alluxio",
		Runtime:     &datav1alpha1.AlluxioRuntime{},
	}

	templateEngine := base.NewTemplateEngine(engine, id, ctx)
	if !reflect.DeepEqual(templateEngine.Implement, engine) && templateEngine.Id != id && !reflect.DeepEqual(templateEngine.Context, ctx) {
		t.Errorf("expected implement %v, get %v; expected id %s, get %s, expected context %v, get %v", engine, templateEngine.Implement, id, templateEngine.Id, ctx, templateEngine.Context)
	}
}

func TestID(t *testing.T) {
	testObjs := []apimachineryRuntime.Object{}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
	engine := &alluxio.AlluxioEngine{
		Client: client,
	}

	id := "test id"

	ctx := runtime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Client:      client,
		Log:         fake.NullLogger(),
		RuntimeType: "alluxio",
		Runtime:     &datav1alpha1.AlluxioRuntime{},
	}

	templateEngine := base.NewTemplateEngine(engine, id, ctx)
	if templateEngine.Id != templateEngine.ID() {
		t.Errorf("expected %s, get %s", templateEngine.Id, templateEngine.ID())
	}
}
