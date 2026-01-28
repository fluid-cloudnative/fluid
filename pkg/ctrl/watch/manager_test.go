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
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

var _ = Describe("IsObjectInManaged", func() {
	var reconciler *FakeRuntimeReconciler

	BeforeEach(func() {
		reconciler = &FakeRuntimeReconciler{}
		// Test methods are called
		_ = reconciler.ControllerName()
		_ = reconciler.ManagedResource()
	})

	Context("when object has controller reference", func() {
		It("should be managed", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "managed-pod",
					Namespace: "default",
					OwnerReferences: []metav1.OwnerReference{
						{
							Kind:       datav1alpha1.JindoRuntimeKind,
							APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
							Controller: ptr.To(true),
						},
					},
				},
			}
			isManaged := isObjectInManaged(pod, reconciler)
			Expect(isManaged).To(BeTrue())
		})
	})

	Context("when object has no controller flag", func() {
		It("should not be managed", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "not-managed-pod",
					Namespace: "default",
					OwnerReferences: []metav1.OwnerReference{
						{
							Kind:       datav1alpha1.JindoRuntimeKind,
							APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
						},
					},
				},
			}
			isManaged := isObjectInManaged(pod, reconciler)
			Expect(isManaged).To(BeFalse())
		})
	})

	Context("when object has wrong kind", func() {
		It("should not be managed", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wrong-kind-pod",
					Namespace: "default",
					OwnerReferences: []metav1.OwnerReference{
						{
							Kind:       "Test",
							APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
							Controller: ptr.To(true),
						},
					},
				},
			}
			isManaged := isObjectInManaged(pod, reconciler)
			Expect(isManaged).To(BeFalse())
		})
	})

	Context("when object has wrong api version", func() {
		It("should not be managed", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wrong-api-pod",
					Namespace: "default",
					OwnerReferences: []metav1.OwnerReference{
						{
							Kind:       datav1alpha1.JindoRuntimeKind,
							APIVersion: "v1",
							Controller: ptr.To(true),
						},
					},
				},
			}
			isManaged := isObjectInManaged(pod, reconciler)
			Expect(isManaged).To(BeFalse())
		})
	})

	Context("when object has no owner references", func() {
		It("should not be managed", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-owner-pod",
					Namespace: "default",
				},
			}
			isManaged := isObjectInManaged(pod, reconciler)
			Expect(isManaged).To(BeFalse())
		})
	})
})

var _ = Describe("IsOwnerMatched", func() {
	var reconciler *FakeRuntimeReconciler

	BeforeEach(func() {
		reconciler = &FakeRuntimeReconciler{}
	})

	Context("when owner reference is matched", func() {
		It("should return true", func() {
			controllerRef := &metav1.OwnerReference{
				Kind:       datav1alpha1.JindoRuntimeKind,
				APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
				Controller: ptr.To(true),
			}
			matched := isOwnerMatched(controllerRef, reconciler)
			Expect(matched).To(BeTrue())
		})
	})

	Context("when kind is unmatched", func() {
		It("should return false", func() {
			controllerRef := &metav1.OwnerReference{
				Kind:       "WrongKind",
				APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
				Controller: ptr.To(true),
			}
			matched := isOwnerMatched(controllerRef, reconciler)
			Expect(matched).To(BeFalse())
		})
	})

	Context("when api version is unmatched", func() {
		It("should return false", func() {
			controllerRef := &metav1.OwnerReference{
				Kind:       datav1alpha1.JindoRuntimeKind,
				APIVersion: "apps/v1",
				Controller: ptr.To(true),
			}
			matched := isOwnerMatched(controllerRef, reconciler)
			Expect(matched).To(BeFalse())
		})
	})
})

var _ = Describe("ControllerInterface", func() {
	var reconciler *FakeRuntimeReconciler

	BeforeEach(func() {
		reconciler = &FakeRuntimeReconciler{}
	})

	It("should call controller name method", func() {
		name := reconciler.ControllerName()
		Expect(name).To(Equal(""))
	})

	It("should return non-nil managed resource", func() {
		resource := reconciler.ManagedResource()
		Expect(resource).NotTo(BeNil())
	})

	It("should execute reconcile without error", func() {
		ctx := context.Background()
		req := ctrl.Request{}
		result, err := reconciler.Reconcile(ctx, req)
		Expect(err).To(BeNil())
		Expect(result).To(Equal(ctrl.Result{}))
	})
})

var _ = Describe("IsObjectInManaged Edge Cases", func() {
	var reconciler *FakeRuntimeReconciler

	BeforeEach(func() {
		reconciler = &FakeRuntimeReconciler{}
	})

	Context("when object has multiple owner references", func() {
		It("should not be managed if matching reference doesn't have controller=true", func() {
			podWithMultipleOwners := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "multi-owner-pod",
					Namespace: "default",
					OwnerReferences: []metav1.OwnerReference{
						{
							Kind:       "ReplicaSet",
							APIVersion: "apps/v1",
							Controller: ptr.To(true),
						},
						{
							Kind:       datav1alpha1.JindoRuntimeKind,
							APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
						},
					},
				},
			}
			isManaged := isObjectInManaged(podWithMultipleOwners, reconciler)
			Expect(isManaged).To(BeFalse())
		})

		It("should be managed if matching controller reference is second item", func() {
			podWithMatchingSecond := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "matching-second-pod",
					Namespace: "default",
					OwnerReferences: []metav1.OwnerReference{
						{
							Kind:       "ReplicaSet",
							APIVersion: "apps/v1",
						},
						{
							Kind:       datav1alpha1.JindoRuntimeKind,
							APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
							Controller: ptr.To(true),
						},
					},
				},
			}
			isManaged := isObjectInManaged(podWithMatchingSecond, reconciler)
			Expect(isManaged).To(BeTrue())
		})
	})
})
