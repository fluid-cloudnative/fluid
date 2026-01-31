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

package jindo

import (
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"

	ctrlhelper "github.com/fluid-cloudnative/fluid/pkg/ctrl"
)

var _ = Describe("JindoEngine Worker", func() {
	Describe("SetupWorkers", func() {
		var runtimeInfoSpark base.RuntimeInfoInterface
		var runtimeInfoHadoop base.RuntimeInfoInterface

		BeforeEach(func() {
			var err error
			// runtimeInfoSpark tests create worker in exclusive mode.
			runtimeInfoSpark, err = base.BuildRuntimeInfo("spark", "big-data", common.JindoRuntime)
			Expect(err).NotTo(HaveOccurred())
			runtimeInfoSpark.SetupWithDataset(&datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
			})

			// runtimeInfoHadoop tests create worker in shareMode mode.
			runtimeInfoHadoop, err = base.BuildRuntimeInfo("hadoop", "big-data", common.JindoRuntime)
			Expect(err).NotTo(HaveOccurred())
			runtimeInfoHadoop.SetupWithDataset(&datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ShareMode},
			})
			nodeSelector := map[string]string{
				"node-select": "true",
			}
			runtimeInfoHadoop.SetFuseNodeSelector(nodeSelector)
		})

		DescribeTable("should set up workers correctly",
			func(testName string, replicas int32, nodeInputs []*v1.Node, worker *appsv1.StatefulSet,
				testRuntime *datav1alpha1.JindoRuntime, runtimeInfoGetter func() base.RuntimeInfoInterface,
				name, namespace string, deprecated bool) {

				runtimeInfo := runtimeInfoGetter()
				runtimeObjs := []runtime.Object{}
				for _, nodeInput := range nodeInputs {
					runtimeObjs = append(runtimeObjs, nodeInput.DeepCopy())
				}
				runtimeObjs = append(runtimeObjs, worker.DeepCopy())

				s := runtime.NewScheme()
				data := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace,
					},
				}
				s.AddKnownTypes(datav1alpha1.GroupVersion, testRuntime)
				s.AddKnownTypes(datav1alpha1.GroupVersion, data)
				s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
				_ = v1.AddToScheme(s)
				runtimeObjs = append(runtimeObjs, testRuntime)
				runtimeObjs = append(runtimeObjs, data)
				mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)

				e := &JindoEngine{
					runtime:     testRuntime,
					runtimeInfo: runtimeInfo,
					Client:      mockClient,
					name:        name,
					namespace:   namespace,
					Log:         ctrl.Log.WithName(name),
				}

				e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)
				err := e.SetupWorkers()
				Expect(err).NotTo(HaveOccurred())

				if !deprecated {
					Expect(*worker.Spec.Replicas).To(Equal(replicas))
				}
			},
			Entry("exclusive mode with node inputs",
				"test0",
				int32(1),
				[]*v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-node-spark",
						},
					},
				},
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-jindofs-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
				},
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Replicas: 1,
					},
				},
				func() base.RuntimeInfoInterface { return runtimeInfoSpark },
				"spark",
				"big-data",
				false,
			),
			Entry("share mode without node inputs",
				"test1",
				int32(1),
				[]*v1.Node{},
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hadoop-jindofs-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
				},
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hadoop",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Replicas: 1,
					},
				},
				func() base.RuntimeInfoInterface { return runtimeInfoHadoop },
				"hadoop",
				"big-data",
				false,
			),
		)
	})

	Describe("ShouldSetupWorkers", func() {
		DescribeTable("should return correct result based on worker phase",
			func(name, namespace string, testRuntime *datav1alpha1.JindoRuntime, wantShould bool) {
				runtimeObjs := []runtime.Object{}
				data := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace,
					},
				}

				s := runtime.NewScheme()
				s.AddKnownTypes(datav1alpha1.GroupVersion, testRuntime)
				s.AddKnownTypes(datav1alpha1.GroupVersion, data)
				_ = v1.AddToScheme(s)
				runtimeObjs = append(runtimeObjs, testRuntime, data)
				mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)
				e := &JindoEngine{
					name:      name,
					namespace: namespace,
					runtime:   testRuntime,
					Client:    mockClient,
				}

				gotShould, err := e.ShouldSetupWorkers()
				Expect(err).NotTo(HaveOccurred())
				Expect(gotShould).To(Equal(wantShould))
			},
			Entry("worker phase none should return true",
				"spark", "big-data",
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark",
						Namespace: "big-data",
					},
					Status: datav1alpha1.RuntimeStatus{
						WorkerPhase: datav1alpha1.RuntimePhaseNone,
					},
				},
				true,
			),
			Entry("worker phase not ready should return false",
				"hadoop", "big-data",
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hadoop",
						Namespace: "big-data",
					},
					Status: datav1alpha1.RuntimeStatus{
						WorkerPhase: datav1alpha1.RuntimePhaseNotReady,
					},
				},
				false,
			),
			Entry("worker phase partial ready should return false",
				"hbase", "big-data",
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase",
						Namespace: "big-data",
					},
					Status: datav1alpha1.RuntimeStatus{
						WorkerPhase: datav1alpha1.RuntimePhasePartialReady,
					},
				},
				false,
			),
			Entry("worker phase ready should return false",
				"tensorflow", "ml",
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tensorflow",
						Namespace: "ml",
					},
					Status: datav1alpha1.RuntimeStatus{
						WorkerPhase: datav1alpha1.RuntimePhaseReady,
					},
				},
				false,
			),
		)
	})

	Describe("CheckWorkersReady", func() {
		DescribeTable("should check workers ready correctly",
			func(name, namespace string, testRuntime *datav1alpha1.JindoRuntime,
				worker *appsv1.StatefulSet, fuse *appsv1.DaemonSet, wantReady bool) {

				runtimeObjs := []runtime.Object{}
				data := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace,
					},
				}

				s := runtime.NewScheme()
				s.AddKnownTypes(datav1alpha1.GroupVersion, testRuntime)
				s.AddKnownTypes(datav1alpha1.GroupVersion, data)
				s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
				s.AddKnownTypes(appsv1.SchemeGroupVersion, fuse)
				_ = v1.AddToScheme(s)

				runtimeObjs = append(runtimeObjs, testRuntime, data, worker, fuse)
				mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)
				e := &JindoEngine{
					runtime:   testRuntime,
					name:      name,
					namespace: namespace,
					Client:    mockClient,
					Log:       ctrl.Log.WithName(name),
				}

				runtimeInfo, err := base.BuildRuntimeInfo(name, namespace, common.JindoRuntime)
				Expect(err).NotTo(HaveOccurred())

				e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)

				gotReady, err := e.CheckWorkersReady()
				Expect(err).NotTo(HaveOccurred())
				Expect(gotReady).To(Equal(wantReady))
			},
			Entry("workers ready when replicas match",
				"spark", "big-data",
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Replicas: 1,
						Fuse:     datav1alpha1.JindoFuseSpec{},
					},
				},
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-jindofs-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
					},
				},
				&appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-jindofs-fuse",
						Namespace: "big-data",
					},
					Status: appsv1.DaemonSetStatus{
						NumberAvailable:        1,
						DesiredNumberScheduled: 1,
						CurrentNumberScheduled: 1,
					},
				},
				true,
			),
			Entry("workers not ready when replicas don't match",
				"hbase", "big-data",
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Replicas: 1,
						Fuse:     datav1alpha1.JindoFuseSpec{},
					},
				},
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase-jindofs-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 0,
					},
				},
				&appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase-jindofs-fuse",
						Namespace: "big-data",
					},
					Status: appsv1.DaemonSetStatus{
						NumberAvailable:        0,
						DesiredNumberScheduled: 1,
						CurrentNumberScheduled: 0,
					},
				},
				false,
			),
		)
	})

	Describe("getWorkerSelectors", func() {
		It("should return correct worker selectors", func() {
			e := &JindoEngine{
				name: "spark",
			}
			got := e.getWorkerSelectors()
			Expect(got).To(Equal("app=jindo,release=spark,role=jindo-worker"))
		})
	})

	Describe("buildWorkersAffinity", func() {
		DescribeTable("should build workers affinity correctly",
			func(dataset *datav1alpha1.Dataset, worker *appsv1.StatefulSet, want *v1.Affinity) {
				s := runtime.NewScheme()
				s.AddKnownTypes(datav1alpha1.GroupVersion, dataset)
				s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
				_ = v1.AddToScheme(s)
				runtimeObjs := []runtime.Object{}
				runtimeObjs = append(runtimeObjs, dataset)
				runtimeObjs = append(runtimeObjs, worker)
				mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)
				runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "jindo")
				Expect(err).NotTo(HaveOccurred())

				e := &JindoEngine{
					name:        dataset.Name,
					namespace:   dataset.Namespace,
					Client:      mockClient,
					runtimeInfo: runtimeInfo,
				}

				resultWorker, err := e.buildWorkersAffinity(worker)
				Expect(err).NotTo(HaveOccurred())
				Expect(reflect.DeepEqual(resultWorker.Spec.Template.Spec.Affinity, want)).To(BeTrue())
			},
			Entry("exclusive mode",
				&datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.DatasetSpec{
						PlacementMode: datav1alpha1.ExclusiveMode,
					},
				},
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1-jindofs-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
				},
				&v1.Affinity{
					PodAntiAffinity: &v1.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
							{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      "fluid.io/dataset",
											Operator: metav1.LabelSelectorOpExists,
										},
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
					NodeAffinity: &v1.NodeAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
							{
								Weight: 100,
								Preference: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "fluid.io/f-big-data-test1",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"true"},
										},
									},
								},
							},
						},
					},
				},
			),
			Entry("shared mode",
				&datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.DatasetSpec{
						PlacementMode: datav1alpha1.ShareMode,
					},
				},
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2-jindofs-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
				},
				&v1.Affinity{
					PodAntiAffinity: &v1.PodAntiAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
							{
								Weight: 50,
								PodAffinityTerm: v1.PodAffinityTerm{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "fluid.io/dataset",
												Operator: metav1.LabelSelectorOpExists,
											},
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
						RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
							{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      "fluid.io/dataset-placement",
											Operator: metav1.LabelSelectorOpIn,
											Values:   []string{"Exclusive"},
										},
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
					NodeAffinity: &v1.NodeAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
							{
								Weight: 100,
								Preference: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "fluid.io/f-big-data-test2",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"true"},
										},
									},
								},
							},
						},
					},
				},
			),
			Entry("dataset with affinity",
				&datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test3",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.DatasetSpec{
						NodeAffinity: &datav1alpha1.CacheableNodeAffinity{
							Required: &v1.NodeSelector{
								NodeSelectorTerms: []v1.NodeSelectorTerm{
									{
										MatchExpressions: []v1.NodeSelectorRequirement{
											{
												Key:      "nodeA",
												Operator: v1.NodeSelectorOpIn,
												Values:   []string{"true"},
											},
										},
									},
								},
							},
						},
					},
				},
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test3-jindofs-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
				},
				&v1.Affinity{
					PodAntiAffinity: &v1.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
							{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      "fluid.io/dataset",
											Operator: metav1.LabelSelectorOpExists,
										},
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
					NodeAffinity: &v1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
							NodeSelectorTerms: []v1.NodeSelectorTerm{
								{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "nodeA",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"true"},
										},
									},
								},
							},
						},
						PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
							{
								Weight: 100,
								Preference: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "fluid.io/f-big-data-test3",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"true"},
										},
									},
								},
							},
						},
					},
				},
			),
		)
	})
})
