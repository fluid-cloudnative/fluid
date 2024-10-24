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

package utils

import (
	v1 "k8s.io/api/core/v1"
	"reflect"
	"testing"
)

func TestInjectNodeSelectorTermsToAffinity(t *testing.T) {
	type args struct {
		expressions []v1.NodeSelectorRequirement
		affinity    *v1.Affinity
	}
	tests := []struct {
		name string
		args args
		want *v1.Affinity
	}{
		{
			name: "test1",
			args: args{
				expressions: []v1.NodeSelectorRequirement{
					{
						Key:      "test",
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{"test"},
					},
				},
				affinity: &v1.Affinity{},
			},
			want: &v1.Affinity{
				NodeAffinity: &v1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
						NodeSelectorTerms: []v1.NodeSelectorTerm{
							{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      "test",
										Operator: v1.NodeSelectorOpIn,
										Values:   []string{"test"},
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
			if got := InjectNodeSelectorRequirements(tt.args.expressions, tt.args.affinity); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InjectNodeSelectorRequirements() = %v, want %v", got, tt.want)
			}
		})
	}
}
