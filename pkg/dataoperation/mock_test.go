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

package dataoperation

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("mockDataloadOperationReconciler", func() {
	var operation mockDataloadOperationReconciler

	BeforeEach(func() {
		operation = mockDataloadOperationReconciler{}
	})

	DescribeTable("should panic for unimplemented operation methods", func(call func()) {
		Expect(call).To(PanicWith("unimplemented"))
	},
		Entry("HasPrecedingOperation", func() { operation.HasPrecedingOperation() }),
		Entry("GetOperationObject", func() { operation.GetOperationObject() }),
		Entry("GetChartsDirectory", func() { operation.GetChartsDirectory() }),
		Entry("GetReleaseNameSpacedName", func() { operation.GetReleaseNameSpacedName() }),
		Entry("GetStatusHandler", func() { operation.GetStatusHandler() }),
		Entry("GetTargetDataset", func() { _, _ = operation.GetTargetDataset() }),
		Entry("GetPossibleTargetDatasetNamespacedNames", func() { operation.GetPossibleTargetDatasetNamespacedNames() }),
		Entry("RemoveTargetDatasetStatusInProgress", func() { operation.RemoveTargetDatasetStatusInProgress(&datav1alpha1.Dataset{}) }),
		Entry("SetTargetDatasetStatusInProgress", func() { operation.SetTargetDatasetStatusInProgress(&datav1alpha1.Dataset{}) }),
		Entry("UpdateOperationApiStatus", func() { _ = operation.UpdateOperationApiStatus(&datav1alpha1.OperationStatus{}) }),
		Entry("UpdateStatusInfoForCompleted", func() { _ = operation.UpdateStatusInfoForCompleted(map[string]string{"k": "v"}) }),
		Entry("Validate", func() { _, _ = operation.Validate(runtime.ReconcileRequestContext{}) }),
	)
})
