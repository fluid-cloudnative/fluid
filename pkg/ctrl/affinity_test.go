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

package ctrl

import (
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilpointer "k8s.io/utils/pointer"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
			runtimeInfo, err := base.BuildRuntimeInfo(tt.fields.dataset.Name, tt.fields.dataset.Namespace, "jindo", datav1alpha1.TieredStore{})
			if err != nil {
				t.Errorf("testcase %s failed due to %v", tt.name, err)
			}
			h := BuildHelper(runtimeInfo, mockClient, fake.NullLogger())

			want := tt.fields.want
			worker, err := h.BuildWorkersAffinity(tt.fields.worker)
			if err != nil {
				t.Errorf("test BuildWorkersAffinity() = %v", err)
			}

			if !reflect.DeepEqual(worker.Spec.Template.Spec.Affinity, want) {
				t.Errorf("testcase %s BuildWorkersAffinity() = %v, want %v", tt.name, worker.Spec.Template.Spec.Affinity, tt.fields.want)
			}
		})
	}
}

func TestBuildWorkersAffinityForEFCRuntime(t *testing.T) {
	tests := []struct {
		name    string
		dataset *datav1alpha1.Dataset
		worker  *appsv1.StatefulSet
		want    *v1.Affinity
	}{
		{
			name: "efc-shared",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-efc",
					Namespace: "big-data",
				},
				Spec: datav1alpha1.DatasetSpec{
					PlacementMode: datav1alpha1.ShareMode,
				},
			},
			worker: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-efc-worker",
					Namespace: "big-data",
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: utilpointer.Int32Ptr(1),
				},
			},
			want: &v1.Affinity{
				PodAntiAffinity: &v1.PodAntiAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{},
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
										Key:      "fluid.io/f-big-data-test-efc",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := runtime.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.dataset)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, tt.worker)
			_ = v1.AddToScheme(s)

			runtimeObjs := []runtime.Object{}
			runtimeObjs = append(runtimeObjs, tt.dataset)
			runtimeObjs = append(runtimeObjs, tt.worker)
			mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)
			runtimeInfo, err := base.BuildRuntimeInfo(tt.dataset.Name, tt.dataset.Namespace, common.EFCRuntime, datav1alpha1.TieredStore{})
			if err != nil {
				t.Fatalf("testcase %s failed due to %v", tt.name, err)
			}
			h := BuildHelper(runtimeInfo, mockClient, fake.NullLogger())

			want := tt.want
			worker, err := h.BuildWorkersAffinity(tt.worker)
			if err != nil {
				t.Fatalf("test BuildWorkersAffinity() = %v", err)
			}

			if !reflect.DeepEqual(worker.Spec.Template.Spec.Affinity, want) {
				t.Fatalf("testcase %s BuildWorkersAffinity() = %v, want %v", tt.name, worker.Spec.Template.Spec.Affinity, tt.want)
			}
		})
	}
}
