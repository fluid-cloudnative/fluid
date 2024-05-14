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
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	utilpointer "k8s.io/utils/pointer"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestCheckWorkerAffinity(t *testing.T) {

	s := runtime.NewScheme()
	name := "check-worker-affinity"
	namespace := "big-data"
	runtimeObjs := []runtime.Object{}
	mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)
	runtimeInfo, err := base.BuildRuntimeInfo(name, namespace, "jindo", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("testcase %s failed due to %v", name, err)
	}
	h := BuildHelper(runtimeInfo, mockClient, fake.NullLogger())

	tests := []struct {
		name   string
		worker *appsv1.StatefulSet
		want   bool
	}{
		{
			name: "no affinity",
			worker: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-affinity-worker",
					Namespace: namespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: utilpointer.Int32(1),
				},
			},
			want: false,
		}, {
			name: "no node affinity",
			worker: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-node-affinity-worker",
					Namespace: namespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: utilpointer.Int32(1),
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Affinity: &v1.Affinity{}}},
				},
			},
			want: false,
		}, {
			name: "other affinity exists",
			worker: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name + "-worker",
					Namespace: namespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: utilpointer.Int32(1),
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Affinity: &v1.Affinity{
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
						},
					},
				},
			},
			want: false,
		}, {
			name: "other affinity exists",
			worker: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name + "-worker",
					Namespace: namespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: utilpointer.Int32(1),
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Affinity: &v1.Affinity{
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
														Key:      "fluid.io/f-big-data-" + name,
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
					},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			found := h.checkWorkerAffinity(tt.worker)

			if found != tt.want {
				t.Errorf("Test case %s checkWorkerAffinity() = %v, want %v", tt.name, found, tt.want)
			}
		})
	}

}

func TestSetupWorkers(t *testing.T) {

	// runtimeInfoSpark tests create worker in exclusive mode.

	runtimeInfoSpark, err := base.BuildRuntimeInfo("spark", "big-data", "jindo", datav1alpha1.TieredStore{})

	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoSpark.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
	})

	// runtimeInfoSpark tests create worker in shareMode mode.
	runtimeInfoHadoop, err := base.BuildRuntimeInfo("hadoop", "big-data", "jindo", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoHadoop.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ShareMode},
	})
	nodeSelector := map[string]string{
		"node-select": "true",
	}
	runtimeInfoHadoop.SetFuseNodeSelector(nodeSelector)

	type fields struct {
		replicas    int32
		nodeInputs  []*v1.Node
		worker      appsv1.StatefulSet
		runtime     *datav1alpha1.JindoRuntime
		runtimeInfo base.RuntimeInfoInterface
		name        string
		namespace   string
	}
	tests := []struct {
		name             string
		fields           fields
		wantedNodeLabels map[string]map[string]string
	}{
		{
			name: "test0",
			fields: fields{
				replicas: 1,
				nodeInputs: []*v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-node-spark",
						},
					},
				},
				worker: appsv1.StatefulSet{

					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-jindofs-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: utilpointer.Int32(1),
					},
				},
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Replicas: 1,
					},
				},
				runtimeInfo: runtimeInfoSpark,
				name:        "spark",
				namespace:   "big-data",
			},
			wantedNodeLabels: map[string]map[string]string{
				"test-node-spark": {
					"fluid.io/dataset-num":                "1",
					"fluid.io/s-jindo-big-data-spark":     "true",
					"fluid.io/s-big-data-spark":           "true",
					"fluid.io/s-h-jindo-t-big-data-spark": "0B",
					"fluid_exclusive":                     "big-data_spark",
				},
			},
		},
		{
			name: "test1",
			fields: fields{
				replicas: 3,
				worker: appsv1.StatefulSet{

					ObjectMeta: metav1.ObjectMeta{
						Name:      "hadoop-jindofs-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: utilpointer.Int32(1),
					},
				},
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hadoop",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Replicas: 3,
					},
				},
				runtimeInfo: runtimeInfoHadoop,
				name:        "hadoop",
				namespace:   "big-data",
			},
			wantedNodeLabels: map[string]map[string]string{
				"test-node-hadoop": {
					"fluid.io/dataset-num":                 "1",
					"fluid.io/s-jindo-big-data-hadoop":     "true",
					"fluid.io/s-big-data-hadoop":           "true",
					"fluid.io/s-h-jindo-t-big-data-hadoop": "0B",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtimeObjs := []runtime.Object{}
			for _, nodeInput := range tt.fields.nodeInputs {
				runtimeObjs = append(runtimeObjs, nodeInput.DeepCopy())
			}
			runtimeObjs = append(runtimeObjs, tt.fields.worker.DeepCopy())

			s := runtime.NewScheme()
			data := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tt.fields.name,
					Namespace: tt.fields.namespace,
				},
			}
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.runtime)
			s.AddKnownTypes(datav1alpha1.GroupVersion, data)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, &tt.fields.worker)
			_ = v1.AddToScheme(s)
			runtimeObjs = append(runtimeObjs, tt.fields.runtime)
			runtimeObjs = append(runtimeObjs, data)
			mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)

			h := BuildHelper(tt.fields.runtimeInfo, mockClient, fake.NullLogger())

			err := h.SetupWorkers(tt.fields.runtime, tt.fields.runtime.Status, &tt.fields.worker)

			if err != nil {
				t.Errorf("test case %s h.SetupWorkers() error = %v", t.Name(), err)
			}

			worker := &appsv1.StatefulSet{}
			key := types.NamespacedName{
				Namespace: tt.fields.worker.Namespace,
				Name:      tt.fields.worker.Name,
			}

			err = mockClient.Get(context.TODO(), key, worker)
			if err != nil {
				t.Errorf("test case %s mockClient.Get() error = %v", t.Name(), err)
			}

			if tt.fields.replicas != *worker.Spec.Replicas {
				t.Errorf("Failed to scale %v for %v", tt.name, tt.fields)
			}

			// for _, node := range tt.fields.nodeInputs {
			// 	newNode, err := kubeclient.GetNode(mockClient, node.Name)
			// 	if err != nil {
			// 		t.Errorf("fail to get the node with the error %v", err)
			// 	}

			// 	if len(newNode.Labels) != len(tt.wantedNodeLabels[node.Name]) {
			// 		t.Errorf("fail to decrease the labels, newNode labels is %v", newNode.Labels)
			// 	}
			// 	if len(newNode.Labels) != 0 && !reflect.DeepEqual(newNode.Labels, tt.wantedNodeLabels[node.Name]) {
			// 		t.Errorf("fail to decrease the labels, newNode labels is %v", newNode.Labels)
			// 	}
			// }
		})
	}
}

