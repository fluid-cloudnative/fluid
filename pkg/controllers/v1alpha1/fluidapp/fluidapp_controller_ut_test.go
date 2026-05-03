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
	"errors"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

var _ = Describe("FluidAppReconciler", func() {
	const (
		testNamespace = "default"
		testPodName   = "fluidapp-pod"
	)

	var scheme *runtime.Scheme

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(datav1alpha1.AddToScheme(scheme)).To(Succeed())
	})

	Describe("ControllerName", func() {
		It("returns the controller name constant", func() {
			r := &FluidAppReconciler{}
			Expect(r.ControllerName()).To(Equal(controllerName))
		})
	})

	Describe("ManagedResource", func() {
		It("returns a pod object", func() {
			r := &FluidAppReconciler{}
			Expect(r.ManagedResource()).To(BeAssignableToTypeOf(&corev1.Pod{}))
		})
	})

	Describe("NewFluidAppReconciler", func() {
		It("constructs a reconciler with the provided dependencies", func() {
			fakeClient := fake.NewFakeClientWithScheme(scheme)
			recorder := record.NewFakeRecorder(10)

			r := NewFluidAppReconciler(fakeClient, fake.NullLogger(), recorder)

			Expect(r).NotTo(BeNil())
			Expect(r.Client).To(Equal(fakeClient))
			Expect(r.Recorder).To(Equal(recorder))
			Expect(r.FluidAppReconcilerImplement).NotTo(BeNil())
			Expect(r.FluidAppReconcilerImplement.Client).To(Equal(fakeClient))
		})
	})

	Describe("Reconcile", func() {
		It("returns no requeue when the pod does not exist", func() {
			fakeClient := fake.NewFakeClientWithScheme(scheme)
			r := NewFluidAppReconciler(fakeClient, fake.NullLogger(), record.NewFakeRecorder(10))

			result, err := r.Reconcile(context.Background(), reconcile.Request{
				NamespacedName: types.NamespacedName{Name: testPodName, Namespace: testNamespace},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})

		It("returns no requeue when the pod should not enter the queue", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testPodName,
					Namespace: testNamespace,
				},
			}
			fakeClient := fake.NewFakeClientWithScheme(scheme, pod)
			r := NewFluidAppReconciler(fakeClient, fake.NullLogger(), record.NewFakeRecorder(10))

			result, err := r.Reconcile(context.Background(), reconcile.Request{
				NamespacedName: types.NamespacedName{Name: testPodName, Namespace: testNamespace},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})

		It("reconciles queueable pods and returns no requeue on successful fuse unmount", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testPodName,
					Namespace: testNamespace,
					Labels: map[string]string{
						common.InjectServerless:  common.True,
						common.InjectSidecarDone: common.True,
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{Name: "app"},
						{
							Name: common.FuseContainerName + "-0",
							Lifecycle: &corev1.Lifecycle{
								PreStop: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{Command: []string{"umount", "/mnt/fuse"}},
								},
							},
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:  "app",
							State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 0}},
						},
						{
							Name:  common.FuseContainerName + "-0",
							State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
						},
					},
				},
			}

			fakeClient := fake.NewFakeClientWithScheme(scheme, pod)
			r := NewFluidAppReconciler(fakeClient, fake.NullLogger(), record.NewFakeRecorder(10))

			patches := gomonkey.ApplyFunc((*FluidAppReconcilerImplement).umountFuseSidecars, func(_ *FluidAppReconcilerImplement, gotPod *corev1.Pod) error {
				Expect(gotPod.Name).To(Equal(testPodName))
				return nil
			})
			defer patches.Reset()

			result, err := r.Reconcile(context.Background(), reconcile.Request{
				NamespacedName: types.NamespacedName{Name: testPodName, Namespace: testNamespace},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})
	})

	Describe("internalReconcile", func() {
		It("requeues when unmounting fuse sidecars returns an error", func() {
			r := NewFluidAppReconciler(fake.NewFakeClientWithScheme(scheme), fake.NullLogger(), record.NewFakeRecorder(10))
			pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: testPodName, Namespace: testNamespace}}
			expectedErr := errors.New("umount failed")

			patches := gomonkey.ApplyFunc((*FluidAppReconcilerImplement).umountFuseSidecars, func(_ *FluidAppReconcilerImplement, gotPod *corev1.Pod) error {
				Expect(gotPod).To(Equal(pod))
				return expectedErr
			})
			defer patches.Reset()

			result, err := r.internalReconcile(reconcileRequestContext{
				Context: context.Background(),
				Log:     fake.NullLogger(),
				pod:     pod,
			})

			Expect(err).To(MatchError(expectedErr))
			Expect(result).To(Equal(ctrl.Result{}))
		})

		It("returns no requeue when unmounting fuse sidecars succeeds", func() {
			r := NewFluidAppReconciler(fake.NewFakeClientWithScheme(scheme), fake.NullLogger(), record.NewFakeRecorder(10))
			pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: testPodName, Namespace: testNamespace}}

			patches := gomonkey.ApplyFunc((*FluidAppReconcilerImplement).umountFuseSidecars, func(_ *FluidAppReconcilerImplement, gotPod *corev1.Pod) error {
				Expect(gotPod).To(Equal(pod))
				return nil
			})
			defer patches.Reset()

			result, err := r.internalReconcile(reconcileRequestContext{
				Context: context.Background(),
				Log:     fake.NullLogger(),
				pod:     pod,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})
	})
})
