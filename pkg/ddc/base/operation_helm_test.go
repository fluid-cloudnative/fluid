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

package base_test

import (
	"errors"

	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	opmock "github.com/fluid-cloudnative/fluid/pkg/dataoperation/mock"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	enginemock "github.com/fluid-cloudnative/fluid/pkg/ddc/base/mock"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

const (
	shouldReturnError = "should return error"
	valuesYamlFile    = "values.yaml"
)

var _ = Describe("InstallDataOperationHelmIfNotExist", func() {
	var (
		ctrl     *gomock.Controller
		mockImpl *enginemock.MockImplement
		mockOp   *opmock.MockOperationInterface
		fakeCtx  cruntime.ReconcileRequestContext
		patches  *gomonkey.Patches
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockImpl = enginemock.NewMockImplement(ctrl)
		mockOp = opmock.NewMockOperationInterface(ctrl)
		fakeCtx = cruntime.ReconcileRequestContext{
			Log:        fake.NullLogger(),
			EngineImpl: "test-engine",
		}
	})

	AfterEach(func() {
		ctrl.Finish()
		if patches != nil {
			patches.Reset()
		}
	})

	Context("when release already exists", func() {
		It("should return nil without installing", func() {
			mockOp.EXPECT().GetOperationType().Return(dataoperation.DataLoadType)
			mockOp.EXPECT().GetReleaseNameSpacedName().Return(types.NamespacedName{Name: "release", Namespace: "ns"})
			patches = gomonkey.ApplyFunc(helm.CheckRelease, func(name, namespace string) (bool, error) {
				return true, nil
			})
			patches.ApplyFunc(helm.InstallRelease, func(name, namespace, valueFile, chartName string) error {
				Fail("InstallRelease should not be called")
				return nil
			})

			err := base.InstallDataOperationHelmIfNotExist(fakeCtx, mockOp, mockImpl)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when CheckRelease fails", func() {
		It(shouldReturnError, func() {
			mockOp.EXPECT().GetOperationType().Return(dataoperation.DataLoadType)
			mockOp.EXPECT().GetReleaseNameSpacedName().Return(types.NamespacedName{Name: "release", Namespace: "ns"})
			patches = gomonkey.ApplyFunc(helm.CheckRelease, func(name, namespace string) (bool, error) {
				return false, errors.New("check failure")
			})

			err := base.InstallDataOperationHelmIfNotExist(fakeCtx, mockOp, mockImpl)
			Expect(err).To(MatchError("check failure"))
		})
	})

	Context("when value file generation fails", func() {
		It(shouldReturnError, func() {
			mockOp.EXPECT().GetOperationType().Return(dataoperation.DataLoadType)
			mockOp.EXPECT().GetReleaseNameSpacedName().Return(types.NamespacedName{Name: "release", Namespace: "ns"})
			patches = gomonkey.ApplyFunc(helm.CheckRelease, func(name, namespace string) (bool, error) {
				return false, nil
			})
			mockImpl.EXPECT().GetDataOperationValueFile(gomock.Any(), gomock.Eq(mockOp)).Return("", errors.New("gen failure"))

			err := base.InstallDataOperationHelmIfNotExist(fakeCtx, mockOp, mockImpl)
			Expect(err).To(MatchError("gen failure"))
		})
	})

	Context("when InstallRelease fails", func() {
		It(shouldReturnError, func() {
			mockOp.EXPECT().GetOperationType().Return(dataoperation.DataLoadType).Times(2)
			mockOp.EXPECT().GetReleaseNameSpacedName().Return(types.NamespacedName{Name: "release", Namespace: "ns"})
			mockOp.EXPECT().GetChartsDirectory().Return("/charts")
			patches = gomonkey.ApplyFunc(helm.CheckRelease, func(name, namespace string) (bool, error) {
				return false, nil
			})
			mockImpl.EXPECT().GetDataOperationValueFile(gomock.Any(), gomock.Eq(mockOp)).Return(valuesYamlFile, nil)
			patches.ApplyFunc(helm.InstallRelease, func(name, namespace, valueFile, chartName string) error {
				return errors.New("install failure")
			})

			err := base.InstallDataOperationHelmIfNotExist(fakeCtx, mockOp, mockImpl)
			Expect(err).To(MatchError("install failure"))
		})
	})

	Context("when installing successfully with DataLoad type", func() {
		It("should use engine chart", func() {
			mockOp.EXPECT().GetOperationType().Return(dataoperation.DataLoadType).Times(2)
			mockOp.EXPECT().GetReleaseNameSpacedName().Return(types.NamespacedName{Name: "release", Namespace: "ns"})
			mockOp.EXPECT().GetChartsDirectory().Return("/charts")
			patches = gomonkey.ApplyFunc(helm.CheckRelease, func(name, namespace string) (bool, error) {
				return false, nil
			})
			mockImpl.EXPECT().GetDataOperationValueFile(gomock.Any(), gomock.Eq(mockOp)).Return(valuesYamlFile, nil)
			patches.ApplyFunc(helm.InstallRelease, func(name, namespace, valueFile, chartName string) error {
				Expect(chartName).To(Equal("/charts/test-engine"))
				return nil
			})

			err := base.InstallDataOperationHelmIfNotExist(fakeCtx, mockOp, mockImpl)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when installing successfully with DataProcess type", func() {
		It("should use common chart", func() {
			mockOp.EXPECT().GetOperationType().Return(dataoperation.DataProcessType).Times(2)
			mockOp.EXPECT().GetReleaseNameSpacedName().Return(types.NamespacedName{Name: "release", Namespace: "ns"})
			mockOp.EXPECT().GetChartsDirectory().Return("/charts")
			patches = gomonkey.ApplyFunc(helm.CheckRelease, func(name, namespace string) (bool, error) {
				return false, nil
			})
			mockImpl.EXPECT().GetDataOperationValueFile(gomock.Any(), gomock.Eq(mockOp)).Return(valuesYamlFile, nil)
			patches.ApplyFunc(helm.InstallRelease, func(name, namespace, valueFile, chartName string) error {
				Expect(chartName).To(Equal("/charts/common"))
				return nil
			})

			err := base.InstallDataOperationHelmIfNotExist(fakeCtx, mockOp, mockImpl)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
