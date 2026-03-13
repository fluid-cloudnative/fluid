/*
Copyright 2021 The Fluid Authors.

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

package juicefs

import (
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

var _ = Describe("JuiceFSRuntimeReconciler", func() {

	Describe("ControllerName", func() {
		It("should return the constant controller name", func() {
			r := &JuiceFSRuntimeReconciler{}
			Expect(r.ControllerName()).To(Equal(controllerName))
			Expect(r.ControllerName()).To(Equal("JuiceFSRuntimeController"))
		})
	})

	Describe("ManagedResource", func() {
		It("should return a JuiceFSRuntime with correct TypeMeta", func() {
			r := &JuiceFSRuntimeReconciler{}
			obj := r.ManagedResource()
			juicefsRuntime, ok := obj.(*datav1alpha1.JuiceFSRuntime)
			Expect(ok).To(BeTrue())
			Expect(juicefsRuntime.Kind).To(Equal(datav1alpha1.JuiceFSRuntimeKind))
			Expect(juicefsRuntime.APIVersion).To(ContainSubstring(datav1alpha1.GroupVersion.Group))
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

	Describe("NewCacheOption", func() {
		It("should return cache options with two ByObject entries", func() {
			opts := NewCacheOption()
			Expect(opts.ByObject).To(HaveLen(2))
		})

		It("should have label selectors for StatefulSet and DaemonSet keyed by type", func() {
			opts := NewCacheOption()
			var seenStatefulSet, seenDaemonSet bool
			for key, byObj := range opts.ByObject {
				Expect(byObj.Label).NotTo(BeNil())
				Expect(byObj.Label.String()).To(ContainSubstring(common.JuiceFSRuntime))
				switch key.(type) {
				case *appsv1.StatefulSet:
					seenStatefulSet = true
				case *appsv1.DaemonSet:
					seenDaemonSet = true
				}
			}
			Expect(seenStatefulSet).To(BeTrue(), "expected StatefulSet key in ByObject")
			Expect(seenDaemonSet).To(BeTrue(), "expected DaemonSet key in ByObject")
		})
	})
})
