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

package jindocache

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeschema "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"

	ctrlhelper "github.com/fluid-cloudnative/fluid/pkg/ctrl"
)

var _ = Describe("JindoCacheEngine", func() {
	var runtimeInfoSpark base.RuntimeInfoInterface
	var runtimeInfoHadoop base.RuntimeInfoInterface

	BeforeEach(func() {
		var err error
		runtimeInfoSpark, err = base.BuildRuntimeInfo("spark", "big-data", common.JindoRuntime)
		Expect(err).NotTo(HaveOccurred())
		runtimeInfoSpark.SetupWithDataset(&datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
		})

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

	Describe("SetupWorkers", func() {
		It("should setup workers in exclusive mode with node", func() {
			runtimeObjs := []runtimeschema.Object{}
			nodeInput := &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node-spark",
				},
			}
			runtimeObjs = append(runtimeObjs, nodeInput.DeepCopy())

			worker := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "spark-jindofs-worker",
					Namespace: "big-data",
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](1),
				},
			}
			runtimeObjs = append(runtimeObjs, worker.DeepCopy())

			runtime := &datav1alpha1.JindoRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "spark",
					Namespace: "big-data",
				},
				Spec: datav1alpha1.JindoRuntimeSpec{
					Replicas: 1,
				},
			}

			s := runtimeschema.NewScheme()
			data := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "spark",
					Namespace: "big-data",
				},
			}
			s.AddKnownTypes(datav1alpha1.GroupVersion, runtime)
			s.AddKnownTypes(datav1alpha1.GroupVersion, data)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
			_ = v1.AddToScheme(s)
			runtimeObjs = append(runtimeObjs, runtime)
			runtimeObjs = append(runtimeObjs, data)
			mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)

			e := &JindoCacheEngine{
				runtime:     runtime,
				runtimeInfo: runtimeInfoSpark,
				Client:      mockClient,
				name:        "spark",
				namespace:   "big-data",
				Log:         ctrl.Log.WithName("spark"),
			}

			e.Helper = ctrlhelper.BuildHelper(runtimeInfoSpark, mockClient, e.Log)
			err := e.SetupWorkers()
			Expect(err).NotTo(HaveOccurred())
			Expect(*worker.Spec.Replicas).To(Equal(int32(1)))
		})

		It("should setup workers in share mode without node", func() {
			runtimeObjs := []runtimeschema.Object{}

			worker := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hadoop-jindofs-worker",
					Namespace: "big-data",
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](1),
				},
			}
			runtimeObjs = append(runtimeObjs, worker.DeepCopy())

			runtime := &datav1alpha1.JindoRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hadoop",
					Namespace: "big-data",
				},
				Spec: datav1alpha1.JindoRuntimeSpec{
					Replicas: 1,
				},
			}

			s := runtimeschema.NewScheme()
			data := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hadoop",
					Namespace: "big-data",
				},
			}
			s.AddKnownTypes(datav1alpha1.GroupVersion, runtime)
			s.AddKnownTypes(datav1alpha1.GroupVersion, data)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
			_ = v1.AddToScheme(s)
			runtimeObjs = append(runtimeObjs, runtime)
			runtimeObjs = append(runtimeObjs, data)
			mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)

			e := &JindoCacheEngine{
				runtime:     runtime,
				runtimeInfo: runtimeInfoHadoop,
				Client:      mockClient,
				name:        "hadoop",
				namespace:   "big-data",
				Log:         ctrl.Log.WithName("hadoop"),
			}

			e.Helper = ctrlhelper.BuildHelper(runtimeInfoHadoop, mockClient, e.Log)
			err := e.SetupWorkers()
			Expect(err).NotTo(HaveOccurred())
			Expect(*worker.Spec.Replicas).To(Equal(int32(1)))
		})
	})

	Describe("ShouldSetupWorkers", func() {
		DescribeTable("should determine if workers need setup",
			func(name, namespace string, runtime *datav1alpha1.JindoRuntime, wantShould bool) {
				runtimeObjs := []runtimeschema.Object{}
				data := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace,
					},
				}

				s := runtimeschema.NewScheme()
				s.AddKnownTypes(datav1alpha1.GroupVersion, runtime)
				s.AddKnownTypes(datav1alpha1.GroupVersion, data)
				_ = v1.AddToScheme(s)
				runtimeObjs = append(runtimeObjs, runtime, data)
				mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)
				e := &JindoCacheEngine{
					name:      name,
					namespace: namespace,
					runtime:   runtime,
					Client:    mockClient,
				}

				gotShould, err := e.ShouldSetupWorkers()
				Expect(err).NotTo(HaveOccurred())
				Expect(gotShould).To(Equal(wantShould))
			},
			Entry("worker phase is none", "spark", "big-data",
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark",
						Namespace: "big-data",
					},
					Status: datav1alpha1.RuntimeStatus{
						WorkerPhase: datav1alpha1.RuntimePhaseNone,
					},
				}, true),
			Entry("worker phase is not ready", "hadoop", "big-data",
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hadoop",
						Namespace: "big-data",
					},
					Status: datav1alpha1.RuntimeStatus{
						WorkerPhase: datav1alpha1.RuntimePhaseNotReady,
					},
				}, false),
			Entry("worker phase is partial ready", "hbase", "big-data",
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase",
						Namespace: "big-data",
					},
					Status: datav1alpha1.RuntimeStatus{
						WorkerPhase: datav1alpha1.RuntimePhasePartialReady,
					},
				}, false),
			Entry("worker phase is ready", "tensorflow", "ml",
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tensorflow",
						Namespace: "ml",
					},
					Status: datav1alpha1.RuntimeStatus{
						WorkerPhase: datav1alpha1.RuntimePhaseReady,
					},
				}, false),
		)
	})

	Describe("CheckWorkersReady", func() {
		It("should return true when workers are ready", func() {
			runtime := &datav1alpha1.JindoRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "spark",
					Namespace: "big-data",
				},
				Spec: datav1alpha1.JindoRuntimeSpec{
					Replicas: 1,
					Fuse:     datav1alpha1.JindoFuseSpec{},
				},
			}
			worker := &appsv1.StatefulSet{
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
			}
			fuse := &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "spark-jindofs-fuse",
					Namespace: "big-data",
				},
				Status: appsv1.DaemonSetStatus{
					NumberAvailable:        1,
					DesiredNumberScheduled: 1,
					CurrentNumberScheduled: 1,
				},
			}

			runtimeObjs := []runtimeschema.Object{}
			data := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "spark",
					Namespace: "big-data",
				},
			}

			s := runtimeschema.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, runtime)
			s.AddKnownTypes(datav1alpha1.GroupVersion, data)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, fuse)
			_ = v1.AddToScheme(s)

			runtimeObjs = append(runtimeObjs, runtime, data, worker, fuse)
			mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)
			e := &JindoCacheEngine{
				runtime:   runtime,
				name:      "spark",
				namespace: "big-data",
				Client:    mockClient,
				Log:       ctrl.Log.WithName("spark"),
			}

			runtimeInfo, err := base.BuildRuntimeInfo("spark", "big-data", common.JindoRuntime)
			Expect(err).NotTo(HaveOccurred())

			e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)

			gotReady, err := e.CheckWorkersReady()
			Expect(err).NotTo(HaveOccurred())
			Expect(gotReady).To(BeTrue())
		})

		It("should return false when workers are not ready", func() {
			runtime := &datav1alpha1.JindoRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "big-data",
				},
				Spec: datav1alpha1.JindoRuntimeSpec{
					Replicas: 1,
					Fuse:     datav1alpha1.JindoFuseSpec{},
				},
			}
			worker := &appsv1.StatefulSet{
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
			}
			fuse := &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase-jindofs-fuse",
					Namespace: "big-data",
				},
				Status: appsv1.DaemonSetStatus{
					NumberAvailable:        0,
					DesiredNumberScheduled: 1,
					CurrentNumberScheduled: 0,
				},
			}

			runtimeObjs := []runtimeschema.Object{}
			data := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "big-data",
				},
			}

			s := runtimeschema.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, runtime)
			s.AddKnownTypes(datav1alpha1.GroupVersion, data)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, fuse)
			_ = v1.AddToScheme(s)

			runtimeObjs = append(runtimeObjs, runtime, data, worker, fuse)
			mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)
			e := &JindoCacheEngine{
				runtime:   runtime,
				name:      "hbase",
				namespace: "big-data",
				Client:    mockClient,
				Log:       ctrl.Log.WithName("hbase"),
			}

			runtimeInfo, err := base.BuildRuntimeInfo("hbase", "big-data", common.JindoRuntime)
			Expect(err).NotTo(HaveOccurred())

			e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)

			gotReady, err := e.CheckWorkersReady()
			Expect(err).NotTo(HaveOccurred())
			Expect(gotReady).To(BeFalse())
		})
	})

	Describe("getWorkerSelectors", func() {
		It("should return correct worker selectors", func() {
			e := &JindoCacheEngine{
				name: "spark",
			}
			got := e.getWorkerSelectors()
			Expect(got).To(Equal("app=jindo,release=spark,role=jindo-worker"))
		})
	})
})
