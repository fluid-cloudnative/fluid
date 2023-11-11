/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package jindocache

import (
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilpointer "k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"

	ctrlhelper "github.com/fluid-cloudnative/fluid/pkg/ctrl"
)

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
	runtimeInfoHadoop.SetupFuseDeployMode(true, nodeSelector)

	type fields struct {
		replicas         int32
		nodeInputs       []*v1.Node
		worker           *appsv1.StatefulSet
		deprecatedWorker *appsv1.DaemonSet
		runtime          *datav1alpha1.JindoRuntime
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
						Name:      "spark-jindofs-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: utilpointer.Int32Ptr(1),
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
				replicas: 1,
				worker: &appsv1.StatefulSet{

					ObjectMeta: metav1.ObjectMeta{
						Name:      "hadoop-jindofs-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: utilpointer.Int32Ptr(1),
					},
				},
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hadoop",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Replicas: 1,
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
		}, {
			name: "deprecated",
			fields: fields{
				replicas: 0,
				worker:   &appsv1.StatefulSet{},
				deprecatedWorker: &appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{
					Name:      "deprecated-jindofs-worker",
					Namespace: "big-data",
				}},
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "deprecated",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Replicas: 1,
					},
				},
				runtimeInfo: runtimeInfoHadoop,
				name:        "deprecated",
				namespace:   "big-data",
				deprecated:  true,
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

			e := &JindoCacheEngine{
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
				t.Errorf("testCase %s JindoCacheEngine.SetupWorkers() error = %v", tt.name, err)
			}

			if !tt.fields.deprecated {
				if tt.fields.replicas != *tt.fields.worker.Spec.Replicas {
					t.Errorf("Failed to scale %v for %v", tt.name, tt.fields)
				}
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

func TestShouldSetupWorkers(t *testing.T) {
	type fields struct {
		name      string
		namespace string
		runtime   *datav1alpha1.JindoRuntime
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
				runtime: &datav1alpha1.JindoRuntime{
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
				runtime: &datav1alpha1.JindoRuntime{
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
				runtime: &datav1alpha1.JindoRuntime{
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
				runtime: &datav1alpha1.JindoRuntime{
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
			e := &JindoCacheEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				runtime:   tt.fields.runtime,
				Client:    mockClient,
			}

			gotShould, err := e.ShouldSetupWorkers()
			if (err != nil) != tt.wantErr {
				t.Errorf("JindoCacheEngine.ShouldSetupWorkers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotShould != tt.wantShould {
				t.Errorf("JindoCacheEngine.ShouldSetupWorkers() = %v, want %v", gotShould, tt.wantShould)
			}
		})
	}
}

func TestCheckWorkersReady(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.JindoRuntime
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
				fuse: &appsv1.DaemonSet{
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
			},
			wantReady: true,
			wantErr:   false,
		},
		{
			name: "test1",
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
				fuse: &appsv1.DaemonSet{
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
			},
			wantReady: false,
			wantErr:   false,
		}, {
			name: "deprecated",
			fields: fields{
				name:      "deprecated",
				namespace: "big-data",
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "deprecated",
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
						Name:      "deprecated-jindofs-worker-0",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 0,
					},
				},
				fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "deprecated-jindofs-worker",
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
			e := &JindoCacheEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
				Log:       ctrl.Log.WithName(tt.fields.name),
			}

			runtimeInfo, err := base.BuildRuntimeInfo(tt.fields.name, tt.fields.namespace, "jindo", datav1alpha1.TieredStore{})
			if err != nil {
				t.Errorf("JindoCacheEngine.CheckWorkersReady() error = %v", err)
			}

			e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)

			gotReady, err := e.CheckWorkersReady()
			if (err != nil) != tt.wantErr {
				t.Errorf("JindoCacheEngine.CheckWorkersReady() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotReady != tt.wantReady {
				t.Errorf("JindoCacheEngine.CheckWorkersReady() = %v, want %v", gotReady, tt.wantReady)
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
			want: "app=jindo,release=spark,role=jindo-worker",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &JindoCacheEngine{
				name: tt.fields.name,
			}
			if got := e.getWorkerSelectors(); got != tt.want {
				t.Errorf("JindoCacheEngine.getWorkerSelectors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildWorkersAffinity(t *testing.T) {
	type fields struct {
		dataset *datav1alpha1.Dataset
		worker  *appsv1.StatefulSet
		want    *v1.Affinity
	}
	tests := []struct {
		name   string
		fields fields
		want   *v1.Affinity
	}{
		{name: "exlusive",
			fields: fields{
				dataset: &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.DatasetSpec{
						PlacementMode: datav1alpha1.ExclusiveMode,
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1-jindofs-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: utilpointer.Int32Ptr(1),
					},
				},
				want: &v1.Affinity{
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
			},
		}, {name: "shared",
			fields: fields{
				dataset: &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.DatasetSpec{
						PlacementMode: datav1alpha1.ShareMode,
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2-jindofs-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: utilpointer.Int32Ptr(1),
					},
				},
				want: &v1.Affinity{
					PodAntiAffinity: &v1.PodAntiAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
							{
								// The default weight is 50
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
			},
		}, {name: "dataset-with-affinity",
			fields: fields{
				dataset: &datav1alpha1.Dataset{
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
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test3-jindofs-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: utilpointer.Int32Ptr(1),
					},
				},
				want: &v1.Affinity{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := runtime.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.dataset)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, tt.fields.worker)
			_ = v1.AddToScheme(s)
			runtimeObjs := []runtime.Object{}
			runtimeObjs = append(runtimeObjs, tt.fields.dataset)
			runtimeObjs = append(runtimeObjs, tt.fields.worker)
			mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)
			e := &JindoCacheEngine{
				name:      tt.fields.dataset.Name,
				namespace: tt.fields.dataset.Namespace,
				Client:    mockClient,
			}

			want := tt.fields.want
			worker, err := e.buildWorkersAffinity(tt.fields.worker)
			if err != nil {
				t.Errorf("JindoCacheEngine.buildWorkersAffinity() = %v", err)
			}

			if !reflect.DeepEqual(worker.Spec.Template.Spec.Affinity, want) {
				t.Errorf("Test case %s JindoCacheEngine.buildWorkersAffinity() = %v, want %v", tt.name, worker.Spec.Template.Spec.Affinity, tt.fields.want)
			}
		})
	}
}
