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

package engine

import (
	"context"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("CacheEngine Worker Component Tests", Label("pkg.ddc.cache.engine.worker_test.go"), func() {
	var (
		engine     *CacheEngine
		runtimeObj *datav1alpha1.CacheRuntime
		fakeClient client.Client
	)

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(scheme)
		_ = appsv1.AddToScheme(scheme)
		_ = corev1.AddToScheme(scheme)

		// Create runtime with None phase (needs setup)
		runtimeObj = &datav1alpha1.CacheRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-runtime",
				Namespace: "default",
			},
			Spec: datav1alpha1.CacheRuntimeSpec{
				RuntimeClassName: "test-class",
			},
		}
		// Initialize status with None phase
		runtimeObj.Status.Master.Phase = datav1alpha1.RuntimePhaseNone
		runtimeObj.Status.Worker.Phase = datav1alpha1.RuntimePhaseNone
		runtimeObj.Status.Client.Phase = datav1alpha1.RuntimePhaseNone

		fakeClient = fake.NewFakeClientWithScheme(scheme, runtimeObj)

		engine = &CacheEngine{
			name:      "test-runtime",
			namespace: "default",
			Client:    fakeClient,
			Log:       ctrl.Log.WithName("test"),
		}
	})

	Describe("Worker Component Setup", func() {
		Context("ShouldSetupWorker", func() {
			It("should return true when worker phase is None", func() {
				shouldSetup, err := engine.ShouldSetupWorker()
				Expect(err).NotTo(HaveOccurred())
				Expect(shouldSetup).To(BeTrue())
			})

			It("should return false when worker phase is NotReady", func() {
				Skip("Requires real Kubernetes API for status update")
			})

			It("should return false when worker phase is Ready", func() {
				Skip("Requires real Kubernetes API for status update")
			})

			It("should return error when runtime not found", func() {
				engine.name = "non-existent-runtime"
				shouldSetup, err := engine.ShouldSetupWorker()
				Expect(err).To(HaveOccurred())
				Expect(shouldSetup).To(BeFalse())
			})
		})

		Context("SetupWorkerComponent", func() {
			var workerValue *common.CacheRuntimeComponentValue

			BeforeEach(func() {
				workerValue = &common.CacheRuntimeComponentValue{
					Name:      "test-runtime-worker",
					Namespace: "default",
					Enabled:   true,
					Replicas:  2,
					Owner: &common.OwnerReference{
						APIVersion: "data.fluid.io/v1alpha1",
						Kind:       "CacheRuntime",
						Name:       "test-runtime",
						UID:        "test-uid",
					},
					WorkloadType: metav1.TypeMeta{
						Kind:       "StatefulSet",
						APIVersion: "apps/v1",
					},
					PodTemplateSpec: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "worker",
									Image: "test-worker:latest",
								},
							},
						},
					},
				}
			})

			It("should setup worker when phase is None", func() {
				ready, err := engine.SetupWorkerComponent(workerValue)
				Expect(err).NotTo(HaveOccurred())
				Expect(ready).To(BeTrue())

				// Verify StatefulSet was created
				sts := &appsv1.StatefulSet{}
				err = engine.Client.Get(context.TODO(),
					client.ObjectKey{Name: "test-runtime-worker", Namespace: "default"}, sts)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should skip setup when worker already initialized", func() {
				Skip("Requires real Kubernetes API for status update")
			})

			It("should update runtime status after setup", func() {
				_, err := engine.SetupWorkerComponent(workerValue)
				Expect(err).NotTo(HaveOccurred())

				// Get updated runtime
				updatedRuntime := &datav1alpha1.CacheRuntime{}
				err = engine.Client.Get(context.TODO(),
					client.ObjectKey{Name: "test-runtime", Namespace: "default"}, updatedRuntime)
				Expect(err).NotTo(HaveOccurred())

				// Verify status was updated
				Expect(updatedRuntime.Status.Worker.Phase).To(Equal(datav1alpha1.RuntimePhaseNotReady))
				Expect(updatedRuntime.Status.Conditions).NotTo(BeEmpty())
				Expect(updatedRuntime.Status.Conditions[0].Type).To(Equal(datav1alpha1.RuntimeWorkersReady))
				Expect(updatedRuntime.Status.Conditions[0].Status).To(Equal(corev1.ConditionTrue))
			})
		})
	})
})
