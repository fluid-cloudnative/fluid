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

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	testScheme *runtime.Scheme
)

var _ = Describe("Ctrl Helper Fuse Tests", Label("pkg.ctrl.fuse_test.go"), func() {
	var helper *Helper
	var resources []runtime.Object
	var fuseDs *appsv1.DaemonSet
	var k8sClient client.Client
	var runtimeInfo base.RuntimeInfoInterface

	BeforeEach(func() {
		fuseDs = mockRuntimeDaemonset("test-helper-fuse", "fluid")
		resources = []runtime.Object{
			fuseDs,
		}
	})

	JustBeforeEach(func() {
		k8sClient = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
		helper = BuildHelper(runtimeInfo, k8sClient, fake.NullLogger())
	})

	Describe("Test Helper.CheckAndUpdateFuseStatus()", func() {
		When("fuse ds is ready", func() {
			BeforeEach(func() {
				fuseDs.Status.DesiredNumberScheduled = 1
				fuseDs.Status.CurrentNumberScheduled = 1
				fuseDs.Status.NumberReady = 1
				fuseDs.Status.NumberAvailable = 1
				fuseDs.Status.NumberUnavailable = 0
			})

			When("applying to AlluxioRuntime", func() {
				var alluxioruntime *datav1alpha1.AlluxioRuntime
				BeforeEach(func() {
					alluxioruntime = &datav1alpha1.AlluxioRuntime{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-alluxio",
							Namespace: "fluid",
						},
						Spec:   datav1alpha1.AlluxioRuntimeSpec{},
						Status: datav1alpha1.RuntimeStatus{},
					}
					resources = append(resources, alluxioruntime)
				})

				It("should update AlluxioRuntime's status successfully", func() {
					getRuntimeFn := func(k8sClient client.Client) (base.RuntimeInterface, error) {
						runtime := &datav1alpha1.AlluxioRuntime{}
						err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, runtime)
						return runtime, err
					}

					ready, err := helper.CheckAndSyncFuseStatus(getRuntimeFn, types.NamespacedName{Namespace: fuseDs.Namespace, Name: fuseDs.Name})
					Expect(err).To(BeNil())
					Expect(ready).To(BeTrue())

					gotRuntime := &datav1alpha1.AlluxioRuntime{}
					err = k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, gotRuntime)
					Expect(err).To(BeNil())
					Expect(gotRuntime.Status.FusePhase).To(Equal(datav1alpha1.RuntimePhaseReady))
					Expect(gotRuntime.Status.DesiredFuseNumberScheduled).To(BeEquivalentTo(1))
					Expect(gotRuntime.Status.CurrentFuseNumberScheduled).To(BeEquivalentTo(1))
					Expect(gotRuntime.Status.FuseNumberReady).To(BeEquivalentTo(1))
					Expect(gotRuntime.Status.FuseNumberAvailable).To(BeEquivalentTo(1))
					Expect(gotRuntime.Status.FuseNumberUnavailable).To(BeEquivalentTo(0))

					Expect(gotRuntime.Status.Conditions).To(HaveLen(1))
					Expect(gotRuntime.Status.Conditions[0].Type).To(Equal(datav1alpha1.RuntimeFusesReady))
					Expect(gotRuntime.Status.Conditions[0].Status).To(Equal(corev1.ConditionTrue))
				})
			})
		})
	})

	Describe("CleanUpFuse", func() {
		var (
			testNodes        []runtime.Object
			fakeClient       client.Client
			h                *Helper
			name             string
			namespace        string
			runtimeType      string
			wantedNodeLabels map[string]map[string]string
			nodeInputs       []*corev1.Node
		)

		JustBeforeEach(func() {
			testNodes = []runtime.Object{}
			for _, nodeInput := range nodeInputs {
				testNodes = append(testNodes, nodeInput.DeepCopy())
			}

			fakeClient = fake.NewFakeClientWithScheme(testScheme, testNodes...)

			var err error
			runtimeInfo, err := base.BuildRuntimeInfo(name, namespace, runtimeType)
			Expect(err).NotTo(HaveOccurred())

			h = &Helper{
				runtimeInfo: runtimeInfo,
				client:      fakeClient,
				log:         fake.NullLogger(),
			}
		})

		Context("when cleaning up Jindo runtime", func() {
			BeforeEach(func() {
				name = "hbase"
				namespace = "fluid"
				runtimeType = "jindo"
				wantedNodeLabels = map[string]map[string]string{
					"no-fuse": {},
					"multiple-fuse": {
						"fluid.io/f-fluid-hadoop":          "true",
						"node-select":                      "true",
						"fluid.io/s-fluid-hbase":           "true",
						"fluid.io/s-h-jindo-d-fluid-hbase": "5B",
						"fluid.io/s-h-jindo-m-fluid-hbase": "1B",
						"fluid.io/s-h-jindo-t-fluid-hbase": "6B",
					},
					"fuse": {
						"fluid.io/dataset-num":    "1",
						"fluid.io/f-fluid-hadoop": "true",
						"node-select":             "true",
					},
				}
				nodeInputs = []*corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "no-fuse",
							Labels: map[string]string{},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "multiple-fuse",
							Labels: map[string]string{
								"fluid.io/f-fluid-hadoop":          "true",
								"node-select":                      "true",
								"fluid.io/f-fluid-hbase":           "true",
								"fluid.io/s-fluid-hbase":           "true",
								"fluid.io/s-h-jindo-d-fluid-hbase": "5B",
								"fluid.io/s-h-jindo-m-fluid-hbase": "1B",
								"fluid.io/s-h-jindo-t-fluid-hbase": "6B",
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "fuse",
							Labels: map[string]string{
								"fluid.io/dataset-num":    "1",
								"fluid.io/f-fluid-hadoop": "true",
								"node-select":             "true",
							},
						},
					},
				}
			})

			It("should clean up one fuse label", func() {
				count, err := h.CleanUpFuse()
				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(Equal(1))

				nodeList := &corev1.NodeList{}
				err = fakeClient.List(context.TODO(), nodeList, &client.ListOptions{})
				Expect(err).NotTo(HaveOccurred())

				for _, node := range nodeList.Items {
					Expect(node.Labels).To(HaveLen(len(wantedNodeLabels[node.Name])))
					if len(node.Labels) != 0 {
						Expect(node.Labels).To(Equal(wantedNodeLabels[node.Name]))
					}
				}
			})
		})

		Context("when cleaning up Alluxio runtime", func() {
			BeforeEach(func() {
				name = "spark"
				namespace = "fluid"
				runtimeType = "alluxio"
				wantedNodeLabels = map[string]map[string]string{
					"no-fuse": {},
					"multiple-fuse": {
						"node-select":                        "true",
						"fluid.io/s-fluid-hbase":             "true",
						"fluid.io/f-fluid-hbase":             "true",
						"fluid.io/s-h-alluxio-d-fluid-hbase": "5B",
						"fluid.io/s-h-alluxio-m-fluid-hbase": "1B",
						"fluid.io/s-h-alluxio-t-fluid-hbase": "6B",
					},
					"fuse": {
						"fluid.io/dataset-num": "1",
						"node-select":          "true",
					},
				}
				nodeInputs = []*corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "no-fuse",
							Labels: map[string]string{},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "multiple-fuse",
							Labels: map[string]string{
								"fluid.io/f-fluid-spark":             "true",
								"node-select":                        "true",
								"fluid.io/f-fluid-hbase":             "true",
								"fluid.io/s-fluid-hbase":             "true",
								"fluid.io/s-h-alluxio-d-fluid-hbase": "5B",
								"fluid.io/s-h-alluxio-m-fluid-hbase": "1B",
								"fluid.io/s-h-alluxio-t-fluid-hbase": "6B",
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "fuse",
							Labels: map[string]string{
								"fluid.io/dataset-num":   "1",
								"fluid.io/f-fluid-spark": "true",
								"node-select":            "true",
							},
						},
					},
				}
			})

			It("should clean up two fuse labels", func() {
				count, err := h.CleanUpFuse()
				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(Equal(2))

				nodeList := &corev1.NodeList{}
				err = fakeClient.List(context.TODO(), nodeList, &client.ListOptions{})
				Expect(err).NotTo(HaveOccurred())

				for _, node := range nodeList.Items {
					Expect(node.Labels).To(HaveLen(len(wantedNodeLabels[node.Name])))
					if len(node.Labels) != 0 {
						Expect(node.Labels).To(Equal(wantedNodeLabels[node.Name]))
					}
				}
			})
		})

		Context("when cleaning up GooseFS runtime with no matching labels", func() {
			BeforeEach(func() {
				name = "hbase"
				namespace = "fluid"
				runtimeType = "goosefs"
				wantedNodeLabels = map[string]map[string]string{
					"no-fuse": {},
					"multiple-fuse": {
						"fluid.io/f-fluid-spark":              "true",
						"node-select":                         "true",
						"fluid.io/s-fluid-hadoop":             "true",
						"fluid.io/f-fluid-hadoop":             "true",
						"fluid.io/s-h-goosefs-d-fluid-hadoop": "5B",
						"fluid.io/s-h-goosefs-m-fluid-hadoop": "1B",
						"fluid.io/s-h-goosefs-t-fluid-hadoop": "6B",
					},
					"fuse": {
						"fluid.io/dataset-num":   "1",
						"fluid.io/f-fluid-spark": "true",
						"node-select":            "true",
					},
				}
				nodeInputs = []*corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "no-fuse",
							Labels: map[string]string{},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "multiple-fuse",
							Labels: map[string]string{
								"fluid.io/f-fluid-spark":              "true",
								"node-select":                         "true",
								"fluid.io/f-fluid-hadoop":             "true",
								"fluid.io/s-fluid-hadoop":             "true",
								"fluid.io/s-h-goosefs-d-fluid-hadoop": "5B",
								"fluid.io/s-h-goosefs-m-fluid-hadoop": "1B",
								"fluid.io/s-h-goosefs-t-fluid-hadoop": "6B",
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "fuse",
							Labels: map[string]string{
								"fluid.io/dataset-num":   "1",
								"fluid.io/f-fluid-spark": "true",
								"node-select":            "true",
							},
						},
					},
				}
			})

			It("should not clean up any fuse labels", func() {
				count, err := h.CleanUpFuse()
				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(Equal(0))

				nodeList := &corev1.NodeList{}
				err = fakeClient.List(context.TODO(), nodeList, &client.ListOptions{})
				Expect(err).NotTo(HaveOccurred())

				for _, node := range nodeList.Items {
					Expect(node.Labels).To(HaveLen(len(wantedNodeLabels[node.Name])))
					if len(node.Labels) != 0 {
						Expect(node.Labels).To(Equal(wantedNodeLabels[node.Name]))
					}
				}
			})
		})

		Context("edge cases", func() {
			When("there are no nodes", func() {
				BeforeEach(func() {
					name = "test"
					namespace = "fluid"
					runtimeType = "jindo"
					nodeInputs = []*corev1.Node{}
					wantedNodeLabels = map[string]map[string]string{}
				})

				It("should return zero count without error", func() {
					count, err := h.CleanUpFuse()
					Expect(err).NotTo(HaveOccurred())
					Expect(count).To(Equal(0))
				})
			})

			When("all nodes have no labels", func() {
				BeforeEach(func() {
					name = "test"
					namespace = "fluid"
					runtimeType = "alluxio"
					nodeInputs = []*corev1.Node{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "node1",
								Labels: map[string]string{},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "node2",
								Labels: map[string]string{},
							},
						},
					}
					wantedNodeLabels = map[string]map[string]string{
						"node1": {},
						"node2": {},
					}
				})

				It("should return zero count without error", func() {
					count, err := h.CleanUpFuse()
					Expect(err).NotTo(HaveOccurred())
					Expect(count).To(Equal(0))

					nodeList := &corev1.NodeList{}
					err = fakeClient.List(context.TODO(), nodeList, &client.ListOptions{})
					Expect(err).NotTo(HaveOccurred())

					for _, node := range nodeList.Items {
						Expect(node.Labels).To(BeEmpty())
					}
				})
			})

			When("node has only the target fuse label", func() {
				BeforeEach(func() {
					name = "dataset"
					namespace = "default"
					runtimeType = "jindo"
					nodeInputs = []*corev1.Node{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "single-label-node",
								Labels: map[string]string{
									"fluid.io/f-default-dataset": "true",
								},
							},
						},
					}
					wantedNodeLabels = map[string]map[string]string{
						"single-label-node": {},
					}
				})

				It("should remove the single fuse label", func() {
					count, err := h.CleanUpFuse()
					Expect(err).NotTo(HaveOccurred())
					Expect(count).To(Equal(1))

					nodeList := &corev1.NodeList{}
					err = fakeClient.List(context.TODO(), nodeList, &client.ListOptions{})
					Expect(err).NotTo(HaveOccurred())

					Expect(nodeList.Items).To(HaveLen(1))
					Expect(nodeList.Items[0].Labels).To(BeEmpty())
				})
			})

			When("multiple nodes have the target fuse label", func() {
				BeforeEach(func() {
					name = "shared"
					namespace = "fluid"
					runtimeType = "alluxio"
					nodeInputs = []*corev1.Node{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "node1",
								Labels: map[string]string{
									"fluid.io/f-fluid-shared": "true",
									"keep-this":               "label",
								},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "node2",
								Labels: map[string]string{
									"fluid.io/f-fluid-shared": "true",
									"another-label":           "value",
								},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "node3",
								Labels: map[string]string{
									"fluid.io/f-fluid-shared": "true",
								},
							},
						},
					}
					wantedNodeLabels = map[string]map[string]string{
						"node1": {
							"keep-this": "label",
						},
						"node2": {
							"another-label": "value",
						},
						"node3": {},
					}
				})
			})

			When("nodes have mixed labels with different runtimes", func() {
				BeforeEach(func() {
					name = "target"
					namespace = "prod"
					runtimeType = "jindo"
					nodeInputs = []*corev1.Node{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "mixed-node",
								Labels: map[string]string{
									"fluid.io/f-prod-target":      "true",
									"fluid.io/f-prod-other":       "true",
									"fluid.io/s-prod-target":      "true",
									"fluid.io/s-prod-other":       "true",
									"fluid.io/dataset-num":        "2",
									"kubernetes.io/hostname":      "mixed-node",
									"node.kubernetes.io/instance": "m5.large",
								},
							},
						},
					}
					wantedNodeLabels = map[string]map[string]string{
						"mixed-node": {
							"fluid.io/f-prod-other":       "true",
							"fluid.io/s-prod-target":      "true",
							"fluid.io/s-prod-other":       "true",
							"fluid.io/dataset-num":        "2",
							"kubernetes.io/hostname":      "mixed-node",
							"node.kubernetes.io/instance": "m5.large",
						},
					}
				})

				It("should only remove the target runtime fuse label", func() {
					count, err := h.CleanUpFuse()
					Expect(err).NotTo(HaveOccurred())
					Expect(count).To(Equal(1))

					nodeList := &corev1.NodeList{}
					err = fakeClient.List(context.TODO(), nodeList, &client.ListOptions{})
					Expect(err).NotTo(HaveOccurred())

					Expect(nodeList.Items).To(HaveLen(1))
					Expect(nodeList.Items[0].Labels).To(Equal(wantedNodeLabels["mixed-node"]))
				})
			})
		})

		Context("with different runtime types", func() {
			When("using JuiceFS runtime", func() {
				BeforeEach(func() {
					name = "juicefs-test"
					namespace = "default"
					runtimeType = "juicefs"
					nodeInputs = []*corev1.Node{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "juicefs-node",
								Labels: map[string]string{
									"fluid.io/f-default-juicefs-test": "true",
									"other-label":                     "value",
								},
							},
						},
					}
					wantedNodeLabels = map[string]map[string]string{
						"juicefs-node": {
							"other-label": "value",
						},
					}
				})

				It("should clean up JuiceFS fuse labels", func() {
					count, err := h.CleanUpFuse()
					Expect(err).NotTo(HaveOccurred())
					Expect(count).To(Equal(1))

					nodeList := &corev1.NodeList{}
					err = fakeClient.List(context.TODO(), nodeList, &client.ListOptions{})
					Expect(err).NotTo(HaveOccurred())

					Expect(nodeList.Items[0].Labels).To(Equal(wantedNodeLabels["juicefs-node"]))
				})
			})

			When("using ThinRuntime", func() {
				BeforeEach(func() {
					name = "thin-test"
					namespace = "system"
					runtimeType = "thin"
					nodeInputs = []*corev1.Node{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "thin-node",
								Labels: map[string]string{
									"fluid.io/f-system-thin-test": "true",
								},
							},
						},
					}
					wantedNodeLabels = map[string]map[string]string{
						"thin-node": {},
					}
				})

				It("should clean up ThinRuntime fuse labels", func() {
					count, err := h.CleanUpFuse()
					Expect(err).NotTo(HaveOccurred())
					Expect(count).To(Equal(1))

					nodeList := &corev1.NodeList{}
					err = fakeClient.List(context.TODO(), nodeList, &client.ListOptions{})
					Expect(err).NotTo(HaveOccurred())

					Expect(nodeList.Items[0].Labels).To(BeEmpty())
				})
			})
		})
	})
})
