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

package nodeaffinitywithcache

import (
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestTieredLocality_hasRepeatedLocality(t1 *testing.T) {
	type args struct {
		pod *corev1.Pod
	}

	tieredLocality := &TieredLocality{
		Preferred: []Preferred{
			{
				Name:   "label.a",
				Weight: 1,
			},
			{
				Name:   "label.b",
				Weight: 2,
			},
		},
		Required: []string{"label.a"},
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "empty affinity and selector",
			args: args{
				pod: &corev1.Pod{
					Spec: corev1.PodSpec{},
				},
			},
			want: false,
		},
		{
			name: "affinity and empty selector, has same label",
			args: args{
				pod: &corev1.Pod{
					Spec: corev1.PodSpec{
						Affinity: &corev1.Affinity{
							NodeAffinity: &corev1.NodeAffinity{
								RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
									NodeSelectorTerms: []corev1.NodeSelectorTerm{
										{
											MatchExpressions: []corev1.NodeSelectorRequirement{
												{
													Key:      "label.b",
													Operator: corev1.NodeSelectorOpIn,
													Values:   []string{"b.value"},
												},
											},
										},
									},
								},
								PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
									{
										Weight: 10,
										Preference: corev1.NodeSelectorTerm{
											MatchExpressions: []corev1.NodeSelectorRequirement{
												{
													Key:      "label.b",
													Operator: corev1.NodeSelectorOpIn,
													Values:   []string{"b.value"},
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
		{
			name: "node selector with same label",
			args: args{
				pod: &corev1.Pod{
					Spec: corev1.PodSpec{
						NodeSelector: map[string]string{
							"label.a": "a-value",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "node selector without same label",
			args: args{
				pod: &corev1.Pod{
					Spec: corev1.PodSpec{
						NodeSelector: map[string]string{
							"label.c": "a-value",
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			if got := tieredLocality.hasRepeatedLocality(tt.args.pod); got != tt.want {
				t1.Errorf("hasRepeatedLocality() = %v, want %v", got, tt.want)
			}
		})
	}
}
