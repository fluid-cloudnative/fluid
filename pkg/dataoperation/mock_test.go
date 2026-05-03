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

	It("should panic for unimplemented operation methods", func() {
		Expect(func() { operation.HasPrecedingOperation() }).To(PanicWith("unimplemented"))
		Expect(func() { operation.GetOperationObject() }).To(PanicWith("unimplemented"))
		Expect(func() { operation.GetChartsDirectory() }).To(PanicWith("unimplemented"))
		Expect(func() { operation.GetReleaseNameSpacedName() }).To(PanicWith("unimplemented"))
		Expect(func() { operation.GetStatusHandler() }).To(PanicWith("unimplemented"))
		Expect(func() { _, _ = operation.GetTargetDataset() }).To(PanicWith("unimplemented"))
		Expect(func() { operation.GetPossibleTargetDatasetNamespacedNames() }).To(PanicWith("unimplemented"))
		Expect(func() { operation.RemoveTargetDatasetStatusInProgress(&datav1alpha1.Dataset{}) }).To(PanicWith("unimplemented"))
		Expect(func() { operation.SetTargetDatasetStatusInProgress(&datav1alpha1.Dataset{}) }).To(PanicWith("unimplemented"))
		Expect(func() { _ = operation.UpdateOperationApiStatus(&datav1alpha1.OperationStatus{}) }).To(PanicWith("unimplemented"))
		Expect(func() { _ = operation.UpdateStatusInfoForCompleted(map[string]string{"k": "v"}) }).To(PanicWith("unimplemented"))
		Expect(func() { _, _ = operation.Validate(runtime.ReconcileRequestContext{}) }).To(PanicWith("unimplemented"))
	})
})
