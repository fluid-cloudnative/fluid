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
	"testing"

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

func TestEFCEngine_ShouldSetupWorkers(t *testing.T) {
	type fields struct {
		name      string
		namespace string
		runtime   *datav1alpha1.EFCRuntime
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
				namespace: "efc",
				runtime: &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test0",
						Namespace: "efc",
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
				namespace: "efc",
				runtime: &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: "efc",
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
				namespace: "efc",
				runtime: &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "efc",
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
				namespace: "efc",
				runtime: &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test3",
						Namespace: "efc",
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
			e := &EFCEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				runtime:   tt.fields.runtime,
				Client:    mockClient,
			}

			gotShould, err := e.ShouldSetupWorkers()
			if (err != nil) != tt.wantErr {
				t.Errorf("EFCEngine.ShouldSetupWorkers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotShould != tt.wantShould {
				t.Errorf("EFCEngine.ShouldSetupWorkers() = %v, want %v", gotShould, tt.wantShould)
			}
		})
	}
}

func TestEFCEngine_SetupWorkers(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("efc", "fluid", common.EFCRuntime)

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
		runtime     *datav1alpha1.EFCRuntime
		runtimeInfo base.RuntimeInfoInterface
		name        string
		namespace   string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "test",
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
				runtime: &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.EFCRuntimeSpec{
						Replicas: 1,
					},
				},
				runtimeInfo: runtimeInfo,
				name:        "test",
				namespace:   "fluid",
			},
		},
		{
			name: "test1",
			fields: fields{
				replicas: 0,
				nodeInputs: []*v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test1-node",
						},
					},
				},
				worker: appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1-worker",
						Namespace: "fluid",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](0),
					},
				},
				runtime: &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.EFCRuntimeSpec{
						Worker: datav1alpha1.EFCCompTemplateSpec{
							Disabled: true,
						},
					},
				},
				runtimeInfo: runtimeInfo,
				name:        "test1",
				namespace:   "fluid",
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

			e := &EFCEngine{
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
				t.Errorf("EFCEngine.SetupWorkers() error = %v", err)
			}
			workers, err := ctrlhelper.GetWorkersAsStatefulset(e.Client,
				types.NamespacedName{Namespace: e.namespace, Name: e.getWorkerName()})
			if err != nil {
				t.Errorf("ctrll.GetWorkersAsStatefulset error = %v", err)
			}
			if tt.fields.replicas != *workers.Spec.Replicas {
				t.Errorf("Failed to scale %v for %v", tt.name, tt.fields)
			}
		})
	}
}

func TestEFCEngine_CheckWorkersReady(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.EFCRuntime
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
				namespace: "efc",
				runtime: &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test0",
						Namespace: "efc",
					},
					Spec: datav1alpha1.EFCRuntimeSpec{
						Replicas: 1,
						Fuse:     datav1alpha1.EFCFuseSpec{},
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test0-worker",
						Namespace: "efc",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      1,
						ReadyReplicas: 1,
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
				namespace: "efc",
				runtime: &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: "efc",
					},
					Spec: datav1alpha1.EFCRuntimeSpec{
						Replicas: 1,
						Fuse:     datav1alpha1.EFCFuseSpec{},
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1-worker",
						Namespace: "efc",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      1,
						ReadyReplicas: 0,
					},
				},
			},
			wantReady: false,
			wantErr:   false,
		},
		{
			name: "test2",
			fields: fields{
				name:      "test2",
				namespace: "efc",
				runtime: &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "efc",
					},
					Spec: datav1alpha1.EFCRuntimeSpec{
						Fuse: datav1alpha1.EFCFuseSpec{},
						Worker: datav1alpha1.EFCCompTemplateSpec{
							Disabled: true,
						},
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2-worker",
						Namespace: "efc",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](0),
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      0,
						ReadyReplicas: 0,
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
			_ = v1.AddToScheme(s)

			runtimeObjs = append(runtimeObjs, tt.fields.runtime, data, tt.fields.worker)
			mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)
			e := &EFCEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
				Log:       ctrl.Log.WithName(tt.fields.name),
			}
			runtimeInfo, err := base.BuildRuntimeInfo(tt.fields.name, tt.fields.namespace, common.EFCRuntime)
			if err != nil {
				t.Errorf("EFCEngine.CheckWorkersReady() error = %v", err)
			}

			e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)

			gotReady, err := e.CheckWorkersReady()
			if (err != nil) != tt.wantErr {
				t.Errorf("EFCEngine.CheckWorkersReady() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotReady != tt.wantReady {
				t.Errorf("EFCEngine.CheckWorkersReady() = %v, want %v", gotReady, tt.wantReady)
			}
		})
	}
}

