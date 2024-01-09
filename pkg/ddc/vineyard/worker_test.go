/*
Copyright 2024 The Fluid Authors.
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

package vineyard

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	ctrlhelper "github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilpointer "k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestSetupWorkers(t *testing.T) {

	// runtimeInfoSpark tests create worker in exclusive mode.

	runtimeInfoSpark, err := base.BuildRuntimeInfo("spark", "big-data", "vineyard", datav1alpha1.TieredStore{})

	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoSpark.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
	})

	// runtimeInfoSpark tests create worker in shareMode mode.
	runtimeInfoHadoop, err := base.BuildRuntimeInfo("hadoop", "big-data", "vineyard", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoHadoop.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ShareMode},
	})
	nodeSelector := map[string]string{
		"node-select": "true",
	}
	runtimeInfoHadoop.SetupFuseDeployMode(true, nodeSelector)

	type fields struct {
		replicas         int32
		nodeInputs       []*v1.Node
		worker           *appsv1.StatefulSet
		deprecatedWorker *appsv1.DaemonSet
		runtime          *datav1alpha1.VineyardRuntime
		runtimeInfo      base.RuntimeInfoInterface
		name             string
		namespace        string
		deprecated       bool
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
				worker: &appsv1.StatefulSet{

					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: utilpointer.Int32Ptr(1),
					},
				},
				runtime: &datav1alpha1.VineyardRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Worker: datav1alpha1.VineyardCompTemplateSpec{
							Replicas: 1,
						},
					},
				},
				runtimeInfo: runtimeInfoSpark,
				name:        "spark",
				namespace:   "big-data",
			},
			wantedNodeLabels: map[string]map[string]string{
				"test-node-spark": {
					"fluid.io/dataset-num":                   "1",
					"fluid.io/s-vineyard-big-data-spark":     "true",
					"fluid.io/s-big-data-spark":              "true",
					"fluid.io/s-h-vineyard-t-big-data-spark": "0B",
					"fluid_exclusive":                        "big-data_spark",
				},
			},
		},
		{
			name: "test1",
			fields: fields{
				replicas: 1,
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hadoop-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: utilpointer.Int32Ptr(1),
					},
				},
				runtime: &datav1alpha1.VineyardRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hadoop",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Worker: datav1alpha1.VineyardCompTemplateSpec{
							Replicas: 1,
						},
					},
				},
				runtimeInfo: runtimeInfoHadoop,
				name:        "hadoop",
				namespace:   "big-data",
			},
			wantedNodeLabels: map[string]map[string]string{
				"test-node-hadoop": {
					"fluid.io/dataset-num":                    "1",
					"fluid.io/s-vineyard-big-data-hadoop":     "true",
					"fluid.io/s-big-data-hadoop":              "true",
					"fluid.io/s-h-vineyard-t-big-data-hadoop": "0B",
				},
			},
		}, {
			name: "deprecated",
			fields: fields{
				replicas: 0,
				worker:   &appsv1.StatefulSet{},
				deprecatedWorker: &appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{
					Name:      "deprecated-worker",
					Namespace: "big-data",
				}},
				runtime: &datav1alpha1.VineyardRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "deprecated",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Worker: datav1alpha1.VineyardCompTemplateSpec{
							Replicas: 1,
						},
					},
				},
				runtimeInfo: runtimeInfoHadoop,
				name:        "deprecated",
				namespace:   "big-data",
				deprecated:  true,
			},
			wantedNodeLabels: map[string]map[string]string{
				"test-node-hadoop": {
					"fluid.io/dataset-num":                    "1",
					"fluid.io/s-vineyard-big-data-hadoop":     "true",
					"fluid.io/s-big-data-hadoop":              "true",
					"fluid.io/s-h-vineyard-t-big-data-hadoop": "0B",
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
			s.AddKnownTypes(appsv1.SchemeGroupVersion, tt.fields.worker)
			if tt.fields.deprecatedWorker != nil {
				s.AddKnownTypes(appsv1.SchemeGroupVersion, tt.fields.deprecatedWorker)
			}
			_ = v1.AddToScheme(s)
			runtimeObjs = append(runtimeObjs, tt.fields.runtime)
			if tt.fields.deprecatedWorker != nil {
				runtimeObjs = append(runtimeObjs, tt.fields.deprecatedWorker)
			}
			runtimeObjs = append(runtimeObjs, data)
			mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)

			e := &VineyardEngine{
				runtime:     tt.fields.runtime,
				runtimeInfo: tt.fields.runtimeInfo,
				Client:      mockClient,
				name:        tt.fields.name,
				namespace:   tt.fields.namespace,
				Log:         ctrl.Log.WithName(tt.fields.name),
			}

			e.Helper = ctrlhelper.BuildHelper(tt.fields.runtimeInfo, mockClient, e.Log)
			err := e.SetupWorkers()
			if err != nil {
				t.Errorf("testCase %s VineyardEngine.SetupWorkers() error = %v", tt.name, err)
			}

			if !tt.fields.deprecated {
				if tt.fields.replicas != *tt.fields.worker.Spec.Replicas {
					t.Errorf("Failed to scale %v for %v", tt.name, tt.fields)
				}
			}
		})
	}
}

func TestShouldSetupWorkers(t *testing.T) {
	type fields struct {
		name      string
		namespace string
		runtime   *datav1alpha1.VineyardRuntime
	}
	tests := []struct {
		name       string
		fields     fields
		wantShould bool
		wantErr    bool
	}{
		{
			name: "test0",
			fields: fields{
				name:      "spark",
				namespace: "big-data",
				runtime: &datav1alpha1.VineyardRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark",
						Namespace: "big-data",
					},
					Status: datav1alpha1.RuntimeStatus{
						WorkerPhase: datav1alpha1.RuntimePhaseNone,
					},
				},
			},
			wantShould: true,
			wantErr:    false,
		},
		{
			name: "test1",
			fields: fields{
				name:      "hadoop",
				namespace: "big-data",
				runtime: &datav1alpha1.VineyardRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hadoop",
						Namespace: "big-data",
					},
					Status: datav1alpha1.RuntimeStatus{
						WorkerPhase: datav1alpha1.RuntimePhaseNotReady,
					},
				},
			},
			wantShould: false,
			wantErr:    false,
		},
		{
			name: "test2",
			fields: fields{
				name:      "hbase",
				namespace: "big-data",
				runtime: &datav1alpha1.VineyardRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase",
						Namespace: "big-data",
					},
					Status: datav1alpha1.RuntimeStatus{
						WorkerPhase: datav1alpha1.RuntimePhasePartialReady,
					},
				},
			},
			wantShould: false,
			wantErr:    false,
		},
		{
			name: "test3",
			fields: fields{
				name:      "tensorflow",
				namespace: "ml",
				runtime: &datav1alpha1.VineyardRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tensorflow",
						Namespace: "ml",
					},
					Status: datav1alpha1.RuntimeStatus{
						WorkerPhase: datav1alpha1.RuntimePhaseReady,
					},
				},
			},
			wantShould: false,
			wantErr:    false,
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
			_ = v1.AddToScheme(s)
			runtimeObjs = append(runtimeObjs, tt.fields.runtime, data)
			mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)
			e := &VineyardEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				runtime:   tt.fields.runtime,
				Client:    mockClient,
			}

			gotShould, err := e.ShouldSetupWorkers()
			if (err != nil) != tt.wantErr {
				t.Errorf("VineyardEngine.ShouldSetupWorkers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotShould != tt.wantShould {
				t.Errorf("VineyardEngine.ShouldSetupWorkers() = %v, want %v", gotShould, tt.wantShould)
			}
		})
	}
}

func TestCheckWorkersReady(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.VineyardRuntime
		worker    *appsv1.StatefulSet
		fuse      *appsv1.DaemonSet
		name      string
		namespace string
	}
	tests := []struct {
		name      string
		fields    fields
		wantReady bool
		wantErr   bool
	}{
		{
			name: "test0",
			fields: fields{
				name:      "spark",
				namespace: "big-data",
				runtime: &datav1alpha1.VineyardRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Worker: datav1alpha1.VineyardCompTemplateSpec{
							Replicas: 1,
						},
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-worker",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
					},
				},
				fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-fuse",
						Namespace: "big-data",
					},
					Status: appsv1.DaemonSetStatus{
						NumberAvailable:        1,
						DesiredNumberScheduled: 1,
						CurrentNumberScheduled: 1,
					},
				},
			},
			wantReady: true,
			wantErr:   false,
		},
		{
			name: "test1",
			fields: fields{
				name:      "hbase",
				namespace: "big-data",
				runtime: &datav1alpha1.VineyardRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Worker: datav1alpha1.VineyardCompTemplateSpec{
							Replicas: 1,
						},
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase-worker",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 0,
					},
				},
				fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase-fuse",
						Namespace: "big-data",
					},
					Status: appsv1.DaemonSetStatus{
						NumberAvailable:        0,
						DesiredNumberScheduled: 1,
						CurrentNumberScheduled: 0,
					},
				},
			},
			wantReady: false,
			wantErr:   false,
		}, {
			name: "deprecated",
			fields: fields{
				name:      "deprecated",
				namespace: "big-data",
				runtime: &datav1alpha1.VineyardRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "deprecated",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Worker: datav1alpha1.VineyardCompTemplateSpec{
							Replicas: 1,
						},
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "deprecated-worker-0",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 0,
					},
				},
				fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "deprecated-worker",
						Namespace: "big-data",
					},
					Status: appsv1.DaemonSetStatus{
						NumberAvailable:        0,
						DesiredNumberScheduled: 1,
						CurrentNumberScheduled: 0,
					},
				},
			},
			wantReady: true,
			wantErr:   false,
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
			s.AddKnownTypes(appsv1.SchemeGroupVersion, tt.fields.fuse)
			_ = v1.AddToScheme(s)

			runtimeObjs = append(runtimeObjs, tt.fields.runtime, data, tt.fields.worker, tt.fields.fuse)
			mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)
			e := &VineyardEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
				Log:       ctrl.Log.WithName(tt.fields.name),
			}

			runtimeInfo, err := base.BuildRuntimeInfo(tt.fields.name, tt.fields.namespace, "vineyard", datav1alpha1.TieredStore{})
			if err != nil {
				t.Errorf("VineyardEngine.CheckWorkersReady() error = %v", err)
			}

			e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)

			gotReady, err := e.CheckWorkersReady()
			if (err != nil) != tt.wantErr {
				t.Errorf("VineyardEngine.CheckWorkersReady() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotReady != tt.wantReady {
				t.Errorf("VineyardEngine.CheckWorkersReady() = %v, want %v", gotReady, tt.wantReady)
			}
		})
	}
}

func TestGetWorkerSelectors(t *testing.T) {
	type fields struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test0",
			fields: fields{
				name: "spark",
			},
			want: "app=vineyard,release=spark,role=vineyard-worker",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &VineyardEngine{
				name: tt.fields.name,
			}
			if got := e.getWorkerSelectors(); got != tt.want {
				t.Errorf("VineyardEngine.getWorkerSelectors() = %v, want %v", got, tt.want)
			}
		})
	}
}
