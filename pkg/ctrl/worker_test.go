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
	"errors"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ctrl Worker Tests Extended", Label("pkg.ctrl.worker_test.go"), func() {
	var helper *Helper
	var resources []runtime.Object
	var workerSts *appsv1.StatefulSet
	var k8sClient client.Client
	var runtimeInfo base.RuntimeInfoInterface

	BeforeEach(func() {
		workerSts = mockRuntimeStatefulset("test-helper-worker", "fluid")
		resources = []runtime.Object{
			workerSts,
		}
	})

	JustBeforeEach(func() {
		k8sClient = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
		helper = BuildHelper(runtimeInfo, k8sClient, fake.NullLogger())
	})

	Describe("Test Helper.CheckAndSyncWorkerStatus() - Extended Coverage", func() {
		When("worker sts is partially ready", func() {
			BeforeEach(func() {
				workerSts.Spec.Replicas = ptr.To[int32](3)
				workerSts.Status.AvailableReplicas = 2
				workerSts.Status.Replicas = 3
				workerSts.Status.ReadyReplicas = 2
			})

			When("applying to JindoRuntime", func() {
				var jindoruntime *datav1alpha1.JindoRuntime
				BeforeEach(func() {
					jindoruntime = &datav1alpha1.JindoRuntime{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-jindo",
							Namespace: "fluid",
						},
						Spec:   datav1alpha1.JindoRuntimeSpec{},
						Status: datav1alpha1.RuntimeStatus{},
					}
					resources = append(resources, jindoruntime)
				})

				It("should update runtime status to PartialReady", func() {
					getRuntimeFn := func(k8sClient client.Client) (base.RuntimeInterface, error) {
						runtime := &datav1alpha1.JindoRuntime{}
						err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: jindoruntime.Namespace, Name: jindoruntime.Name}, runtime)
						return runtime, err
					}

					ready, err := helper.CheckAndSyncWorkerStatus(getRuntimeFn, types.NamespacedName{Namespace: workerSts.Namespace, Name: workerSts.Name})
					Expect(err).To(BeNil())
					Expect(ready).To(BeTrue())

					gotRuntime := &datav1alpha1.JindoRuntime{}
					err = k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: jindoruntime.Namespace, Name: jindoruntime.Name}, gotRuntime)
					Expect(err).To(BeNil())
					Expect(gotRuntime.Status.WorkerPhase).To(Equal(datav1alpha1.RuntimePhasePartialReady))
					Expect(gotRuntime.Status.DesiredWorkerNumberScheduled).To(BeEquivalentTo(3))
					Expect(gotRuntime.Status.CurrentWorkerNumberScheduled).To(BeEquivalentTo(3))
					Expect(gotRuntime.Status.WorkerNumberReady).To(BeEquivalentTo(2))
					Expect(gotRuntime.Status.WorkerNumberAvailable).To(BeEquivalentTo(2))
					Expect(gotRuntime.Status.WorkerNumberUnavailable).To(BeEquivalentTo(1))
				})
			})
		})

		When("worker sts is not ready", func() {
			BeforeEach(func() {
				workerSts.Spec.Replicas = ptr.To[int32](3)
				workerSts.Status.AvailableReplicas = 0
				workerSts.Status.Replicas = 0
				workerSts.Status.ReadyReplicas = 0
			})

			When("applying to AlluxioRuntime", func() {
				var alluxioruntime *datav1alpha1.AlluxioRuntime
				BeforeEach(func() {
					alluxioruntime = &datav1alpha1.AlluxioRuntime{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-alluxio-notready",
							Namespace: "fluid",
						},
						Spec:   datav1alpha1.AlluxioRuntimeSpec{},
						Status: datav1alpha1.RuntimeStatus{},
					}
					resources = append(resources, alluxioruntime)
				})

				It("should update runtime status to NotReady", func() {
					getRuntimeFn := func(k8sClient client.Client) (base.RuntimeInterface, error) {
						runtime := &datav1alpha1.AlluxioRuntime{}
						err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, runtime)
						return runtime, err
					}

					ready, err := helper.CheckAndSyncWorkerStatus(getRuntimeFn, types.NamespacedName{Namespace: workerSts.Namespace, Name: workerSts.Name})
					Expect(err).To(BeNil())
					Expect(ready).To(BeFalse())

					gotRuntime := &datav1alpha1.AlluxioRuntime{}
					err = k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, gotRuntime)
					Expect(err).To(BeNil())
					Expect(gotRuntime.Status.WorkerPhase).To(Equal(datav1alpha1.RuntimePhaseNotReady))
					Expect(gotRuntime.Status.WorkerNumberUnavailable).To(BeEquivalentTo(0))

					Expect(gotRuntime.Status.Conditions).To(HaveLen(1))
					Expect(gotRuntime.Status.Conditions[0].Type).To(Equal(datav1alpha1.RuntimeWorkersReady))
					Expect(gotRuntime.Status.Conditions[0].Status).To(Equal(corev1.ConditionFalse))
				})
			})
		})

		When("worker sts has nil replicas", func() {
			BeforeEach(func() {
				workerSts.Spec.Replicas = nil
				workerSts.Status.AvailableReplicas = 0
				workerSts.Status.Replicas = 0
				workerSts.Status.ReadyReplicas = 0
			})

			When("applying to JindoRuntime", func() {
				var jindoruntime *datav1alpha1.JindoRuntime
				BeforeEach(func() {
					jindoruntime = &datav1alpha1.JindoRuntime{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-jindo-nil-replicas",
							Namespace: "fluid",
						},
						Spec:   datav1alpha1.JindoRuntimeSpec{},
						Status: datav1alpha1.RuntimeStatus{},
					}
					resources = append(resources, jindoruntime)
				})

				It("should handle nil replicas correctly", func() {
					getRuntimeFn := func(k8sClient client.Client) (base.RuntimeInterface, error) {
						runtime := &datav1alpha1.JindoRuntime{}
						err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: jindoruntime.Namespace, Name: jindoruntime.Name}, runtime)
						return runtime, err
					}

					ready, err := helper.CheckAndSyncWorkerStatus(getRuntimeFn, types.NamespacedName{Namespace: workerSts.Namespace, Name: workerSts.Name})
					Expect(err).To(BeNil())
					// FIX: When replicas is nil (0), and no replicas are ready, it should return true
					// because there are no workers expected, so the state is considered "ready"
					Expect(ready).To(BeTrue())

					gotRuntime := &datav1alpha1.JindoRuntime{}
					err = k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: jindoruntime.Namespace, Name: jindoruntime.Name}, gotRuntime)
					Expect(err).To(BeNil())
					Expect(gotRuntime.Status.DesiredWorkerNumberScheduled).To(BeEquivalentTo(0))
				})
			})
		})

		When("getRuntimeFn returns error", func() {
			BeforeEach(func() {
				workerSts.Spec.Replicas = ptr.To[int32](1)
			})

			It("should return error from getRuntimeFn", func() {
				getRuntimeFn := func(k8sClient client.Client) (base.RuntimeInterface, error) {
					return nil, errors.New("failed to get runtime")
				}

				ready, err := helper.CheckAndSyncWorkerStatus(getRuntimeFn, types.NamespacedName{Namespace: workerSts.Namespace, Name: workerSts.Name})
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to update worker ready status"))
				Expect(ready).To(BeFalse())
			})
		})

		When("worker statefulset not found", func() {
			It("should return error", func() {
				getRuntimeFn := func(k8sClient client.Client) (base.RuntimeInterface, error) {
					return &datav1alpha1.AlluxioRuntime{}, nil
				}

				ready, err := helper.CheckAndSyncWorkerStatus(getRuntimeFn, types.NamespacedName{Namespace: "nonexistent", Name: "nonexistent"})
				Expect(err).NotTo(BeNil())
				Expect(ready).To(BeFalse())
			})
		})

		When("runtime status has existing conditions", func() {
			var alluxioruntime *datav1alpha1.AlluxioRuntime
			BeforeEach(func() {
				workerSts.Spec.Replicas = ptr.To[int32](1)
				workerSts.Status.AvailableReplicas = 1
				workerSts.Status.Replicas = 1
				workerSts.Status.ReadyReplicas = 1

				alluxioruntime = &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-alluxio-existing-cond",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.AlluxioRuntimeSpec{},
					Status: datav1alpha1.RuntimeStatus{
						Conditions: []datav1alpha1.RuntimeCondition{
							{
								Type:   datav1alpha1.RuntimeWorkersInitialized,
								Status: corev1.ConditionTrue,
							},
						},
					},
				}
				resources = append(resources, alluxioruntime)
			})

			It("should update existing conditions", func() {
				getRuntimeFn := func(k8sClient client.Client) (base.RuntimeInterface, error) {
					runtime := &datav1alpha1.AlluxioRuntime{}
					err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, runtime)
					return runtime, err
				}

				ready, err := helper.CheckAndSyncWorkerStatus(getRuntimeFn, types.NamespacedName{Namespace: workerSts.Namespace, Name: workerSts.Name})
				Expect(err).To(BeNil())
				Expect(ready).To(BeTrue())

				gotRuntime := &datav1alpha1.AlluxioRuntime{}
				err = k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, gotRuntime)
				Expect(err).To(BeNil())
				Expect(gotRuntime.Status.Conditions).To(HaveLen(2))
			})
		})
	})

	Describe("Test Helper.SetupWorkers() - Extended Coverage", func() {
		var alluxioruntime *datav1alpha1.AlluxioRuntime
		var dataset *datav1alpha1.Dataset

		When("worker replicas need to be scaled up", func() {
			BeforeEach(func() {
				workerSts.Spec.Replicas = ptr.To[int32](1)

				// FIX: Create the Dataset object that the runtime depends on
				dataset = &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-alluxio-scale",
						Namespace: "fluid",
					},
				}

				alluxioruntime = &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-alluxio-scale",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Replicas: 3,
					},
					Status: datav1alpha1.RuntimeStatus{
						DesiredWorkerNumberScheduled: 1,
					},
				}
				runtimeInfo, _ = base.BuildRuntimeInfo("test-alluxio-scale", "fluid", common.AlluxioRuntime)
				resources = append(resources, dataset, alluxioruntime)
			})

			It("should scale up workers successfully", func() {
				err := helper.SetupWorkers(alluxioruntime, alluxioruntime.Status, workerSts)
				Expect(err).To(BeNil())

				updatedWorker := &appsv1.StatefulSet{}
				err = k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: workerSts.Namespace, Name: workerSts.Name}, updatedWorker)
				Expect(err).To(BeNil())
				Expect(*updatedWorker.Spec.Replicas).To(BeEquivalentTo(3))
			})
		})

		When("worker replicas are already at desired level", func() {
			BeforeEach(func() {
				workerSts.Spec.Replicas = ptr.To[int32](3)

				dataset = &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-alluxio-no-scale",
						Namespace: "fluid",
					},
				}

				alluxioruntime = &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-alluxio-no-scale",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Replicas: 3,
					},
					Status: datav1alpha1.RuntimeStatus{
						DesiredWorkerNumberScheduled: 3,
					},
				}
				runtimeInfo, _ = base.BuildRuntimeInfo("test-alluxio-no-scale", "fluid", common.AlluxioRuntime)
				resources = append(resources, dataset, alluxioruntime)
			})

			It("should not scale workers", func() {
				err := helper.SetupWorkers(alluxioruntime, alluxioruntime.Status, workerSts)
				Expect(err).To(BeNil())

				updatedWorker := &appsv1.StatefulSet{}
				err = k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: workerSts.Namespace, Name: workerSts.Name}, updatedWorker)
				Expect(err).To(BeNil())
				Expect(*updatedWorker.Spec.Replicas).To(BeEquivalentTo(3))
			})
		})

		When("worker replicas need to be scaled down", func() {
			BeforeEach(func() {
				workerSts.Spec.Replicas = ptr.To[int32](5)

				// FIX: Create the Dataset object
				dataset = &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-alluxio-scale-down",
						Namespace: "fluid",
					},
				}

				alluxioruntime = &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-alluxio-scale-down",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Replicas: 2,
					},
					Status: datav1alpha1.RuntimeStatus{
						DesiredWorkerNumberScheduled: 5,
					},
				}
				runtimeInfo, _ = base.BuildRuntimeInfo("test-alluxio-scale-down", "fluid", common.AlluxioRuntime)
				resources = append(resources, dataset, alluxioruntime)
			})

			It("should scale down workers successfully", func() {
				err := helper.SetupWorkers(alluxioruntime, alluxioruntime.Status, workerSts)
				Expect(err).To(BeNil())

				updatedWorker := &appsv1.StatefulSet{}
				err = k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: workerSts.Namespace, Name: workerSts.Name}, updatedWorker)
				Expect(err).To(BeNil())
				Expect(*updatedWorker.Spec.Replicas).To(BeEquivalentTo(2))
			})
		})

		When("worker replicas is nil", func() {
			BeforeEach(func() {
				workerSts.Spec.Replicas = nil

				// FIX: Create the Dataset object
				dataset = &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-alluxio-nil-replicas",
						Namespace: "fluid",
					},
				}

				alluxioruntime = &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-alluxio-nil-replicas",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Replicas: 2,
					},
					Status: datav1alpha1.RuntimeStatus{},
				}
				runtimeInfo, _ = base.BuildRuntimeInfo("test-alluxio-nil-replicas", "fluid", common.AlluxioRuntime)
				resources = append(resources, dataset, alluxioruntime)
			})

			It("should handle nil replicas and scale correctly", func() {
				err := helper.SetupWorkers(alluxioruntime, alluxioruntime.Status, workerSts)
				Expect(err).To(BeNil())

				updatedWorker := &appsv1.StatefulSet{}
				err = k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: workerSts.Namespace, Name: workerSts.Name}, updatedWorker)
				Expect(err).To(BeNil())
				Expect(*updatedWorker.Spec.Replicas).To(BeEquivalentTo(2))
			})
		})

		When("runtime status needs update", func() {
			BeforeEach(func() {
				workerSts.Spec.Replicas = ptr.To[int32](2)

				dataset = &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-alluxio-status-update",
						Namespace: "fluid",
					},
				}

				alluxioruntime = &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-alluxio-status-update",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Replicas: 2,
					},
					Status: datav1alpha1.RuntimeStatus{
						DesiredWorkerNumberScheduled: 1,
					},
				}
				runtimeInfo, _ = base.BuildRuntimeInfo("test-alluxio-status-update", "fluid", common.AlluxioRuntime)
				resources = append(resources, dataset, alluxioruntime)
			})

			It("should update runtime status with condition", func() {
				err := helper.SetupWorkers(alluxioruntime, alluxioruntime.Status, workerSts)
				Expect(err).To(BeNil())

				updatedRuntime := &datav1alpha1.AlluxioRuntime{}
				err = k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, updatedRuntime)
				Expect(err).To(BeNil())
				Expect(updatedRuntime.Status.Conditions).NotTo(BeEmpty())

				hasInitializedCondition := false
				for _, cond := range updatedRuntime.Status.Conditions {
					if cond.Type == datav1alpha1.RuntimeWorkersInitialized {
						hasInitializedCondition = true
						Expect(cond.Status).To(Equal(corev1.ConditionTrue))
					}
				}
				Expect(hasInitializedCondition).To(BeTrue())
			})
		})
	})

	Describe("Test Helper.TearDownWorkers() - Extended Coverage", func() {
		var node1, node2, node3 *corev1.Node
		var jindoRuntimeInfo base.RuntimeInfoInterface

		BeforeEach(func() {
			jindoRuntimeInfo, _ = base.BuildRuntimeInfo("test-jindo-teardown", "fluid", common.JindoRuntime)

			node1 = &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
					Labels: map[string]string{
						"fluid.io/s-fluid-test-jindo-teardown":       "true",
						"fluid.io/s-jindo-fluid-test-jindo-teardown": "true",
						"fluid.io/dataset-num":                       "1",
					},
				},
			}

			node2 = &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-2",
					Labels: map[string]string{
						"fluid.io/s-fluid-test-jindo-teardown":       "true",
						"fluid.io/s-jindo-fluid-test-jindo-teardown": "true",
						"fluid_exclusive":                            "fluid_test-jindo-teardown",
						"fluid.io/dataset-num":                       "1",
					},
				},
			}

			node3 = &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "node-3",
					Labels: map[string]string{},
				},
			}

			resources = []runtime.Object{node1, node2, node3}
		})

		When("tearing down workers with exclusive mode", func() {
			BeforeEach(func() {
				dataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-jindo-teardown",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.DatasetSpec{
						PlacementMode: datav1alpha1.ExclusiveMode,
					},
				}
				jindoRuntimeInfo.SetupWithDataset(dataset)
			})

			It("should remove worker labels including exclusive label", func() {
				err := helper.TearDownWorkers(jindoRuntimeInfo)
				Expect(err).To(BeNil())

				updatedNode2 := &corev1.Node{}
				err = k8sClient.Get(context.TODO(), types.NamespacedName{Name: "node-2"}, updatedNode2)
				Expect(err).To(BeNil())

				Expect(updatedNode2.Labels).NotTo(HaveKey("fluid_exclusive"))
				Expect(updatedNode2.Labels).NotTo(HaveKey("fluid.io/s-fluid-test-jindo-teardown"))
				Expect(updatedNode2.Labels).NotTo(HaveKey("fluid.io/s-jindo-fluid-test-jindo-teardown"))
			})
		})

		When("tearing down workers with share mode", func() {
			BeforeEach(func() {
				dataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-jindo-teardown",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.DatasetSpec{
						PlacementMode: datav1alpha1.ShareMode,
					},
				}
				jindoRuntimeInfo.SetupWithDataset(dataset)
			})

			It("should remove worker labels from all labeled nodes", func() {
				err := helper.TearDownWorkers(jindoRuntimeInfo)
				Expect(err).To(BeNil())

				updatedNode1 := &corev1.Node{}
				err = k8sClient.Get(context.TODO(), types.NamespacedName{Name: "node-1"}, updatedNode1)
				Expect(err).To(BeNil())
				Expect(updatedNode1.Labels).NotTo(HaveKey("fluid.io/s-fluid-test-jindo-teardown"))
			})
		})

		When("node has no labels", func() {
			It("should handle nodes without labels gracefully", func() {
				err := helper.TearDownWorkers(jindoRuntimeInfo)
				Expect(err).To(BeNil())

				updatedNode3 := &corev1.Node{}
				err = k8sClient.Get(context.TODO(), types.NamespacedName{Name: "node-3"}, updatedNode3)
				Expect(err).To(BeNil())
				Expect(updatedNode3.Labels).To(BeEmpty())
			})
		})
	})
})
