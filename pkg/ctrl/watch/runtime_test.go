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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var _ = Describe("RuntimeEventHandler", func() {
	var (
		handler        runtimeEventHandler
		fakeReconciler *FakeRuntimeReconciler
	)

	BeforeEach(func() {
		handler = runtimeEventHandler{}
		fakeReconciler = &FakeRuntimeReconciler{}
	})

	Describe("onCreateFunc", func() {
		var (
			createRuntimeEvent event.CreateEvent
			f                  func(event.CreateEvent) bool
		)

		BeforeEach(func() {
			f = handler.onCreateFunc(fakeReconciler)
		})

		Context("when the Object is RuntimeInterface", func() {
			BeforeEach(func() {
				createRuntimeEvent = event.CreateEvent{
					Object: &datav1alpha1.JindoRuntime{},
				}
			})

			It("should return true to reconcile the event", func() {
				predicate := f(createRuntimeEvent)
				Expect(predicate).To(BeTrue())
			})
		})

		Context("when the Object is not RuntimeInterface", func() {
			BeforeEach(func() {
				createRuntimeEvent = event.CreateEvent{
					Object: &corev1.Pod{},
				}
			})

			It("should return false to skip the event", func() {
				predicate := f(createRuntimeEvent)
				Expect(predicate).To(BeFalse())
			})
		})
	})

	Describe("onUpdateFunc", func() {
		var (
			updateRuntimeEvent event.UpdateEvent
			f                  func(event.UpdateEvent) bool
		)

		BeforeEach(func() {
			f = handler.onUpdateFunc(fakeReconciler)
		})

		Context("when resource versions are different and objects are RuntimeInterface", func() {
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

			It("should return true to reconcile the event", func() {
				predicate := f(updateRuntimeEvent)
				Expect(predicate).To(BeTrue())
			})
		})

		Context("when resource versions are equal", func() {
			BeforeEach(func() {
				updateRuntimeEvent = event.UpdateEvent{
					ObjectOld: &datav1alpha1.JindoRuntime{
						ObjectMeta: metav1.ObjectMeta{
							ResourceVersion: "456",
						},
					},
					ObjectNew: &datav1alpha1.JindoRuntime{
						ObjectMeta: metav1.ObjectMeta{
							ResourceVersion: "456",
						},
					},
				}
			})

			It("should return false to skip the event", func() {
				predicate := f(updateRuntimeEvent)
				Expect(predicate).To(BeFalse())
			})
		})

		Context("when both objects are not RuntimeInterface", func() {
			BeforeEach(func() {
				updateRuntimeEvent = event.UpdateEvent{
					ObjectOld: &corev1.Pod{},
					ObjectNew: &corev1.Pod{},
				}
			})

			It("should return false to skip the event", func() {
				predicate := f(updateRuntimeEvent)
				Expect(predicate).To(BeFalse())
			})
		})

		Context("when old object is not RuntimeInterface but new object is", func() {
			BeforeEach(func() {
				updateRuntimeEvent = event.UpdateEvent{
					ObjectOld: &corev1.Pod{},
					ObjectNew: &datav1alpha1.JindoRuntime{},
				}
			})

			It("should return false to skip the event", func() {
				predicate := f(updateRuntimeEvent)
				Expect(predicate).To(BeFalse())
			})
		})
	})

	Describe("onDeleteFunc", func() {
		var (
			delRuntimeEvent event.DeleteEvent
			f               func(event.DeleteEvent) bool
		)

		BeforeEach(func() {
			f = handler.onDeleteFunc(fakeReconciler)
		})

		Context("when the Object is RuntimeInterface", func() {
			BeforeEach(func() {
				delRuntimeEvent = event.DeleteEvent{
					Object: &datav1alpha1.JindoRuntime{},
				}
			})

			It("should return true to reconcile the event", func() {
				predicate := f(delRuntimeEvent)
				Expect(predicate).To(BeTrue())
			})
		})

		Context("when the Object is not RuntimeInterface", func() {
			BeforeEach(func() {
				delRuntimeEvent = event.DeleteEvent{
					Object: &corev1.Pod{},
				}
			})

			It("should return false to skip the event", func() {
				predicate := f(delRuntimeEvent)
				Expect(predicate).To(BeFalse())
			})
		})
	})
})
