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
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"k8s.io/utils/ptr"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestCheckWorkerAffinity(t *testing.T) {

	s := runtime.NewScheme()
	name := "check-worker-affinity"
	namespace := "big-data"
	runtimeObjs := []runtime.Object{}
	mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)
	runtimeInfo, err := base.BuildRuntimeInfo(name, namespace, common.JindoRuntime)
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
					Replicas: ptr.To[int32](1),
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
					Replicas: ptr.To[int32](1),
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
					Replicas: ptr.To[int32](1),
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
					Replicas: ptr.To[int32](1),
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
