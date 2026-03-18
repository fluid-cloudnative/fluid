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

package efc

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

var _ = Describe("RuntimeReconciler (EFC)", func() {

	Describe("ControllerName", func() {
		It("should return the constant controller name", func() {
			r := &RuntimeReconciler{}
			Expect(r.ControllerName()).To(Equal("EFCRuntimeController"))
		})
	})

	Describe("ManagedResource", func() {
		It("should return an EFCRuntime with correct TypeMeta", func() {
			r := &RuntimeReconciler{}
			obj := r.ManagedResource()
			efcRuntime, ok := obj.(*datav1alpha1.EFCRuntime)
			Expect(ok).To(BeTrue())
			Expect(efcRuntime.Kind).To(Equal(datav1alpha1.EFCRuntimeKind))
			Expect(efcRuntime.APIVersion).To(ContainSubstring(datav1alpha1.GroupVersion.Group))
		})
	})

	Describe("NewRuntimeReconciler", func() {
		It("should initialize reconciler with all required fields set", func() {
			s := runtime.NewScheme()
			fakeClient := fake.NewFakeClientWithScheme(s)
			log := ctrl.Log.WithName("test")
			recorder := record.NewFakeRecorder(10)

			r := NewRuntimeReconciler(fakeClient, log, s, recorder)
			Expect(r).NotTo(BeNil())
			Expect(r.Scheme).To(Equal(s))
			Expect(r.mutex).NotTo(BeNil())
			Expect(r.engines).NotTo(BeNil())
			Expect(r.RuntimeReconciler).NotTo(BeNil())
		})
	})
})
