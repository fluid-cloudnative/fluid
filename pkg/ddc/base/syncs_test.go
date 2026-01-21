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
	"errors"
	"os"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	enginemock "github.com/fluid-cloudnative/fluid/pkg/ddc/base/mock"
	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachineryRuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

var mockError = errors.New("mock error")

const (
	mockDatasetName       = "fluid-data-set"
	mockNamespace         = "default"
	shouldReturnErrorDesc = "should return error"
)

var _ = Describe("Sync Error Paths", func() {
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
		_ = os.Setenv("FLUID_SYNC_RETRY_DURATION", "0s")
		t = base.NewTemplateEngine(impl, "default-test", fakeCtx)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Sync", func() {
		Context("when SyncMetadata fails", func() {
			It(shouldReturnErrorDesc, func() {
				impl.EXPECT().SyncMetadata().Return(mockError).Times(1)

				err := t.Sync(fakeCtx)
				Expect(err).To(Equal(mockError))
			})
		})

		Context("when SyncReplicas fails", func() {
			It(shouldReturnErrorDesc, func() {
				gomock.InOrder(
					impl.EXPECT().SyncMetadata().Return(nil).Times(1),
					impl.EXPECT().SyncReplicas(gomock.Eq(fakeCtx)).Return(mockError).Times(1),
				)

				err := t.Sync(fakeCtx)
				Expect(err).To(Equal(mockError))
			})
		})

		Context("when SyncRuntime fails", func() {
			It(shouldReturnErrorDesc, func() {
				gomock.InOrder(
					impl.EXPECT().SyncMetadata().Return(nil).Times(1),
					impl.EXPECT().SyncReplicas(gomock.Eq(fakeCtx)).Return(nil).Times(1),
					impl.EXPECT().SyncRuntime(gomock.Eq(fakeCtx)).Return(false, mockError).Times(1),
				)

				err := t.Sync(fakeCtx)
				Expect(err).To(Equal(mockError))
			})
		})

		Context("when SyncRuntime returns updated true", func() {
			It("should return early without error", func() {
				gomock.InOrder(
					impl.EXPECT().SyncMetadata().Return(nil).Times(1),
					impl.EXPECT().SyncReplicas(gomock.Eq(fakeCtx)).Return(nil).Times(1),
					impl.EXPECT().SyncRuntime(gomock.Eq(fakeCtx)).Return(true, nil).Times(1),
				)

				err := t.Sync(fakeCtx)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when CheckRuntimeHealthy fails", func() {
			It(shouldReturnErrorDesc, func() {
				gomock.InOrder(
					impl.EXPECT().SyncMetadata().Return(nil).Times(1),
					impl.EXPECT().SyncReplicas(gomock.Eq(fakeCtx)).Return(nil).Times(1),
					impl.EXPECT().SyncRuntime(gomock.Eq(fakeCtx)).Return(false, nil).Times(1),
					impl.EXPECT().CheckRuntimeHealthy().Return(mockError).Times(1),
				)

				err := t.Sync(fakeCtx)
				Expect(err).To(Equal(mockError))
			})
		})

		Context("when CheckAndUpdateRuntimeStatus fails", func() {
			It(shouldReturnErrorDesc, func() {
				gomock.InOrder(
					impl.EXPECT().SyncMetadata().Return(nil).Times(1),
					impl.EXPECT().SyncReplicas(gomock.Eq(fakeCtx)).Return(nil).Times(1),
					impl.EXPECT().SyncRuntime(gomock.Eq(fakeCtx)).Return(false, nil).Times(1),
					impl.EXPECT().CheckRuntimeHealthy().Return(nil).Times(1),
					impl.EXPECT().CheckAndUpdateRuntimeStatus().Return(false, mockError).Times(1),
				)

				err := t.Sync(fakeCtx)
				Expect(err).To(Equal(mockError))
			})
		})

		Context("when UpdateCacheOfDataset fails", func() {
			It(shouldReturnErrorDesc, func() {
				gomock.InOrder(
					impl.EXPECT().SyncMetadata().Return(nil).Times(1),
					impl.EXPECT().SyncReplicas(gomock.Eq(fakeCtx)).Return(nil).Times(1),
					impl.EXPECT().SyncRuntime(gomock.Eq(fakeCtx)).Return(false, nil).Times(1),
					impl.EXPECT().CheckRuntimeHealthy().Return(nil).Times(1),
					impl.EXPECT().CheckAndUpdateRuntimeStatus().Return(true, nil).Times(1),
					impl.EXPECT().UpdateCacheOfDataset().Return(mockError).Times(1),
				)

				err := t.Sync(fakeCtx)
				Expect(err).To(Equal(mockError))
			})
		})

		Context("when UpdateOnUFSChange fails", func() {
			It(shouldReturnErrorDesc, func() {
				datasetWithNewMountPoints := &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
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
					impl.EXPECT().SyncReplicas(gomock.Eq(fakeCtx)).Return(nil).Times(1),
					impl.EXPECT().SyncRuntime(gomock.Eq(fakeCtx)).Return(false, nil).Times(1),
					impl.EXPECT().CheckRuntimeHealthy().Return(nil).Times(1),
					impl.EXPECT().CheckAndUpdateRuntimeStatus().Return(true, nil).Times(1),
					impl.EXPECT().UpdateCacheOfDataset().Return(nil).Times(1),
					impl.EXPECT().ShouldUpdateUFS().Return(ufsToUpdate).Times(1),
					impl.EXPECT().UpdateOnUFSChange(ufsToUpdate).Return(false, mockError).Times(1),
				)

				err := t.Sync(fakeCtx)
				Expect(err).To(Equal(mockError))
			})
		})

		Context("when ShouldUpdateUFS returns UFSToUpdate with ShouldUpdate false", func() {
			It("should skip UpdateOnUFSChange and succeed", func() {
				// Create a UFSToUpdate with no changes (ShouldUpdate returns false)
				datasetWithNoChanges := &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								Name: "spark",
							},
						},
					},
					Status: datav1alpha1.DatasetStatus{
						Mounts: []datav1alpha1.Mount{
							{
								Name: "spark",
							},
						},
					},
				}
				ufsToUpdate := utils.NewUFSToUpdate(datasetWithNoChanges)
				ufsToUpdate.AnalyzePathsDelta()

				gomock.InOrder(
					impl.EXPECT().SyncMetadata().Return(nil).Times(1),
					impl.EXPECT().SyncReplicas(gomock.Eq(fakeCtx)).Return(nil).Times(1),
					impl.EXPECT().SyncRuntime(gomock.Eq(fakeCtx)).Return(false, nil).Times(1),
					impl.EXPECT().CheckRuntimeHealthy().Return(nil).Times(1),
					impl.EXPECT().CheckAndUpdateRuntimeStatus().Return(true, nil).Times(1),
					impl.EXPECT().UpdateCacheOfDataset().Return(nil).Times(1),
					impl.EXPECT().ShouldUpdateUFS().Return(ufsToUpdate).Times(1),
					// UpdateOnUFSChange should NOT be called since ShouldUpdate returns false
					impl.EXPECT().SyncScheduleInfoToCacheNodes().Return(nil).Times(1),
				)

				err := t.Sync(fakeCtx)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when SyncScheduleInfoToCacheNodes fails", func() {
			It(shouldReturnErrorDesc, func() {
				gomock.InOrder(
					impl.EXPECT().SyncMetadata().Return(nil).Times(1),
					impl.EXPECT().SyncReplicas(gomock.Eq(fakeCtx)).Return(nil).Times(1),
					impl.EXPECT().SyncRuntime(gomock.Eq(fakeCtx)).Return(false, nil).Times(1),
					impl.EXPECT().CheckRuntimeHealthy().Return(nil).Times(1),
					impl.EXPECT().CheckAndUpdateRuntimeStatus().Return(true, nil).Times(1),
					impl.EXPECT().UpdateCacheOfDataset().Return(nil).Times(1),
					impl.EXPECT().ShouldUpdateUFS().Return(nil).Times(1),
					impl.EXPECT().SyncScheduleInfoToCacheNodes().Return(mockError).Times(1),
				)

				err := t.Sync(fakeCtx)
				Expect(err).To(Equal(mockError))
			})
		})
	})
})
