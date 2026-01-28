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

package watch

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("runtimeEventHandler", func() {
	var (
		handler    *runtimeEventHandler
		reconciler *FakeRuntimeReconciler
	)

	BeforeEach(func() {
		handler = &runtimeEventHandler{}
		reconciler = &FakeRuntimeReconciler{}
	})

	Describe("onCreateFunc", func() {
		It("should reconcile if the object is RuntimeInterface", func() {
			createRuntimeEvent := event.CreateEvent{
				Object: &datav1alpha1.JindoRuntime{},
			}
			f := handler.onCreateFunc(reconciler)
			Expect(f(createRuntimeEvent)).To(BeTrue())
		})

		It("should not reconcile if the object is not RuntimeInterface", func() {
			createRuntimeEvent := event.CreateEvent{
				Object: &corev1.Pod{},
			}
			f := handler.onCreateFunc(reconciler)
			Expect(f(createRuntimeEvent)).To(BeFalse())
		})
	})

	Describe("onUpdateFunc", func() {
		var updateRuntimeEvent event.UpdateEvent

		BeforeEach(func() {
			updateRuntimeEvent = event.UpdateEvent{
				ObjectOld: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						ResourceVersion: "123",
					},
				},
				ObjectNew: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						ResourceVersion: "456",
					},
				},
			}
		})

		It("should reconcile if resource versions differ and both are RuntimeInterface", func() {
			f := handler.onUpdateFunc(reconciler)
			Expect(f(updateRuntimeEvent)).To(BeTrue())
		})

		It("should not reconcile if resource versions are equal", func() {
			updateRuntimeEvent.ObjectOld.(*datav1alpha1.JindoRuntime).SetResourceVersion("456")
			f := handler.onUpdateFunc(reconciler)
			Expect(f(updateRuntimeEvent)).To(BeFalse())
		})

		It("should not reconcile if both objects are not RuntimeInterface", func() {
			updateRuntimeEvent.ObjectOld = &corev1.Pod{}
			updateRuntimeEvent.ObjectNew = &corev1.Pod{}
			f := handler.onUpdateFunc(reconciler)
			Expect(f(updateRuntimeEvent)).To(BeFalse())
		})

		It("should not reconcile if only new object is RuntimeInterface", func() {
			updateRuntimeEvent.ObjectOld = &corev1.Pod{}
			updateRuntimeEvent.ObjectNew = &datav1alpha1.JindoRuntime{}
			f := handler.onUpdateFunc(reconciler)
			Expect(f(updateRuntimeEvent)).To(BeFalse())
		})
	})

	Describe("onDeleteFunc", func() {
		It("should reconcile if the object is RuntimeInterface", func() {
			delRuntimeEvent := event.DeleteEvent{
				Object: &datav1alpha1.JindoRuntime{},
			}
			f := handler.onDeleteFunc(reconciler)
			Expect(f(delRuntimeEvent)).To(BeTrue())
		})

		It("should not reconcile if the object is not RuntimeInterface", func() {
			delRuntimeEvent := event.DeleteEvent{
				Object: &corev1.Pod{},
			}
			f := handler.onDeleteFunc(reconciler)
			Expect(f(delRuntimeEvent)).To(BeFalse())
		})
	})
})
