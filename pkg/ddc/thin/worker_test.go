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
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	ctrlhelper "github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestThinEngine_ShouldSetupWorkers(t *testing.T) {
	type fields struct {
		name      string
		namespace string
		runtime   *datav1alpha1.ThinRuntime
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
				name:      "test0",
				namespace: "thin",
				runtime: &datav1alpha1.ThinRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test0",
						Namespace: "thin",
					},
					Spec: datav1alpha1.ThinRuntimeSpec{
						Worker: datav1alpha1.ThinCompTemplateSpec{
							Enabled: true,
						},
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
				name:      "test1",
				namespace: "thin",
				runtime: &datav1alpha1.ThinRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: "thin",
					},
					Spec: datav1alpha1.ThinRuntimeSpec{
						Worker: datav1alpha1.ThinCompTemplateSpec{
							Enabled: true,
						},
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
				name:      "test2",
				namespace: "thin",
				runtime: &datav1alpha1.ThinRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "thin",
					},
					Spec: datav1alpha1.ThinRuntimeSpec{
						Worker: datav1alpha1.ThinCompTemplateSpec{
							Enabled: true,
						},
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
				name:      "test3",
				namespace: "thin",
				runtime: &datav1alpha1.ThinRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test3",
						Namespace: "thin",
					},
					Spec: datav1alpha1.ThinRuntimeSpec{
						Worker: datav1alpha1.ThinCompTemplateSpec{
							Enabled: true,
						},
					},
					Status: datav1alpha1.RuntimeStatus{
						WorkerPhase: datav1alpha1.RuntimePhaseReady,
					},
				},
			},
			wantShould: false,
			wantErr:    false,
		},
		{
			name: "test4",
			fields: fields{
				name:      "test3",
				namespace: "thin",
				runtime: &datav1alpha1.ThinRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test3",
						Namespace: "thin",
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
			e := &ThinEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				runtime:   tt.fields.runtime,
				Client:    mockClient,
			}

			gotShould, err := e.ShouldSetupWorkers()
			if (err != nil) != tt.wantErr {
				t.Errorf("ThinEngine.ShouldSetupWorkers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotShould != tt.wantShould {
				t.Errorf("ThinEngine.ShouldSetupWorkers() = %v, want %v", gotShould, tt.wantShould)
			}
		})
	}
}

func TestThinEngine_SetupWorkers(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("thin", "fluid", common.ThinRuntime)

	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfo.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
	})

	nodeSelector := map[string]string{
		"node-select": "true",
	}
	runtimeInfo.SetFuseNodeSelector(nodeSelector)

	type fields struct {
		replicas    int32
		nodeInputs  []*v1.Node
		worker      appsv1.StatefulSet
		runtime     *datav1alpha1.ThinRuntime
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
							Name: "test-node",
						},
					},
				},
				worker: appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-worker",
						Namespace: "fluid",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
				},
				runtime: &datav1alpha1.ThinRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.ThinRuntimeSpec{
						Replicas: 1,
					},
				},
				runtimeInfo: runtimeInfo,
				name:        "test",
				namespace:   "fluid",
			},
			wantedNodeLabels: map[string]map[string]string{
				"test-node": {
					"fluid.io/dataset-num":           "1",
					"fluid.io/s-fluid-thin":          "true",
					"fluid.io/s-h-thin-t-fluid-thin": "0B",
					"fluid.io/s-thin-fluid-thin":     "true",
					"fluid_exclusive":                "fluid_thin",
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

			e := &ThinEngine{
				runtime:     tt.fields.runtime,
				runtimeInfo: tt.fields.runtimeInfo,
				Client:      mockClient,
				name:        tt.fields.name,
				namespace:   tt.fields.namespace,
				Log:         ctrl.Log.WithName(tt.fields.name),
			}

			e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)
			err = e.SetupWorkers()
			if err != nil {
				t.Errorf("ThinEngine.SetupWorkers() error = %v", err)
			}
			if tt.fields.replicas != *tt.fields.worker.Spec.Replicas {
				t.Errorf("Failed to scale %v for %v", tt.name, tt.fields)
			}
		})
	}
}

func TestThinEngine_CheckWorkersReady(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.ThinRuntime
		fuse      *appsv1.DaemonSet
		worker    *appsv1.StatefulSet
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
				name:      "test0",
				namespace: "thin",
				runtime: &datav1alpha1.ThinRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test0",
						Namespace: "thin",
					},
					Spec: datav1alpha1.ThinRuntimeSpec{
						Replicas: 1,
						Worker: datav1alpha1.ThinCompTemplateSpec{
							Enabled: true,
						},
						Fuse: datav1alpha1.ThinFuseSpec{},
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test0-worker",
						Namespace: "thin",
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      1,
						ReadyReplicas: 1,
					},
				},
				fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test0-fuse",
						Namespace: "thin",
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
				name:      "test1",
				namespace: "thin",
				runtime: &datav1alpha1.ThinRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: "thin",
					},
					Spec: datav1alpha1.ThinRuntimeSpec{
						Replicas: 1,
						Worker: datav1alpha1.ThinCompTemplateSpec{
							Enabled: true,
						},
						Fuse: datav1alpha1.ThinFuseSpec{},
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1-worker",
						Namespace: "thin",
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      1,
						ReadyReplicas: 0,
					},
				},
				fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1-fuse",
						Namespace: "thin",
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
			s.AddKnownTypes(appsv1.SchemeGroupVersion, tt.fields.fuse)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, tt.fields.worker)
			_ = v1.AddToScheme(s)

			runtimeObjs = append(runtimeObjs, tt.fields.runtime, data, tt.fields.fuse, tt.fields.worker)
			mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)
			e := &ThinEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
				Log:       ctrl.Log.WithName(tt.fields.name),
			}
			runtimeInfo, err := base.BuildRuntimeInfo(tt.fields.name, tt.fields.namespace, common.ThinRuntime)
			if err != nil {
				t.Errorf("ThinEngine.CheckWorkersReady() error = %v", err)
			}

			e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)

			gotReady, err := e.CheckWorkersReady()
			if (err != nil) != tt.wantErr {
				t.Errorf("ThinEngine.CheckWorkersReady() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotReady != tt.wantReady {
				t.Errorf("ThinEngine.CheckWorkersReady() = %v, want %v", gotReady, tt.wantReady)
			}
		})
	}
}

func TestThinEngine_GetWorkerSelectors(t *testing.T) {
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
			want: "app=thin,release=spark,role=thin-worker",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ThinEngine{
				name: tt.fields.name,
			}
			if got := e.getWorkerSelectors(); got != tt.want {
				t.Errorf("ThinEngine.getWorkerSelectors() = %v, want %v", got, tt.want)
			}
		})
	}
}
