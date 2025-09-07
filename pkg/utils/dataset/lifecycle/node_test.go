/*
  Copyright 2023 The Fluid Authors.

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

package lifecycle

import (
	"context"
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dataset Lifecycle Node Tests", Label("pkg.utils.dataset.lifecycle.node_test.go"), func() {
	var (
		client      client.Client
		runtimeInfo base.RuntimeInfoInterface
		testScheme  *runtime.Scheme
		resources   []runtime.Object
	)

	BeforeEach(func() {
		testScheme = runtime.NewScheme()
		_ = corev1.AddToScheme(testScheme)
		_ = appsv1.AddToScheme(testScheme)
		_ = datav1alpha1.AddToScheme(testScheme)

		var err error
		runtimeInfo, err = base.BuildRuntimeInfo("hbase", "fluid", common.AlluxioRuntime)
		Expect(err).To(BeNil())
	})

	JustBeforeEach(func() {
		client = fake.NewFakeClientWithScheme(testScheme, resources...)
	})

	Describe("Test labelCacheNode()", func() {
		When("given exclusive placement mode runtime", func() {
			BeforeEach(func() {
				runtimeInfo.SetupWithDataset(&datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
				})

				resources = []runtime.Object{
					&corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-node-exclusive",
						},
					},
				}
			})

			It("should add exclusive labels to node", func() {
				node := corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-node-exclusive",
					},
				}

				err := labelCacheNode(node, runtimeInfo, client)
				Expect(err).To(BeNil())

				gotNode, err := kubeclient.GetNode(client, "test-node-exclusive")
				Expect(err).To(BeNil())
				Expect(gotNode.Labels).To(HaveKeyWithValue("fluid.io/dataset-num", "1"))
				Expect(gotNode.Labels).To(HaveKeyWithValue("fluid.io/s-alluxio-fluid-hbase", "true"))
				Expect(gotNode.Labels).To(HaveKeyWithValue("fluid.io/s-fluid-hbase", "true"))
				Expect(gotNode.Labels).To(HaveKeyWithValue("fluid_exclusive", "fluid_hbase"))
			})
		})

		When("given share placement mode runtime", func() {
			BeforeEach(func() {
				runtimeInfo.SetupWithDataset(&datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ShareMode},
				})

				resources = []runtime.Object{
					&corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-node-share",
						},
					},
				}
			})

			It("should add share mode labels to node", func() {
				node := corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-node-share",
					},
				}

				err := labelCacheNode(node, runtimeInfo, client)
				Expect(err).To(BeNil())

				gotNode, err := kubeclient.GetNode(client, "test-node-share")
				Expect(err).To(BeNil())
				Expect(gotNode.Labels).To(HaveKeyWithValue("fluid.io/dataset-num", "1"))
				Expect(gotNode.Labels).To(HaveKeyWithValue("fluid.io/s-alluxio-fluid-hbase", "true"))
				Expect(gotNode.Labels).To(HaveKeyWithValue("fluid.io/s-fluid-hbase", "true"))
				Expect(gotNode.Labels).NotTo(HaveKey("fluid_exclusive"))
			})
		})

		When("given runtime with tiered store", func() {
			BeforeEach(func() {
				tieredStore := datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{
						{
							MediumType: common.Memory,
							Quota:      resource.NewQuantity(1<<30, resource.BinarySI),
						},
						{
							MediumType: common.SSD,
							Quota:      resource.NewQuantity(2<<30, resource.BinarySI),
						},
						{
							MediumType: common.HDD,
							Quota:      resource.NewQuantity(3<<30, resource.BinarySI),
						},
					},
				}

				var err error
				runtimeInfo, err = base.BuildRuntimeInfo("spark", "fluid", "alluxio", base.WithTieredStore(tieredStore))
				Expect(err).To(BeNil())
				runtimeInfo.SetupWithDataset(&datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
				})

				resources = []runtime.Object{
					&corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-node-tiered",
						},
					},
				}
			})

			It("should add capacity labels to node", func() {
				node := corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-node-tiered",
					},
				}

				err := labelCacheNode(node, runtimeInfo, client)
				Expect(err).To(BeNil())

				gotNode, err := kubeclient.GetNode(client, "test-node-tiered")
				Expect(err).To(BeNil())
				Expect(gotNode.Labels).To(HaveKeyWithValue("fluid.io/s-h-alluxio-m-fluid-spark", "1GiB"))
				Expect(gotNode.Labels).To(HaveKeyWithValue("fluid.io/s-h-alluxio-d-fluid-spark", "5GiB"))
				Expect(gotNode.Labels).To(HaveKeyWithValue("fluid.io/s-h-alluxio-t-fluid-spark", "6GiB"))
			})
		})
	})

	Describe("Test DecreaseDatasetNum()", func() {
		When("node has dataset-num label with value 2", func() {
			It("should decrease dataset number to 1", func() {
				node := &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"fluid.io/dataset-num": "2"}},
				}
				runtimeInfo := &base.RuntimeInfo{}

				var labels common.LabelsToModify
				err := DecreaseDatasetNum(node, runtimeInfo, &labels)
				Expect(err).To(BeNil())

				// Check that the label modification is correct
				modifications := labels.GetLabels()
				Expect(modifications).To(HaveLen(1))
				Expect(modifications[0].GetLabelKey()).To(Equal("fluid.io/dataset-num"))
				Expect(modifications[0].GetLabelValue()).To(Equal("1"))
				Expect(modifications[0].GetOperationType()).To(Equal(common.UpdateLabel))
			})
		})

		When("node has dataset-num label with value 1", func() {
			It("should delete dataset-num label", func() {
				node := &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"fluid.io/dataset-num": "1"}},
				}
				runtimeInfo := &base.RuntimeInfo{}

				var labels common.LabelsToModify
				err := DecreaseDatasetNum(node, runtimeInfo, &labels)
				Expect(err).To(BeNil())

				// Check that the label is marked for deletion
				modifications := labels.GetLabels()
				Expect(modifications).To(HaveLen(1))
				Expect(modifications[0].GetLabelKey()).To(Equal("fluid.io/dataset-num"))
				Expect(modifications[0].GetOperationType()).To(Equal(common.DeleteLabel))
			})
		})

		When("node has invalid dataset-num label", func() {
			It("should return error", func() {
				node := &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"fluid.io/dataset-num": "invalid"}},
				}
				runtimeInfo := &base.RuntimeInfo{}

				var labels common.LabelsToModify
				err := DecreaseDatasetNum(node, runtimeInfo, &labels)
				Expect(err).NotTo(BeNil())
			})
		})

		When("node has no dataset-num label", func() {
			It("should not modify labels", func() {
				node := &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{},
				}
				runtimeInfo := &base.RuntimeInfo{}

				var labels common.LabelsToModify
				err := DecreaseDatasetNum(node, runtimeInfo, &labels)
				Expect(err).To(BeNil())
				Expect(labels.GetLabels()).To(BeEmpty())
			})
		})
	})

	Describe("Test increaseDatasetNum", func() {
		When("node has existing dataset-num label", func() {
			It("should increase dataset number", func() {
				node := &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"fluid.io/dataset-num": "1"}},
				}
				runtimeInfo := &base.RuntimeInfo{}

				var labels common.LabelsToModify
				err := increaseDatasetNum(node, runtimeInfo, &labels)
				Expect(err).To(BeNil())

				modifications := labels.GetLabels()
				Expect(modifications).To(HaveLen(1))
				Expect(modifications[0].GetLabelKey()).To(Equal("fluid.io/dataset-num"))
				Expect(modifications[0].GetLabelValue()).To(Equal("2"))
				Expect(modifications[0].GetOperationType()).To(Equal(common.UpdateLabel))
			})
		})

		When("node has no dataset-num label", func() {
			It("should add dataset-num label with value 1", func() {
				node := &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{},
				}
				runtimeInfo := &base.RuntimeInfo{}

				var labels common.LabelsToModify
				err := increaseDatasetNum(node, runtimeInfo, &labels)
				Expect(err).To(BeNil())

				modifications := labels.GetLabels()
				Expect(modifications).To(HaveLen(1))
				Expect(modifications[0].GetLabelKey()).To(Equal("fluid.io/dataset-num"))
				Expect(modifications[0].GetLabelValue()).To(Equal("1"))
				Expect(modifications[0].GetOperationType()).To(Equal(common.AddLabel))
			})
		})
	})

	Describe("Test hasRuntimeLabel", func() {
		When("node has runtime label", func() {
			It("should return true", func() {
				node := corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"fluid.io/s-alluxio-fluid-hbase": "true"},
					},
				}

				runtimeInfo, err := base.BuildRuntimeInfo("hbase", "fluid", "alluxio")
				Expect(err).To(BeNil())

				found := hasRuntimeLabel(node, runtimeInfo)
				Expect(found).To(BeTrue())
			})
		})

		When("node has no runtime label", func() {
			It("should return false", func() {
				node := corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"other-label": "value"},
					},
				}

				runtimeInfo, err := base.BuildRuntimeInfo("hbase", "fluid", "alluxio")
				Expect(err).To(BeNil())

				found := hasRuntimeLabel(node, runtimeInfo)
				Expect(found).To(BeFalse())
			})
		})

		When("node has no labels", func() {
			It("should return false", func() {
				node := corev1.Node{
					ObjectMeta: metav1.ObjectMeta{},
				}

				runtimeInfo, err := base.BuildRuntimeInfo("hbase", "fluid", "alluxio")
				Expect(err).To(BeNil())

				found := hasRuntimeLabel(node, runtimeInfo)
				Expect(found).To(BeFalse())
			})
		})
	})

	Describe("Test unlabelCacheNode()", func() {
		When("given exclusive placement mode runtime", func() {
			BeforeEach(func() {
				runtimeInfo.SetupWithDataset(&datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
				})

				resources = []runtime.Object{
					&corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-node-exclusive",
							Labels: map[string]string{
								"fluid.io/dataset-num":               "1",
								"fluid.io/s-alluxio-fluid-hbase":     "true",
								"fluid.io/s-fluid-hbase":             "true",
								"fluid.io/s-h-alluxio-t-fluid-hbase": "0B",
								"fluid_exclusive":                    "fluid_hbase",
								"test":                               "abc",
							},
						},
					},
				}
			})

			It("should remove fluid labels and keep other labels", func() {
				node := corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-node-exclusive",
						Labels: map[string]string{
							"fluid.io/dataset-num":               "1",
							"fluid.io/s-alluxio-fluid-hbase":     "true",
							"fluid.io/s-fluid-hbase":             "true",
							"fluid.io/s-h-alluxio-t-fluid-hbase": "0B",
							"fluid_exclusive":                    "fluid_hbase",
							"test":                               "abc",
						},
					},
				}

				err := unlabelCacheNode(node, runtimeInfo, client)
				Expect(err).To(BeNil())

				gotNode, err := kubeclient.GetNode(client, "test-node-exclusive")
				Expect(err).To(BeNil())
				Expect(gotNode.Labels).To(HaveKeyWithValue("test", "abc"))
				Expect(gotNode.Labels).NotTo(HaveKey("fluid.io/s-alluxio-fluid-hbase"))
				Expect(gotNode.Labels).NotTo(HaveKey("fluid_exclusive"))
			})
		})

		When("given share placement mode runtime", func() {
			BeforeEach(func() {
				runtimeInfo.SetupWithDataset(&datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ShareMode},
				})

				resources = []runtime.Object{
					&corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-node-share",
							Labels: map[string]string{
								"fluid.io/dataset-num":               "1",
								"fluid.io/s-alluxio-fluid-hbase":     "true",
								"fluid.io/s-fluid-hbase":             "true",
								"fluid.io/s-h-alluxio-t-fluid-hbase": "0B",
							},
						},
					},
				}
			})

			It("should remove all fluid labels", func() {
				node := corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-node-share",
						Labels: map[string]string{
							"fluid.io/dataset-num":               "1",
							"fluid.io/s-alluxio-fluid-spark":     "true",
							"fluid.io/s-fluid-spark":             "true",
							"fluid.io/s-h-alluxio-t-fluid-spark": "0B",
						},
					},
				}

				err := unlabelCacheNode(node, runtimeInfo, client)
				Expect(err).To(BeNil())

				gotNode, err := kubeclient.GetNode(client, "test-node-share")
				Expect(err).To(BeNil())
				Expect(gotNode.Labels).To(BeEmpty())
			})
		})

		When("given runtime with tiered store", func() {
			BeforeEach(func() {
				tieredStore := datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{
						{
							MediumType: common.Memory,
							Quota:      resource.NewQuantity(1, resource.BinarySI),
						},
						{
							MediumType: common.SSD,
							Quota:      resource.NewQuantity(2, resource.BinarySI),
						},
						{
							MediumType: common.HDD,
							Quota:      resource.NewQuantity(3, resource.BinarySI),
						},
					},
				}

				var err error
				runtimeInfo, err = base.BuildRuntimeInfo("spark", "fluid", "alluxio", base.WithTieredStore(tieredStore))
				Expect(err).To(BeNil())
				runtimeInfo.SetupWithDataset(&datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
				})

				resources = []runtime.Object{
					&corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-node-tiered",
							Labels: map[string]string{
								"fluid.io/dataset-num":               "1",
								"fluid.io/s-alluxio-fluid-spark":     "true",
								"fluid.io/s-fluid-spark":             "true",
								"fluid.io/s-h-alluxio-d-fluid-spark": "5B",
								"fluid.io/s-h-alluxio-m-fluid-spark": "1B",
								"fluid.io/s-h-alluxio-t-fluid-spark": "6B",
								"fluid_exclusive":                    "fluid_spark",
							},
						},
					},
				}
			})

			It("should remove all fluid labels including capacity labels", func() {
				node := corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-node-tiered",
						Labels: map[string]string{
							"fluid.io/dataset-num":               "1",
							"fluid.io/s-alluxio-fluid-spark":     "true",
							"fluid.io/s-fluid-spark":             "true",
							"fluid.io/s-h-alluxio-d-fluid-spark": "5B",
							"fluid.io/s-h-alluxio-m-fluid-spark": "1B",
							"fluid.io/s-h-alluxio-t-fluid-spark": "6B",
							"fluid_exclusive":                    "fluid_spark",
						},
					},
				}

				err := unlabelCacheNode(node, runtimeInfo, client)
				Expect(err).To(BeNil())

				gotNode, err := kubeclient.GetNode(client, "test-node-tiered")
				Expect(err).To(BeNil())
				Expect(gotNode.Labels).To(BeEmpty())
			})
		})
	})

	Describe("Test calculateNodeDifferences", func() {
		Context("when given current and previous node lists", func() {
			It("should calculate correct differences", func() {
				currentNodes := []string{"node-1", "node-2", "node-3"}
				previousNodes := []string{"node-2", "node-3", "node-4"}

				nodesToAdd, nodesToRemove := calculateNodeDifferences(currentNodes, previousNodes)

				Expect(nodesToAdd).To(Equal([]string{"node-1"}))
				Expect(nodesToRemove).To(Equal([]string{"node-4"}))
			})
		})

		Context("when given identical node lists", func() {
			It("should return empty differences", func() {
				currentNodes := []string{"node-1", "node-2"}
				previousNodes := []string{"node-1", "node-2"}

				nodesToAdd, nodesToRemove := calculateNodeDifferences(currentNodes, previousNodes)

				Expect(nodesToAdd).To(BeEmpty())
				Expect(nodesToRemove).To(BeEmpty())
			})
		})

		Context("when given completely different node lists", func() {
			It("should return all nodes as differences", func() {
				currentNodes := []string{"node-1", "node-2"}
				previousNodes := []string{"node-3", "node-4"}

				nodesToAdd, nodesToRemove := calculateNodeDifferences(currentNodes, previousNodes)

				Expect(nodesToAdd).To(Equal([]string{"node-1", "node-2"}))
				Expect(nodesToRemove).To(Equal([]string{"node-3", "node-4"}))
			})
		})
	})

	Describe("Test SyncScheduleInfoToCacheNodes", func() {
		var (
			nodes []*corev1.Node
			pods  []*corev1.Pod
			sts   *appsv1.StatefulSet
		)

		BeforeEach(func() {
			nodeNumToMock := 3
			for i := 0; i < nodeNumToMock; i++ {
				node := &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: fmt.Sprintf("node-%d", i),
					},
				}
				nodes = append(nodes, node)
			}

			// mock a pod for each node
			for i := 0; i < nodeNumToMock; i++ {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("hbase-worker-%d", i),
						Namespace: "fluid",
						Labels: map[string]string{
							"app": "hbase-worker",
						},
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "apps/v1",
								Kind:       "StatefulSet",
								Name:       "hbase-worker",
								UID:        "test-uid",
								Controller: ptr.To(true),
							},
						},
					},
					Spec: corev1.PodSpec{
						NodeName: fmt.Sprintf("node-%d", i),
					},
				}
				pods = append(pods, pod)
			}

			sts = &appsv1.StatefulSet{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps/v1",
					Kind:       "StatefulSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "fluid",
					Name:      "hbase-worker",
					UID:       "test-uid",
				},
				Spec: appsv1.StatefulSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "hbase-worker",
						},
					},
				},
			}
		})

		When("calling SyncScheduleInfoToCacheNodes several times to mock runtime scaling", func() {

			BeforeEach(func() {
				// 复用现有的 runtimeInfo 设置
				runtimeInfo.SetupWithDataset(&datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
				})

				resources = []runtime.Object{sts}
				for _, node := range nodes {
					resources = append(resources, node)
				}
			})

			It("should successfully add or remove schedule info on proper nodes", func() {
				By("scaling two pods to node-0 and node-1, the two nodes should have runtime labels", func() {
					Expect(client.Create(context.TODO(), pods[0])).To(Succeed())
					Expect(client.Create(context.TODO(), pods[1])).To(Succeed())

					Expect(SyncScheduleInfoToCacheNodes(runtimeInfo, client)).To(Succeed())

					gotNode0, err := kubeclient.GetNode(client, "node-0")
					Expect(err).To(BeNil())
					Expect(gotNode0).NotTo(BeNil())
					Expect(gotNode0.Labels).To(HaveKeyWithValue(runtimeInfo.GetRuntimeLabelName(), "true"))

					gotNode1, err := kubeclient.GetNode(client, "node-1")
					Expect(err).To(BeNil())
					Expect(gotNode1).NotTo(BeNil())
					Expect(gotNode1.Labels).To(HaveKeyWithValue(runtimeInfo.GetRuntimeLabelName(), "true"))

					gotNode2, err := kubeclient.GetNode(client, "node-2")
					Expect(err).To(BeNil())
					Expect(gotNode2).NotTo(BeNil())
					Expect(gotNode2.Labels).NotTo(HaveKey(runtimeInfo.GetRuntimeLabelName()))
				})

				By("scaling out a third pod to node-2, all pods should have runtime labels", func() {
					Expect(client.Create(context.TODO(), pods[2])).To(Succeed())

					Expect(SyncScheduleInfoToCacheNodes(runtimeInfo, client)).To(Succeed())

					for i := 0; i < 3; i++ {
						gotNode, err := kubeclient.GetNode(client, fmt.Sprintf("node-%d", i))
						Expect(err).To(BeNil())
						Expect(gotNode).NotTo(BeNil())
						Expect(gotNode.Labels).To(HaveKeyWithValue(runtimeInfo.GetRuntimeLabelName(), "true"))
					}
				})

				By("scaling in pods on node-1 and node-2, only node-0 should have runtime labels", func() {
					Expect(client.Delete(context.TODO(), pods[2])).To(Succeed())
					Expect(client.Delete(context.TODO(), pods[1])).To(Succeed())

					Expect(SyncScheduleInfoToCacheNodes(runtimeInfo, client)).To(Succeed())

					gotNode0, err := kubeclient.GetNode(client, "node-0")
					Expect(err).To(BeNil())
					Expect(gotNode0).NotTo(BeNil())
					Expect(gotNode0.Labels).To(HaveKeyWithValue(runtimeInfo.GetRuntimeLabelName(), "true"))

					gotNode1, err := kubeclient.GetNode(client, "node-1")
					Expect(err).To(BeNil())
					Expect(gotNode1).NotTo(BeNil())
					Expect(gotNode1.Labels).NotTo(HaveKey(runtimeInfo.GetRuntimeLabelName()))

					gotNode2, err := kubeclient.GetNode(client, "node-2")
					Expect(err).To(BeNil())
					Expect(gotNode2).NotTo(BeNil())
					Expect(gotNode2.Labels).NotTo(HaveKey(runtimeInfo.GetRuntimeLabelName()))
				})

				By("scaling in the last pod node node-0, now all nodes should not have runtime labels", func() {
					Expect(client.Delete(context.TODO(), pods[0])).To(Succeed())

					Expect(SyncScheduleInfoToCacheNodes(runtimeInfo, client)).To(Succeed())

					for i := 0; i < 3; i++ {
						gotNode, err := kubeclient.GetNode(client, fmt.Sprintf("node-%d", i))
						Expect(err).To(BeNil())
						Expect(gotNode).NotTo(BeNil())
						Expect(gotNode.Labels).NotTo(HaveKey(runtimeInfo.GetRuntimeLabelName()))
					}
				})
			})
		})
	})
})
