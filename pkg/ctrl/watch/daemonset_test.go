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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var _ = Describe("DaemonSetEventHandler", func() {
	var (
		handler               *daemonsetEventHandler
		fakeRuntimeReconciler *FakeRuntimeReconciler
	)

	BeforeEach(func() {
		handler = &daemonsetEventHandler{}
		fakeRuntimeReconciler = &FakeRuntimeReconciler{}
	})

	Describe("onCreateFunc", func() {
		var createFunc func(e event.CreateEvent) bool

		BeforeEach(func() {
			createFunc = handler.onCreateFunc(fakeRuntimeReconciler)
		})

		Context("when the object is a DaemonSet with RuntimeInterface owner", func() {
			It("should return true and reconcile the event", func() {
				createEvent := event.CreateEvent{
					Object: &appsv1.DaemonSet{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-daemonset",
							Namespace: "test-namespace",
							OwnerReferences: []metav1.OwnerReference{
								{
									Kind:       datav1alpha1.JindoRuntimeKind,
									APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
									Controller: ptr.To(true),
								},
							},
						},
					},
				}

				result := createFunc(createEvent)
				Expect(result).To(BeTrue(), "Expected the event to be reconciled")
			})
		})

		Context("when the object is not a DaemonSet", func() {
			It("should return false and skip the event", func() {
				createEvent := event.CreateEvent{
					Object: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-pod",
							Namespace: "test-namespace",
						},
					},
				}

				result := createFunc(createEvent)
				Expect(result).To(BeFalse(), "Expected the event to be skipped")
			})
		})

		Context("when the object has deletion timestamp set", func() {
			It("should return false and skip the event", func() {
				now := metav1.Now()
				createEvent := event.CreateEvent{
					Object: &appsv1.DaemonSet{
						ObjectMeta: metav1.ObjectMeta{
							Name:              "test-daemonset",
							Namespace:         "test-namespace",
							DeletionTimestamp: &now,
						},
					},
				}

				result := createFunc(createEvent)
				Expect(result).To(BeFalse(), "Expected the event to be skipped when deleting")
			})
		})

		Context("when the object is not managed by the controller", func() {
			It("should return false and skip the event", func() {
				createEvent := event.CreateEvent{
					Object: &appsv1.DaemonSet{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-daemonset",
							Namespace: "test-namespace",
							// No owner references - not managed
						},
					},
				}

				result := createFunc(createEvent)
				Expect(result).To(BeFalse(), "Expected the event to be skipped when not managed")
			})
		})
	})

	Describe("onUpdateFunc", func() {
		var updateFunc func(e event.UpdateEvent) bool

		BeforeEach(func() {
			updateFunc = handler.onUpdateFunc(fakeRuntimeReconciler)
		})

		Context("when both objects are valid DaemonSets with different resource versions", func() {
			It("should return true and reconcile the event", func() {
				updateEvent := event.UpdateEvent{
					ObjectOld: &appsv1.DaemonSet{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-daemonset",
							Namespace: "test-namespace",
							OwnerReferences: []metav1.OwnerReference{
								{
									Kind:       datav1alpha1.JindoRuntimeKind,
									APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
									Controller: ptr.To(true),
								},
							},
							ResourceVersion: "123",
						},
					},
					ObjectNew: &appsv1.DaemonSet{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-daemonset",
							Namespace: "test-namespace",
							OwnerReferences: []metav1.OwnerReference{
								{
									Kind:       datav1alpha1.JindoRuntimeKind,
									APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
									Controller: ptr.To(true),
								},
							},
							ResourceVersion: "456",
						},
					},
				}

				result := updateFunc(updateEvent)
				Expect(result).To(BeTrue(), "Expected the event to be reconciled")
			})
		})

		Context("when the new object is not a DaemonSet", func() {
			It("should return false and skip the event", func() {
				updateEvent := event.UpdateEvent{
					ObjectOld: &appsv1.DaemonSet{
						ObjectMeta: metav1.ObjectMeta{
							ResourceVersion: "123",
						},
					},
					ObjectNew: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-pod",
							Namespace: "test-namespace",
						},
					},
				}

				result := updateFunc(updateEvent)
				Expect(result).To(BeFalse(), "Expected the event to be skipped")
			})
		})

		Context("when the new object is not managed by the controller", func() {
			It("should return false and skip the event", func() {
				updateEvent := event.UpdateEvent{
					ObjectOld: &appsv1.DaemonSet{
						ObjectMeta: metav1.ObjectMeta{
							ResourceVersion: "123",
						},
					},
					ObjectNew: &appsv1.DaemonSet{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "test-daemonset",
							Namespace:       "test-namespace",
							ResourceVersion: "456",
							// No owner references - not managed
						},
					},
				}

				result := updateFunc(updateEvent)
				Expect(result).To(BeFalse(), "Expected the event to be skipped when not managed")
			})
		})

		Context("when the old object is not a DaemonSet", func() {
			It("should return false and skip the event", func() {
				updateEvent := event.UpdateEvent{
					ObjectOld: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-pod",
							Namespace: "test-namespace",
						},
					},
					ObjectNew: &appsv1.DaemonSet{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-daemonset",
							Namespace: "test-namespace",
							OwnerReferences: []metav1.OwnerReference{
								{
									Kind:       datav1alpha1.JindoRuntimeKind,
									APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
									Controller: ptr.To(true),
								},
							},
							ResourceVersion: "456",
						},
					},
				}

				result := updateFunc(updateEvent)
				Expect(result).To(BeFalse(), "Expected the event to be skipped")
			})
		})

		Context("when resource versions are the same", func() {
			It("should return false and skip the event", func() {
				updateEvent := event.UpdateEvent{
					ObjectOld: &appsv1.DaemonSet{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-daemonset",
							Namespace: "test-namespace",
							OwnerReferences: []metav1.OwnerReference{
								{
									Kind:       datav1alpha1.JindoRuntimeKind,
									APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
									Controller: ptr.To(true),
								},
							},
							ResourceVersion: "456",
						},
					},
					ObjectNew: &appsv1.DaemonSet{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-daemonset",
							Namespace: "test-namespace",
							OwnerReferences: []metav1.OwnerReference{
								{
									Kind:       datav1alpha1.JindoRuntimeKind,
									APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
									Controller: ptr.To(true),
								},
							},
							ResourceVersion: "456",
						},
					},
				}

				result := updateFunc(updateEvent)
				Expect(result).To(BeFalse(), "Expected the event to be skipped when resource versions match")
			})
		})
	})

	Describe("onDeleteFunc", func() {
		var deleteFunc func(e event.DeleteEvent) bool

		BeforeEach(func() {
			deleteFunc = handler.onDeleteFunc(fakeRuntimeReconciler)
		})

		Context("when the object is a DaemonSet with RuntimeInterface owner", func() {
			It("should return true and reconcile the event", func() {
				deleteEvent := event.DeleteEvent{
					Object: &appsv1.DaemonSet{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-daemonset",
							Namespace: "test-namespace",
							OwnerReferences: []metav1.OwnerReference{
								{
									Kind:       datav1alpha1.JindoRuntimeKind,
									APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
									Controller: ptr.To(true),
								},
							},
						},
					},
				}

				result := deleteFunc(deleteEvent)
				Expect(result).To(BeTrue(), "Expected the event to be reconciled")
			})
		})

		Context("when the object is not a DaemonSet", func() {
			It("should return false and skip the event", func() {
				deleteEvent := event.DeleteEvent{
					Object: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-pod",
							Namespace: "test-namespace",
						},
					},
				}

				result := deleteFunc(deleteEvent)
				Expect(result).To(BeFalse(), "Expected the event to be skipped")
			})
		})

		Context("when the object is not managed by the controller", func() {
			It("should return false and skip the event", func() {
				deleteEvent := event.DeleteEvent{
					Object: &appsv1.DaemonSet{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-daemonset",
							Namespace: "test-namespace",
							// No owner references - not managed
						},
					},
				}

				result := deleteFunc(deleteEvent)
				Expect(result).To(BeFalse(), "Expected the event to be skipped when not managed")
			})
		})
	})
})
