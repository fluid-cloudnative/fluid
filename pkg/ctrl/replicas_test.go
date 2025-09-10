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

package ctrl

import (
	"context"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ctrl Replicas Tests", func() {
	// Shared variables for all tests
	var (
		helper       *Helper
		resources    []runtime.Object
		k8sClient    client.Client
		runtimeInfo  base.RuntimeInfoInterface
		fluidRuntime *datav1alpha1.JuiceFSRuntime
		dataset      *datav1alpha1.Dataset
		workerSts    *appsv1.StatefulSet
	)

	BeforeEach(func() {
		// Initialize shared resources
		fluidRuntime = &datav1alpha1.JuiceFSRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.JuiceFSRuntimeSpec{
				Replicas: 3,
			},
			Status: datav1alpha1.RuntimeStatus{},
		}

		dataset = &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "fluid",
			},
		}

		workerSts = &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-worker",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: nil, // should be set in each test case
			},
		}
	})

	JustBeforeEach(func() {
		k8sClient = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
		runtimeInfo, _ = base.BuildRuntimeInfo(fluidRuntime.Name, fluidRuntime.Namespace, common.JuiceFSRuntime)
		helper = BuildHelper(runtimeInfo, k8sClient, fake.NullLogger())
	})

	Describe("Test Helper.SyncReplicas()", func() {
		When("Runtime has a spec.replicas larger than worker sts's spec.replicas", func() {
			BeforeEach(func() {
				// Override specific settings for scale out test
				fluidRuntime.Spec.Replicas = 3
				workerSts.Spec.Replicas = ptr.To[int32](2)
				fluidRuntime.Status.DesiredWorkerNumberScheduled = 2

				resources = []runtime.Object{
					fluidRuntime,
					dataset,
					workerSts,
				}
			})

			It("should scale out worker statefulset and update runtime conditions", func() {
				ctx := cruntime.ReconcileRequestContext{
					Log:      fake.NullLogger(),
					Recorder: record.NewFakeRecorder(300),
				}

				// Perform sync replicas operation
				err := helper.SyncReplicas(ctx, fluidRuntime, fluidRuntime.Status, workerSts)
				Expect(err).NotTo(HaveOccurred())

				// Verify the updated runtime
				updatedRuntime := &datav1alpha1.JuiceFSRuntime{}
				err = k8sClient.Get(context.TODO(), types.NamespacedName{
					Namespace: fluidRuntime.Namespace,
					Name:      fluidRuntime.Name,
				}, updatedRuntime)
				Expect(err).NotTo(HaveOccurred())

				// Check that conditions are updated
				Expect(updatedRuntime.Status.Conditions).To(HaveLen(2))
				Expect(updatedRuntime.Status.Conditions[0].Type).To(Equal(datav1alpha1.RuntimeWorkerScaledOut))
				Expect(updatedRuntime.Status.Conditions[0].Status).To(Equal(corev1.ConditionTrue))
				Expect(updatedRuntime.Status.Conditions[1].Type).To(Equal(datav1alpha1.RuntimeWorkersInitialized))
				Expect(updatedRuntime.Status.Conditions[1].Status).To(Equal(corev1.ConditionTrue))

				updatedSts := &appsv1.StatefulSet{}
				err = k8sClient.Get(ctx, types.NamespacedName{Name: "test-worker", Namespace: "fluid"}, updatedSts)
				Expect(err).NotTo(HaveOccurred())
				Expect(*updatedSts.Spec.Replicas).To(Equal(int32(3)))
			})
		})

		When("Runtime has a spec.replicas smaller than worker sts's spec.replicas", func() {
			BeforeEach(func() {
				// Override specific settings for scale in test
				fluidRuntime.Spec.Replicas = 2
				workerSts.Spec.Replicas = ptr.To[int32](3)
				fluidRuntime.Status.DesiredWorkerNumberScheduled = 3

				resources = []runtime.Object{
					fluidRuntime,
					dataset,
					workerSts,
				}
			})

			It("should scale in worker statefulset and update runtime conditions", func() {
				ctx := cruntime.ReconcileRequestContext{
					Log:      fake.NullLogger(),
					Recorder: record.NewFakeRecorder(300),
				}

				// Perform sync replicas operation
				err := helper.SyncReplicas(ctx, fluidRuntime, fluidRuntime.Status, workerSts)
				Expect(err).NotTo(HaveOccurred())

				// Verify the updated runtime
				updatedRuntime := &datav1alpha1.JuiceFSRuntime{}
				err = k8sClient.Get(context.TODO(), types.NamespacedName{
					Namespace: fluidRuntime.Namespace,
					Name:      fluidRuntime.Name,
				}, updatedRuntime)
				Expect(err).NotTo(HaveOccurred())

				// Check that conditions are updated
				Expect(updatedRuntime.Status.Conditions).To(HaveLen(2))
				Expect(updatedRuntime.Status.Conditions[0].Type).To(Equal(datav1alpha1.RuntimeWorkerScaledIn))
				Expect(updatedRuntime.Status.Conditions[0].Status).To(Equal(corev1.ConditionTrue))
				Expect(updatedRuntime.Status.Conditions[1].Type).To(Equal(datav1alpha1.RuntimeWorkersInitialized))
				Expect(updatedRuntime.Status.Conditions[1].Status).To(Equal(corev1.ConditionTrue))

				updatedSts := &appsv1.StatefulSet{}
				err = k8sClient.Get(ctx, types.NamespacedName{Name: "test-worker", Namespace: "fluid"}, updatedSts)
				Expect(err).NotTo(HaveOccurred())
				Expect(*updatedSts.Spec.Replicas).To(Equal(int32(2)))
			})
		})

		Context("When no runtime has a spec.replicas equal to worker sts's spec.replicas", func() {
			BeforeEach(func() {
				// Override specific settings for no action test
				fluidRuntime.Spec.Replicas = 2
				workerSts.Spec.Replicas = ptr.To[int32](2)
				fluidRuntime.Status.DesiredWorkerNumberScheduled = 2

				resources = []runtime.Object{
					fluidRuntime,
					dataset,
					workerSts,
				}
			})

			It("should not update runtime conditions when no scaling is needed", func() {
				ctx := cruntime.ReconcileRequestContext{
					Log:      fake.NullLogger(),
					Recorder: record.NewFakeRecorder(300),
				}

				// Perform sync replicas operation
				err := helper.SyncReplicas(ctx, fluidRuntime, fluidRuntime.Status, workerSts)
				Expect(err).NotTo(HaveOccurred())

				// Verify the updated runtime
				updatedRuntime := &datav1alpha1.JuiceFSRuntime{}
				err = k8sClient.Get(context.TODO(), types.NamespacedName{
					Namespace: fluidRuntime.Namespace,
					Name:      fluidRuntime.Name,
				}, updatedRuntime)
				Expect(err).NotTo(HaveOccurred())

				// Check that conditions are not updated
				Expect(updatedRuntime.Status.Conditions).To(HaveLen(0))
			})
		})
	})
})
