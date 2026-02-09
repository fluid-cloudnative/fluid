/*
  Copyright 2022 The Fluid Authors.

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

package thin

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	ctrlhelper "github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

const (
	workerTestNamespace      = "thin"
	workerTestFluidNs        = "fluid"
	workerTestName0          = "test0"
	workerTestName1          = "test1"
	workerTestName2          = "test2"
	workerTestName3          = "test3"
	workerTestName           = "test"
	workerTestNodeName       = "test-node"
	workerTestNodeSelect     = "node-select"
	workerTestWorkerSuffix   = "-worker"
	workerTestFuseSuffix     = "-fuse"
	workerTestSparkName      = "spark"
	workerTestWorkerSelector = "app=thin,release=spark,role=thin-worker"
)

var _ = Describe("ShouldSetupWorkers", Label("pkg.ddc.thin.worker_test.go"), func() {
	DescribeTable("should return correct result based on worker phase and enablement",
		func(name string, workerEnabled bool, workerPhase datav1alpha1.RuntimePhase, expected bool) {
			thinRuntime := &datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: workerTestNamespace,
				},
				Spec: datav1alpha1.ThinRuntimeSpec{
					Worker: datav1alpha1.ThinCompTemplateSpec{
						Enabled: workerEnabled,
					},
				},
				Status: datav1alpha1.RuntimeStatus{
					WorkerPhase: workerPhase,
				},
			}

			data := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: workerTestNamespace,
				},
			}

			s := runtime.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, thinRuntime)
			s.AddKnownTypes(datav1alpha1.GroupVersion, data)
			err := v1.AddToScheme(s)
			Expect(err).NotTo(HaveOccurred())

			mockClient := fake.NewFakeClientWithScheme(s, thinRuntime, data)
			e := &ThinEngine{
				name:      name,
				namespace: workerTestNamespace,
				runtime:   thinRuntime,
				Client:    mockClient,
			}

			gotShould, err := e.ShouldSetupWorkers()
			Expect(err).NotTo(HaveOccurred())
			Expect(gotShould).To(Equal(expected))
		},
		Entry("worker phase None and enabled returns true",
			workerTestName0, true, datav1alpha1.RuntimePhaseNone, true,
		),
		Entry("worker phase NotReady returns false",
			workerTestName1, true, datav1alpha1.RuntimePhaseNotReady, false,
		),
		Entry("worker phase PartialReady returns false",
			workerTestName2, true, datav1alpha1.RuntimePhasePartialReady, false,
		),
		Entry("worker phase Ready returns false",
			workerTestName3, true, datav1alpha1.RuntimePhaseReady, false,
		),
		Entry("worker not enabled returns false",
			workerTestName3, false, datav1alpha1.RuntimePhaseNone, false,
		),
	)
})

var _ = Describe("SetupWorkers", Label("pkg.ddc.thin.worker_test.go"), func() {
	Context("when setting up workers with runtimeInfo", func() {
		It("should scale workers to desired replicas", func() {
			runtimeInfo, err := base.BuildRuntimeInfo(workerTestName, workerTestFluidNs, common.ThinRuntime)
			Expect(err).NotTo(HaveOccurred())

			runtimeInfo.SetupWithDataset(&datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
			})

			nodeSelector := map[string]string{
				workerTestNodeSelect: "true",
			}
			runtimeInfo.SetFuseNodeSelector(nodeSelector)

			nodeInputs := []*v1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: workerTestNodeName,
					},
				},
			}

			worker := appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      workerTestName + workerTestWorkerSuffix,
					Namespace: workerTestFluidNs,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](0),
				},
			}

			thinRuntime := &datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      workerTestName,
					Namespace: workerTestFluidNs,
				},
				Spec: datav1alpha1.ThinRuntimeSpec{
					Replicas: 1,
					Worker: datav1alpha1.ThinCompTemplateSpec{
						Enabled: true,
					},
				},
			}

			data := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      workerTestName,
					Namespace: workerTestFluidNs,
				},
			}

			runtimeObjs := []runtime.Object{}
			for _, nodeInput := range nodeInputs {
				runtimeObjs = append(runtimeObjs, nodeInput.DeepCopy())
			}
			runtimeObjs = append(runtimeObjs, worker.DeepCopy())

			s := runtime.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, thinRuntime)
			s.AddKnownTypes(datav1alpha1.GroupVersion, data)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, &worker)
			err = v1.AddToScheme(s)
			Expect(err).NotTo(HaveOccurred())

			runtimeObjs = append(runtimeObjs, thinRuntime, data)
			mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)

			e := &ThinEngine{
				runtime:     thinRuntime,
				runtimeInfo: runtimeInfo,
				Client:      mockClient,
				name:        workerTestName,
				namespace:   workerTestFluidNs,
				Log:         ctrl.Log.WithName(workerTestName),
			}

			e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)
			err = e.SetupWorkers()
			Expect(err).NotTo(HaveOccurred())

			// Verify that the StatefulSet was scaled to the desired replicas
			var updatedWorker appsv1.StatefulSet
			err = mockClient.Get(context.TODO(), types.NamespacedName{
				Name:      workerTestName + workerTestWorkerSuffix,
				Namespace: workerTestFluidNs,
			}, &updatedWorker)
			Expect(err).NotTo(HaveOccurred())
			Expect(*updatedWorker.Spec.Replicas).To(Equal(int32(1)))
		})
	})
})

var _ = Describe("CheckWorkersReady", Label("pkg.ddc.thin.worker_test.go"), func() {
	Context("when workers and fuse are ready", func() {
		It("should return true", func() {
			thinRuntime := &datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      workerTestName0,
					Namespace: workerTestNamespace,
				},
				Spec: datav1alpha1.ThinRuntimeSpec{
					Replicas: 1,
					Worker: datav1alpha1.ThinCompTemplateSpec{
						Enabled: true,
					},
					Fuse: datav1alpha1.ThinFuseSpec{},
				},
			}

			worker := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      workerTestName0 + workerTestWorkerSuffix,
					Namespace: workerTestNamespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](1),
				},
				Status: appsv1.StatefulSetStatus{
					Replicas:      1,
					ReadyReplicas: 1,
				},
			}

			fuse := &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      workerTestName0 + workerTestFuseSuffix,
					Namespace: workerTestNamespace,
				},
				Status: appsv1.DaemonSetStatus{
					NumberAvailable:        1,
					DesiredNumberScheduled: 1,
					CurrentNumberScheduled: 1,
				},
			}

			data := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      workerTestName0,
					Namespace: workerTestNamespace,
				},
			}

			s := runtime.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, thinRuntime)
			s.AddKnownTypes(datav1alpha1.GroupVersion, data)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, fuse)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
			err := v1.AddToScheme(s)
			Expect(err).NotTo(HaveOccurred())

			mockClient := fake.NewFakeClientWithScheme(s, thinRuntime, data, fuse, worker)
			e := &ThinEngine{
				runtime:   thinRuntime,
				name:      workerTestName0,
				namespace: workerTestNamespace,
				Client:    mockClient,
				Log:       ctrl.Log.WithName(workerTestName0),
			}

			testRuntimeInfo, err := base.BuildRuntimeInfo(workerTestName0, workerTestNamespace, common.ThinRuntime)
			Expect(err).NotTo(HaveOccurred())

			e.Helper = ctrlhelper.BuildHelper(testRuntimeInfo, mockClient, e.Log)

			gotReady, err := e.CheckWorkersReady()
			Expect(err).NotTo(HaveOccurred())
			Expect(gotReady).To(BeTrue())
		})
	})

	Context("when workers are not ready", func() {
		It("should return false", func() {
			thinRuntime := &datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      workerTestName1,
					Namespace: workerTestNamespace,
				},
				Spec: datav1alpha1.ThinRuntimeSpec{
					Replicas: 1,
					Worker: datav1alpha1.ThinCompTemplateSpec{
						Enabled: true,
					},
					Fuse: datav1alpha1.ThinFuseSpec{},
				},
			}

			worker := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      workerTestName1 + workerTestWorkerSuffix,
					Namespace: workerTestNamespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](1),
				},
				Status: appsv1.StatefulSetStatus{
					Replicas:      1,
					ReadyReplicas: 0,
				},
			}

			fuse := &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      workerTestName1 + workerTestFuseSuffix,
					Namespace: workerTestNamespace,
				},
				Status: appsv1.DaemonSetStatus{
					NumberAvailable:        0,
					DesiredNumberScheduled: 1,
					CurrentNumberScheduled: 0,
				},
			}

			data := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      workerTestName1,
					Namespace: workerTestNamespace,
				},
			}

			s := runtime.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, thinRuntime)
			s.AddKnownTypes(datav1alpha1.GroupVersion, data)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, fuse)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
			err := v1.AddToScheme(s)
			Expect(err).NotTo(HaveOccurred())

			mockClient := fake.NewFakeClientWithScheme(s, thinRuntime, data, fuse, worker)
			e := &ThinEngine{
				runtime:   thinRuntime,
				name:      workerTestName1,
				namespace: workerTestNamespace,
				Client:    mockClient,
				Log:       ctrl.Log.WithName(workerTestName1),
			}

			testRuntimeInfo, err := base.BuildRuntimeInfo(workerTestName1, workerTestNamespace, common.ThinRuntime)
			Expect(err).NotTo(HaveOccurred())

			e.Helper = ctrlhelper.BuildHelper(testRuntimeInfo, mockClient, e.Log)

			gotReady, err := e.CheckWorkersReady()
			Expect(err).NotTo(HaveOccurred())
			Expect(gotReady).To(BeFalse())
		})
	})
})

var _ = Describe("GetWorkerSelectors", Label("pkg.ddc.thin.worker_test.go"), func() {
	Context("when getting worker selectors", func() {
		It("should return the correct selector string", func() {
			e := &ThinEngine{
				name: workerTestSparkName,
			}

			got := e.getWorkerSelectors()
			Expect(got).To(Equal(workerTestWorkerSelector))
		})
	})
})