func TestEFCEngine_GetWorkerSelectors(t *testing.T) {
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
			want: "app=efc,release=spark,role=efc-worker",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &EFCEngine{
				name: tt.fields.name,
			}
			if got := e.getWorkerSelectors(); got != tt.want {
				t.Errorf("EFCEngine.getWorkerSelectors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEFCEngine_syncWorkersEndpoints(t *testing.T) {
	type fields struct {
		worker    *appsv1.StatefulSet
		pods      []*v1.Pod
		configMap *v1.ConfigMap
		name      string
		namespace string
	}
	tests := []struct {
		name      string
		fields    fields
		wantErr   bool
		wantCount int
	}{
		{
			name:      "test",
			wantErr:   false,
			wantCount: 1,
			fields: fields{
				name:      "spark",
				namespace: "big-data",
				worker: &appsv1.StatefulSet{
					TypeMeta: metav1.TypeMeta{
						Kind:       "StatefulSet",
						APIVersion: "apps/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-worker",
						Namespace: "big-data",
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
				},
				pods: []*v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "spark-worker-0",
							Namespace: "big-data",
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
					},
				},
				configMap: &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-worker-endpoints",
						Namespace: "big-data",
					},
					Data: map[string]string{
						WorkerEndpointsDataName: workerEndpointsConfigMapData,
					},
				},
			},
		},
		{
			name:      "test2",
			wantErr:   true,
			wantCount: 1,
			fields: fields{
				name:      "spark",
				namespace: "big-data",
				worker: &appsv1.StatefulSet{
					TypeMeta: metav1.TypeMeta{
						Kind:       "StatefulSet",
						APIVersion: "apps/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-worker",
						Namespace: "big-data",
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
				},
				pods: []*v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "spark-worker-0",
							Namespace: "big-data",
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
					},
				},
				configMap: &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase-worker-endpoints",
						Namespace: "big-data",
					},
					Data: map[string]string{
						WorkerEndpointsDataName: workerEndpointsConfigMapData,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtimeObjs := []runtime.Object{}

			s := runtime.NewScheme()
			s.AddKnownTypes(appsv1.SchemeGroupVersion, tt.fields.worker)
			s.AddKnownTypes(v1.SchemeGroupVersion, tt.fields.configMap)
			_ = v1.AddToScheme(s)

			runtimeObjs = append(runtimeObjs, tt.fields.worker)
			runtimeObjs = append(runtimeObjs, tt.fields.configMap)
			for _, pod := range tt.fields.pods {
				runtimeObjs = append(runtimeObjs, pod)
			}
			mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)
			e := &EFCEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
				Log:       ctrl.Log.WithName(tt.fields.name),
			}
			runtimeInfo, err := base.BuildRuntimeInfo(tt.fields.name, tt.fields.namespace, common.EFCRuntime)
			if err != nil {
				t.Errorf("EFCEngine.CheckWorkersReady() error = %v", err)
			}

			e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)

			count, err := e.syncWorkersEndpoints()
			if (err != nil) != tt.wantErr {
				t.Errorf("EFCEngine.syncWorkersEndpoints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if count != tt.wantCount {
				t.Errorf("EFCEngine.syncWorkersEndpoints() count = %v, wantCount %v", count, tt.wantCount)
				return
			}
		})
	}
}
