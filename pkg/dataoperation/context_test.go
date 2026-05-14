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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"

	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

var _ = Describe("ReconcileRequestContext", func() {
	It("should preserve embedded runtime and operation-specific fields", func() {
		context := ReconcileRequestContext{
			ReconcileRequestContext: cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{
					Namespace: "fluid-system",
					Name:      "sample-op",
				},
				FinalizerName: "fluid-runtime-finalizer",
			},
			OpStatus:            nil,
			DataOpFinalizerName: "fluid.io/finalizer",
		}

		Expect(context.NamespacedName).To(Equal(types.NamespacedName{Namespace: "fluid-system", Name: "sample-op"}))
		Expect(context.FinalizerName).To(Equal("fluid-runtime-finalizer"))
		Expect(context.OpStatus).To(BeNil())
		Expect(context.DataObject).To(BeNil())
		Expect(context.DataOpFinalizerName).To(Equal("fluid.io/finalizer"))
	})
})
