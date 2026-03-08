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

package fluidapp

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func newFluidAppScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = corev1.AddToScheme(s)
	return s
}

var _ = Describe("FluidAppReconciler", func() {
	const (
		defaultNamespace = "default"
		testPodName      = "test-pod"
	)

	Describe("ControllerName", func() {
		It("should return FluidAppController", func() {
			reconciler := &FluidAppReconciler{}
			Expect(reconciler.ControllerName()).To(Equal("FluidAppController"))
		})
	})

	Describe("ManagedResource", func() {
		It("should return a Pod object", func() {
			reconciler := &FluidAppReconciler{}
			obj := reconciler.ManagedResource()
			Expect(obj).NotTo(BeNil())
			_, ok := obj.(*corev1.Pod)
			Expect(ok).To(BeTrue())
		})
	})

	Describe("NewFluidAppReconciler", func() {
		It("should create a new reconciler", func() {
			s := newFluidAppScheme()
			c := fake.NewFakeClientWithScheme(s)
			reconciler := NewFluidAppReconciler(c, fake.NullLogger(), record.NewFakeRecorder(100))
			Expect(reconciler).NotTo(BeNil())
			Expect(reconciler.Client).NotTo(BeNil())
			Expect(reconciler.FluidAppReconcilerImplement).NotTo(BeNil())
		})
	})

	Describe("Reconcile", func() {
		Context("when pod does not exist", func() {
			It("should return no error and not requeue", func() {
				s := newFluidAppScheme()
				c := fake.NewFakeClientWithScheme(s)
				reconciler := NewFluidAppReconciler(c, fake.NullLogger(), record.NewFakeRecorder(100))

				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "non-existent-pod",
						Namespace: defaultNamespace,
					},
				}

				result, err := reconciler.Reconcile(context.Background(), req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())
			})
		})

		Context("when pod should not be in queue", func() {
			It("should return no error and not requeue", func() {
				s := newFluidAppScheme()
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testPodName,
						Namespace: defaultNamespace,
					},
				}
				c := fake.NewFakeClientWithScheme(s, pod)
				reconciler := NewFluidAppReconciler(c, fake.NullLogger(), record.NewFakeRecorder(100))

				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      testPodName,
						Namespace: defaultNamespace,
					},
				}

				result, err := reconciler.Reconcile(context.Background(), req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())
			})
		})
	})

	Describe("internalReconcile", func() {
		Context("when pod has fuse containers", func() {
			It("should handle pod reconciliation with fuse sidecars", func() {
				s := newFluidAppScheme()
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testPodName,
						Namespace: defaultNamespace,
						Labels: map[string]string{
							common.LabelAnnotationManagedBy: common.Fluid,
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "test"},
							{
								Name: common.FuseContainerName + "-0",
								VolumeMounts: []corev1.VolumeMount{{
									Name:      "fuse-mount",
									MountPath: "/mnt/fuse",
								}},
							},
						},
					},
				}
				c := fake.NewFakeClientWithScheme(s, pod)
				reconciler := NewFluidAppReconciler(c, fake.NullLogger(), record.NewFakeRecorder(100))

				ctx := reconcileRequestContext{
					Context: context.Background(),
					Log:     fake.NullLogger(),
					pod:     pod,
					NamespacedName: types.NamespacedName{
						Name:      testPodName,
						Namespace: defaultNamespace,
					},
				}

				result, err := reconciler.internalReconcile(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))
			})
		})

		It("should handle pod reconciliation", func() {
			s := newFluidAppScheme()
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testPodName,
					Namespace: defaultNamespace,
					Labels: map[string]string{
						common.LabelAnnotationManagedBy: common.Fluid,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "test"}},
				},
			}
			c := fake.NewFakeClientWithScheme(s, pod)
			reconciler := NewFluidAppReconciler(c, fake.NullLogger(), record.NewFakeRecorder(100))

			ctx := reconcileRequestContext{
				Context: context.Background(),
				Log:     fake.NullLogger(),
				pod:     pod,
				NamespacedName: types.NamespacedName{
					Name:      testPodName,
					Namespace: defaultNamespace,
				},
			}

			result, err := reconciler.internalReconcile(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})
	})
})