func TestCheckWorkersReady(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.JindoRuntime
		worker    *appsv1.StatefulSet
		name      string
		namespace string
	}
	tests := []struct {
		name      string
		fields    fields
		wantReady bool
		wantPhase datav1alpha1.RuntimePhase
	}{
		{
			name: "ready",
			fields: fields{
				name:      "spark",
				namespace: "big-data",
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Replicas: 1,
						Fuse: datav1alpha1.JindoFuseSpec{
							Global: true,
						},
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-jindofs-worker",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
					},
				},
			},
			wantReady: true,
			wantPhase: datav1alpha1.RuntimePhaseReady,
		},
		{
			name: "noReady",
			fields: fields{
				name:      "hbase",
				namespace: "big-data",
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Replicas: 1,
						Fuse: datav1alpha1.JindoFuseSpec{
							Global: true,
						},
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase-jindofs-worker",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 0,
					},
				},
			},
			wantReady: false,
			wantPhase: datav1alpha1.RuntimePhaseNotReady,
		},
		{
			name: "partialReady",
			fields: fields{
				name:      "hbase",
				namespace: "big-data",
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Replicas: 2,
						Fuse: datav1alpha1.JindoFuseSpec{
							Global: true,
						},
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase-jindofs-worker",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
					},
				},
			},
			wantReady: true,
			wantPhase: datav1alpha1.RuntimePhasePartialReady,
		}, {
			name: "nochage",
			fields: fields{
				name:      "hbase",
				namespace: "big-data",
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Replicas: 2,
						Fuse: datav1alpha1.JindoFuseSpec{
							Global: true,
						},
					},
					Status: datav1alpha1.RuntimeStatus{
						WorkerPhase: datav1alpha1.RuntimePhasePartialReady,
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase-jindofs-worker",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
					},
				},
			},
			wantReady: true,
			wantPhase: datav1alpha1.RuntimePhasePartialReady,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtimeObjs := []runtime.Object{}
			data := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tt.fields.name,
					Namespace: tt.fields.namespace,
				},
			}

			s := runtime.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.runtime)
			s.AddKnownTypes(datav1alpha1.GroupVersion, data)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, tt.fields.worker)
			_ = v1.AddToScheme(s)

			runtimeObjs = append(runtimeObjs, tt.fields.runtime, data, tt.fields.worker)
			mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)
			// e := &jindo.JindoEngine{
			// 	runtime:   tt.fields.runtime,
			// 	name:      tt.fields.name,
			// 	namespace: tt.fields.namespace,
			// 	Client:    mockClient,
			// 	Log:       fake.NullLogger(),
			// }

			runtimeInfo, err := base.BuildRuntimeInfo(tt.fields.name, tt.fields.namespace, "jindo", datav1alpha1.TieredStore{})
			if err != nil {
				t.Errorf("testcase %s failed due to %v", tt.fields.name, err)
			}

			h := BuildHelper(runtimeInfo, mockClient, fake.NullLogger())

			gotReady, err := h.CheckWorkersReady(tt.fields.runtime, tt.fields.runtime.Status, tt.fields.worker)

			if err != nil {
				t.Errorf("CheckWorkersReady() got error %v", err)
			}

			if gotReady != tt.wantReady {
				t.Errorf("CheckWorkersReady() = %v, want %v", gotReady, tt.wantReady)
			}

			runtime := &datav1alpha1.JindoRuntime{}

			err = mockClient.Get(context.TODO(), types.NamespacedName{
				Namespace: tt.fields.namespace,
				Name:      tt.fields.name,
			}, runtime)

			if err != nil {
				t.Errorf("CheckWorkersReady() got error %v", err)
			}

			if runtime.Status.WorkerPhase != tt.wantPhase {
				t.Errorf("CheckWorkersReady() = %v, want %v", runtime.Status.WorkerPhase, tt.wantPhase)
			}
		})
	}
}
