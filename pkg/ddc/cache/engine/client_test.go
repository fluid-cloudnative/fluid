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

var _ = Describe("CacheEngine Client Component Tests", Label("pkg.ddc.cache.engine.client_test.go"), func() {
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

	Describe("Client Component Setup", func() {
		Context("ShouldSetupClient", func() {
			It("should return true when client phase is None", func() {
				shouldSetup, err := engine.ShouldSetupClient()
				Expect(err).NotTo(HaveOccurred())
				Expect(shouldSetup).To(BeTrue())
			})

			It("should return false when client phase is NotReady", func() {
				Skip("Requires real Kubernetes API for status update")
			})

			It("should return false when client phase is Ready", func() {
				Skip("Requires real Kubernetes API for status update")
			})

			It("should return error when runtime not found", func() {
				engine.name = "non-existent-runtime"
				shouldSetup, err := engine.ShouldSetupClient()
				Expect(err).To(HaveOccurred())
				Expect(shouldSetup).To(BeFalse())
			})
		})

		Context("SetupClientComponent", func() {
			var clientValue *common.CacheRuntimeComponentValue

			BeforeEach(func() {
				clientValue = &common.CacheRuntimeComponentValue{
					Name:      "test-runtime-client",
					Namespace: "default",
					Enabled:   true,
					Replicas:  1,
					Owner: &common.OwnerReference{
						APIVersion: "data.fluid.io/v1alpha1",
						Kind:       "CacheRuntime",
						Name:       "test-runtime",
						UID:        "test-uid",
					},
					WorkloadType: metav1.TypeMeta{
						Kind:       "DaemonSet",
						APIVersion: "apps/v1",
					},
					PodTemplateSpec: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "client",
									Image: "test-client:latest",
								},
							},
						},
					},
				}
			})

			It("should setup client when phase is None", func() {
				ready, err := engine.SetupClientComponent(clientValue)
				Expect(err).NotTo(HaveOccurred())
				Expect(ready).To(BeTrue())

				// Verify DaemonSet was created
				ds := &appsv1.DaemonSet{}
				err = engine.Client.Get(context.TODO(),
					client.ObjectKey{Name: "test-runtime-client", Namespace: "default"}, ds)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should skip setup when client already initialized", func() {
				Skip("Requires real Kubernetes API for status update")
			})

			It("should update runtime status after setup", func() {
				_, err := engine.SetupClientComponent(clientValue)
				Expect(err).NotTo(HaveOccurred())

				// Get updated runtime
				updatedRuntime := &datav1alpha1.CacheRuntime{}
				err = engine.Client.Get(context.TODO(),
					client.ObjectKey{Name: "test-runtime", Namespace: "default"}, updatedRuntime)
				Expect(err).NotTo(HaveOccurred())

				// Verify status was updated
				Expect(updatedRuntime.Status.Client.Phase).To(Equal(datav1alpha1.RuntimePhaseNotReady))
				Expect(updatedRuntime.Status.Conditions).NotTo(BeEmpty())
				Expect(updatedRuntime.Status.Conditions[0].Type).To(Equal(datav1alpha1.RuntimeFusesInitialized))
				Expect(updatedRuntime.Status.Conditions[0].Status).To(Equal(corev1.ConditionTrue))
			})
		})
	})
})
