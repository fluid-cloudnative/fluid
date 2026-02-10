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

package efc

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/types"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	ctrlhelper "github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"k8s.io/utils/ptr"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
)

const (
	workerTestNamespaceEFC = "efc"
	workerTestNamespaceFluid = "fluid"
	workerTestNamespaceBigData = "big-data"
)

var _ = Describe("EFCEngine Worker", Label("pkg.ddc.efc.worker_test.go"), func() {
	Describe("ShouldSetupWorkers", func() {
		Context("when WorkerPhase is RuntimePhaseNone", func() {
			It("should return true", func() {
				efcRuntime := &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test0",
						Namespace: workerTestNamespaceEFC,
					},
					Status: datav1alpha1.RuntimeStatus{
						WorkerPhase: datav1alpha1.RuntimePhaseNone,
					},
				}
				data := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test0",
						Namespace: workerTestNamespaceEFC,
					},
				}

				s := runtime.NewScheme()
				s.AddKnownTypes(datav1alpha1.GroupVersion, efcRuntime)
				s.AddKnownTypes(datav1alpha1.GroupVersion, data)
				err := v1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				mockClient := fake.NewFakeClientWithScheme(s, efcRuntime, data)
				e := &EFCEngine{
					name:      "test0",
					namespace: workerTestNamespaceEFC,
					runtime:   efcRuntime,
					Client:    mockClient,
				}

				gotShould, err := e.ShouldSetupWorkers()
				Expect(err).NotTo(HaveOccurred())
				Expect(gotShould).To(BeTrue())
			})
		})

		Context("when WorkerPhase is RuntimePhaseNotReady", func() {
			It("should return false", func() {
				efcRuntime := &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: workerTestNamespaceEFC,
					},
					Status: datav1alpha1.RuntimeStatus{
						WorkerPhase: datav1alpha1.RuntimePhaseNotReady,
					},
				}
				data := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: workerTestNamespaceEFC,
					},
				}

				s := runtime.NewScheme()
				s.AddKnownTypes(datav1alpha1.GroupVersion, efcRuntime)
				s.AddKnownTypes(datav1alpha1.GroupVersion, data)
				err := v1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				mockClient := fake.NewFakeClientWithScheme(s, efcRuntime, data)
				e := &EFCEngine{
					name:      "test1",
					namespace: workerTestNamespaceEFC,
					runtime:   efcRuntime,
					Client:    mockClient,
				}

				gotShould, err := e.ShouldSetupWorkers()
				Expect(err).NotTo(HaveOccurred())
				Expect(gotShould).To(BeFalse())
			})
		})

		Context("when WorkerPhase is RuntimePhasePartialReady", func() {
			It("should return false", func() {
				efcRuntime := &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: workerTestNamespaceEFC,
					},
					Status: datav1alpha1.RuntimeStatus{
						WorkerPhase: datav1alpha1.RuntimePhasePartialReady,
					},
				}
				data := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: workerTestNamespaceEFC,
					},
				}

				s := runtime.NewScheme()
				s.AddKnownTypes(datav1alpha1.GroupVersion, efcRuntime)
				s.AddKnownTypes(datav1alpha1.GroupVersion, data)
				err := v1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				mockClient := fake.NewFakeClientWithScheme(s, efcRuntime, data)
				e := &EFCEngine{
					name:      "test2",
					namespace: workerTestNamespaceEFC,
					runtime:   efcRuntime,
					Client:    mockClient,
				}

				gotShould, err := e.ShouldSetupWorkers()
				Expect(err).NotTo(HaveOccurred())
				Expect(gotShould).To(BeFalse())
			})
		})

		Context("when WorkerPhase is RuntimePhaseReady", func() {
			It("should return false", func() {
				efcRuntime := &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test3",
						Namespace: workerTestNamespaceEFC,
					},
					Status: datav1alpha1.RuntimeStatus{
						WorkerPhase: datav1alpha1.RuntimePhaseReady,
					},
				}
				data := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test3",
						Namespace: workerTestNamespaceEFC,
					},
				}

				s := runtime.NewScheme()
				s.AddKnownTypes(datav1alpha1.GroupVersion, efcRuntime)
				s.AddKnownTypes(datav1alpha1.GroupVersion, data)
				err := v1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				mockClient := fake.NewFakeClientWithScheme(s, efcRuntime, data)
				e := &EFCEngine{
					name:      "test3",
					namespace: workerTestNamespaceEFC,
					runtime:   efcRuntime,
					Client:    mockClient,
				}

				gotShould, err := e.ShouldSetupWorkers()
				Expect(err).NotTo(HaveOccurred())
				Expect(gotShould).To(BeFalse())
			})
		})
	})

	Describe("SetupWorkers", func() {
		var runtimeInfo base.RuntimeInfoInterface

		BeforeEach(func() {
			var err error
			runtimeInfo, err = base.BuildRuntimeInfo(workerTestNamespaceEFC, workerTestNamespaceFluid, common.EFCRuntime)
			Expect(err).NotTo(HaveOccurred())

			runtimeInfo.SetupWithDataset(&datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
			})

			nodeSelector := map[string]string{
				"node-select": "true",
			}
			runtimeInfo.SetFuseNodeSelector(nodeSelector)
		})

		Context("when setting up workers with replicas", func() {
			It("should setup workers successfully", func() {
				nodeInputs := []*v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-node",
						},
					},
				}
				worker := appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-worker",
						Namespace: workerTestNamespaceFluid,
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
				}
				efcRuntime := &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: workerTestNamespaceFluid,
					},
					Spec: datav1alpha1.EFCRuntimeSpec{
						Replicas: 1,
					},
				}
				data := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: workerTestNamespaceFluid,
					},
				}

				runtimeObjs := []runtime.Object{}
				for _, nodeInput := range nodeInputs {
					runtimeObjs = append(runtimeObjs, nodeInput.DeepCopy())
				}
				runtimeObjs = append(runtimeObjs, worker.DeepCopy())

				s := runtime.NewScheme()
				s.AddKnownTypes(datav1alpha1.GroupVersion, efcRuntime)
				s.AddKnownTypes(datav1alpha1.GroupVersion, data)
				s.AddKnownTypes(appsv1.SchemeGroupVersion, &worker)
				err := v1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				runtimeObjs = append(runtimeObjs, efcRuntime, data)
				mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)

				e := &EFCEngine{
					runtime:     efcRuntime,
					runtimeInfo: runtimeInfo,
					Client:      mockClient,
					name:        "test",
					namespace:   workerTestNamespaceFluid,
					Log:         ctrl.Log.WithName("test"),
				}

				e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)
				err = e.SetupWorkers()
				Expect(err).NotTo(HaveOccurred())

				workers, err := ctrlhelper.GetWorkersAsStatefulset(e.Client,
					types.NamespacedName{Namespace: e.namespace, Name: e.getWorkerName()})
				Expect(err).NotTo(HaveOccurred())
				Expect(*workers.Spec.Replicas).To(Equal(int32(1)))
			})
		})

		Context("when setting up workers with disabled worker", func() {
			It("should setup workers with 0 replicas", func() {
				nodeInputs := []*v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test1-node",
						},
					},
				}
				worker := appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1-worker",
						Namespace: workerTestNamespaceFluid,
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](0),
					},
				}
				efcRuntime := &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: workerTestNamespaceFluid,
					},
					Spec: datav1alpha1.EFCRuntimeSpec{
						Worker: datav1alpha1.EFCCompTemplateSpec{
							Disabled: true,
						},
					},
				}
				data := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: workerTestNamespaceFluid,
					},
				}

				runtimeObjs := []runtime.Object{}
				for _, nodeInput := range nodeInputs {
					runtimeObjs = append(runtimeObjs, nodeInput.DeepCopy())
				}
				runtimeObjs = append(runtimeObjs, worker.DeepCopy())

				s := runtime.NewScheme()
				s.AddKnownTypes(datav1alpha1.GroupVersion, efcRuntime)
				s.AddKnownTypes(datav1alpha1.GroupVersion, data)
				s.AddKnownTypes(appsv1.SchemeGroupVersion, &worker)
				err := v1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				runtimeObjs = append(runtimeObjs, efcRuntime, data)
				mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)

				e := &EFCEngine{
					runtime:     efcRuntime,
					runtimeInfo: runtimeInfo,
					Client:      mockClient,
					name:        "test1",
					namespace:   workerTestNamespaceFluid,
					Log:         ctrl.Log.WithName("test1"),
				}

				e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)
				err = e.SetupWorkers()
				Expect(err).NotTo(HaveOccurred())

				workers, err := ctrlhelper.GetWorkersAsStatefulset(e.Client,
					types.NamespacedName{Namespace: e.namespace, Name: e.getWorkerName()})
				Expect(err).NotTo(HaveOccurred())
				Expect(*workers.Spec.Replicas).To(Equal(int32(0)))
			})
		})
	})

	Describe("CheckWorkersReady", func() {
		Context("when all workers are ready", func() {
			It("should return true", func() {
				efcRuntime := &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test0",
						Namespace: workerTestNamespaceEFC,
					},
					Spec: datav1alpha1.EFCRuntimeSpec{
						Replicas: 1,
						Fuse:     datav1alpha1.EFCFuseSpec{},
					},
				}
				worker := &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test0-worker",
						Namespace: workerTestNamespaceEFC,
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      1,
						ReadyReplicas: 1,
					},
				}
				data := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test0",
						Namespace: workerTestNamespaceEFC,
					},
				}

				s := runtime.NewScheme()
				s.AddKnownTypes(datav1alpha1.GroupVersion, efcRuntime)
				s.AddKnownTypes(datav1alpha1.GroupVersion, data)
				s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
				err := v1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				mockClient := fake.NewFakeClientWithScheme(s, efcRuntime, data, worker)
				e := &EFCEngine{
					runtime:   efcRuntime,
					name:      "test0",
					namespace: workerTestNamespaceEFC,
					Client:    mockClient,
					Log:       ctrl.Log.WithName("test0"),
				}
				runtimeInfo, err := base.BuildRuntimeInfo("test0", workerTestNamespaceEFC, common.EFCRuntime)
				Expect(err).NotTo(HaveOccurred())

				e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)

				gotReady, err := e.CheckWorkersReady()
				Expect(err).NotTo(HaveOccurred())
				Expect(gotReady).To(BeTrue())
			})
		})

		Context("when workers are not ready", func() {
			It("should return false", func() {
				efcRuntime := &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: workerTestNamespaceEFC,
					},
					Spec: datav1alpha1.EFCRuntimeSpec{
						Replicas: 1,
						Fuse:     datav1alpha1.EFCFuseSpec{},
					},
				}
				worker := &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1-worker",
						Namespace: workerTestNamespaceEFC,
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      1,
						ReadyReplicas: 0,
					},
				}
				data := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: workerTestNamespaceEFC,
					},
				}

				s := runtime.NewScheme()
				s.AddKnownTypes(datav1alpha1.GroupVersion, efcRuntime)
				s.AddKnownTypes(datav1alpha1.GroupVersion, data)
				s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
				err := v1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				mockClient := fake.NewFakeClientWithScheme(s, efcRuntime, data, worker)
				e := &EFCEngine{
					runtime:   efcRuntime,
					name:      "test1",
					namespace: workerTestNamespaceEFC,
					Client:    mockClient,
					Log:       ctrl.Log.WithName("test1"),
				}
				runtimeInfo, err := base.BuildRuntimeInfo("test1", workerTestNamespaceEFC, common.EFCRuntime)
				Expect(err).NotTo(HaveOccurred())

				e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)

				gotReady, err := e.CheckWorkersReady()
				Expect(err).NotTo(HaveOccurred())
				Expect(gotReady).To(BeFalse())
			})
		})

		Context("when worker is disabled", func() {
			It("should return true", func() {
				efcRuntime := &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: workerTestNamespaceEFC,
					},
					Spec: datav1alpha1.EFCRuntimeSpec{
						Fuse: datav1alpha1.EFCFuseSpec{},
						Worker: datav1alpha1.EFCCompTemplateSpec{
							Disabled: true,
						},
					},
				}
				worker := &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2-worker",
						Namespace: workerTestNamespaceEFC,
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](0),
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      0,
						ReadyReplicas: 0,
					},
				}
				data := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: workerTestNamespaceEFC,
					},
				}

				s := runtime.NewScheme()
				s.AddKnownTypes(datav1alpha1.GroupVersion, efcRuntime)
				s.AddKnownTypes(datav1alpha1.GroupVersion, data)
				s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
				err := v1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				mockClient := fake.NewFakeClientWithScheme(s, efcRuntime, data, worker)
				e := &EFCEngine{
					runtime:   efcRuntime,
					name:      "test2",
					namespace: workerTestNamespaceEFC,
					Client:    mockClient,
					Log:       ctrl.Log.WithName("test2"),
				}
				runtimeInfo, err := base.BuildRuntimeInfo("test2", workerTestNamespaceEFC, common.EFCRuntime)
				Expect(err).NotTo(HaveOccurred())

				e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)

				gotReady, err := e.CheckWorkersReady()
				Expect(err).NotTo(HaveOccurred())
				Expect(gotReady).To(BeTrue())
			})
		})
	})

	Describe("getWorkerSelectors", func() {
		Context("when getting worker selectors", func() {
			It("should return correct selector string", func() {
				e := &EFCEngine{
					name: "spark",
				}
				got := e.getWorkerSelectors()
				Expect(got).To(Equal("app=efc,release=spark,role=efc-worker"))
			})
		})
	})

	Describe("syncWorkersEndpoints", func() {
		Context("when syncing workers endpoints successfully", func() {
			It("should return correct count", func() {
				worker := &appsv1.StatefulSet{
					TypeMeta: metav1.TypeMeta{
						Kind:       "StatefulSet",
						APIVersion: "apps/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-worker",
						Namespace: workerTestNamespaceBigData,
						UID:       "uid1",
					},
					Spec: appsv1.StatefulSetSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app":              "efc",
								"role":             "efc-worker",
								"fluid.io/dataset": "big-data-spark",
							},
						},
					},
				}
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-worker-0",
						Namespace: workerTestNamespaceBigData,
						OwnerReferences: []metav1.OwnerReference{{
							Kind:       "StatefulSet",
							APIVersion: "apps/v1",
							Name:       "spark-worker",
							UID:        "uid1",
							Controller: ptr.To(true),
						}},
						Labels: map[string]string{
							"app":              "efc",
							"role":             "efc-worker",
							"fluid.io/dataset": "big-data-spark",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "efc-worker",
								Ports: []v1.ContainerPort{
									{
										Name:          "rpc",
										ContainerPort: 7788,
									},
								},
							},
						},
					},
					Status: v1.PodStatus{
						PodIP: "127.0.0.1",
						Phase: v1.PodRunning,
						Conditions: []v1.PodCondition{{
							Type:   v1.PodReady,
							Status: v1.ConditionTrue,
						}},
					},
				}
				configMap := &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-worker-endpoints",
						Namespace: workerTestNamespaceBigData,
					},
					Data: map[string]string{
						WorkerEndpointsDataName: workerEndpointsConfigMapData,
					},
				}

				s := runtime.NewScheme()
				s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
				s.AddKnownTypes(v1.SchemeGroupVersion, configMap)
				err := v1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				mockClient := fake.NewFakeClientWithScheme(s, worker, configMap, pod)
				e := &EFCEngine{
					name:      "spark",
					namespace: workerTestNamespaceBigData,
					Client:    mockClient,
					Log:       ctrl.Log.WithName("spark"),
				}
				runtimeInfo, err := base.BuildRuntimeInfo("spark", workerTestNamespaceBigData, common.EFCRuntime)
				Expect(err).NotTo(HaveOccurred())

				e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)

				count, err := e.syncWorkersEndpoints()
				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(Equal(1))
			})
		})

		Context("when config map not found", func() {
			It("should return error", func() {
				worker := &appsv1.StatefulSet{
					TypeMeta: metav1.TypeMeta{
						Kind:       "StatefulSet",
						APIVersion: "apps/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-worker",
						Namespace: workerTestNamespaceBigData,
						UID:       "uid1",
					},
					Spec: appsv1.StatefulSetSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app":              "efc",
								"role":             "efc-worker",
								"fluid.io/dataset": "big-data-spark",
							},
						},
					},
				}
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-worker-0",
						Namespace: workerTestNamespaceBigData,
						OwnerReferences: []metav1.OwnerReference{{
							Kind:       "StatefulSet",
							APIVersion: "apps/v1",
							Name:       "spark-worker",
							UID:        "uid1",
							Controller: ptr.To(true),
						}},
						Labels: map[string]string{
							"app":              "efc",
							"role":             "efc-worker",
							"fluid.io/dataset": "big-data-spark",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "efc-worker",
								Ports: []v1.ContainerPort{
									{
										Name:          "rpc",
										ContainerPort: 7788,
									},
								},
							},
						},
					},
					Status: v1.PodStatus{
						PodIP: "127.0.0.1",
						Phase: v1.PodRunning,
						Conditions: []v1.PodCondition{{
							Type:   v1.PodReady,
							Status: v1.ConditionTrue,
						}},
					},
				}
				configMap := &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase-worker-endpoints",
						Namespace: workerTestNamespaceBigData,
					},
					Data: map[string]string{
						WorkerEndpointsDataName: workerEndpointsConfigMapData,
					},
				}

				s := runtime.NewScheme()
				s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
				s.AddKnownTypes(v1.SchemeGroupVersion, configMap)
				err := v1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				mockClient := fake.NewFakeClientWithScheme(s, worker, configMap, pod)
				e := &EFCEngine{
					name:      "spark",
					namespace: workerTestNamespaceBigData,
					Client:    mockClient,
					Log:       ctrl.Log.WithName("spark"),
				}
				runtimeInfo, err := base.BuildRuntimeInfo("spark", workerTestNamespaceBigData, common.EFCRuntime)
				Expect(err).NotTo(HaveOccurred())

				e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)

				count, err := e.syncWorkersEndpoints()
				Expect(err).To(HaveOccurred())
				Expect(count).To(Equal(1))
			})
		})
	})
})
