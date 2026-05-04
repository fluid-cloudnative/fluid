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

package mock

import (
	"errors"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	fluidcommon "github.com/fluid-cloudnative/fluid/pkg/common"
	dataoperation "github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	flruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("generated dataoperation mocks", func() {
	var ctrl *gomock.Controller

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should build operation interfaces from the builder mock", func() {
		builder := NewMockOperationInterfaceBuilder(ctrl)
		operation := NewMockOperationInterface(ctrl)

		builder.EXPECT().Build(gomock.Nil()).Return(operation, nil)

		built, err := builder.Build(nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(built).To(Equal(operation))
	})

	It("should delegate getters through gomock", func() {
		operation := NewMockOperationInterface(ctrl)
		ttl := int32(42)
		dataset := &datav1alpha1.Dataset{}
		namespacedNames := []types.NamespacedName{{Namespace: "fluid", Name: "dataset"}}
		releaseName := types.NamespacedName{Namespace: "fluid", Name: "release"}
		statusHandler := NewMockStatusHandler(ctrl)

		operation.EXPECT().GetChartsDirectory().Return("/charts")
		operation.EXPECT().GetOperationObject().Return(nil)
		operation.EXPECT().GetOperationType().Return(dataoperation.DataProcessType)
		operation.EXPECT().GetParallelTaskNumber().Return(int32(3))
		operation.EXPECT().GetPossibleTargetDatasetNamespacedNames().Return(namespacedNames)
		operation.EXPECT().GetReleaseNameSpacedName().Return(releaseName)
		operation.EXPECT().GetStatusHandler().Return(statusHandler)
		operation.EXPECT().GetTTL().Return(&ttl, nil)
		operation.EXPECT().GetTargetDataset().Return(dataset, nil)

		Expect(operation.GetChartsDirectory()).To(Equal("/charts"))
		Expect(operation.GetOperationObject()).To(BeNil())
		Expect(operation.GetOperationType()).To(Equal(dataoperation.DataProcessType))
		Expect(operation.GetParallelTaskNumber()).To(Equal(int32(3)))
		Expect(operation.GetPossibleTargetDatasetNamespacedNames()).To(Equal(namespacedNames))
		Expect(operation.GetReleaseNameSpacedName()).To(Equal(releaseName))
		Expect(operation.GetStatusHandler()).To(Equal(statusHandler))
		retrievedTTL, err := operation.GetTTL()
		Expect(err).NotTo(HaveOccurred())
		Expect(retrievedTTL).To(Equal(&ttl))
		retrievedDataset, err := operation.GetTargetDataset()
		Expect(err).NotTo(HaveOccurred())
		Expect(retrievedDataset).To(Equal(dataset))
	})

	It("should delegate preceding operation checks through gomock", func() {
		operation := NewMockOperationInterface(ctrl)

		operation.EXPECT().HasPrecedingOperation().Return(true)

		Expect(operation.HasPrecedingOperation()).To(BeTrue())
	})

	It("should delegate dataset progress status updates through gomock", func() {
		operation := NewMockOperationInterface(ctrl)
		dataset := &datav1alpha1.Dataset{}

		operation.EXPECT().RemoveTargetDatasetStatusInProgress(dataset)
		operation.EXPECT().SetTargetDatasetStatusInProgress(dataset)

		operation.RemoveTargetDatasetStatusInProgress(dataset)
		operation.SetTargetDatasetStatusInProgress(dataset)
	})

	It("should delegate operation status updates through gomock", func() {
		operation := NewMockOperationInterface(ctrl)
		opStatus := &datav1alpha1.OperationStatus{}
		statusErr := errors.New("status update failed")

		operation.EXPECT().UpdateOperationApiStatus(opStatus).Return(statusErr)

		Expect(operation.UpdateOperationApiStatus(opStatus)).To(MatchError(statusErr))
	})

	It("should delegate completion status updates through gomock", func() {
		operation := NewMockOperationInterface(ctrl)
		completionErr := errors.New("completion update failed")
		completionInfo := map[string]string{"result": "done"}

		operation.EXPECT().UpdateStatusInfoForCompleted(completionInfo).Return(completionErr)

		Expect(operation.UpdateStatusInfoForCompleted(completionInfo)).To(MatchError(completionErr))
	})

	It("should delegate validation through gomock", func() {
		operation := NewMockOperationInterface(ctrl)
		conditions := []datav1alpha1.Condition{{Type: fluidcommon.Complete}}
		validateErr := errors.New("validate failed")
		ctx := flruntime.ReconcileRequestContext{}

		operation.EXPECT().Validate(gomock.Any()).Return(conditions, validateErr)

		retrievedConditions, err := operation.Validate(ctx)
		Expect(err).To(MatchError(validateErr))
		Expect(retrievedConditions).To(Equal(conditions))
	})

	It("should delegate status handling through gomock", func() {
		statusHandler := NewMockStatusHandler(ctrl)
		opStatus := &datav1alpha1.OperationStatus{}
		updatedStatus := &datav1alpha1.OperationStatus{Duration: "5s"}
		statusErr := errors.New("status failed")

		statusHandler.EXPECT().GetOperationStatus(flruntime.ReconcileRequestContext{}, opStatus).Return(updatedStatus, statusErr)

		result, err := statusHandler.GetOperationStatus(flruntime.ReconcileRequestContext{}, opStatus)

		Expect(err).To(MatchError(statusErr))
		Expect(result).To(Equal(updatedStatus))
	})
})
