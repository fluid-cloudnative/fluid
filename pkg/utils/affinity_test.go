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

func TestToNodeSelectorRequirements(t *testing.T) {
	tests := []struct {
		name      string
		selectors map[string]string
		want      []v1.NodeSelectorRequirement
	}{
		{
			name:      "nil_map",
			selectors: nil,
			want:      nil,
		},
		{
			name:      "empty_map",
			selectors: map[string]string{},
			want:      nil,
		},
		{
			name: "single_entry",
			selectors: map[string]string{
				"zone": "us-east-1a",
			},
			want: []v1.NodeSelectorRequirement{
				{
					Key:      "zone",
					Operator: v1.NodeSelectorOpIn,
					Values:   []string{"us-east-1a"},
				},
			},
		},
		{
			name: "multiple_entries_sorted",
			selectors: map[string]string{
				"region":   "us-east-1",
				"arch":     "amd64",
				"zone":     "us-east-1a",
				"instance": "large",
			},
			want: []v1.NodeSelectorRequirement{
				{
					Key:      "arch",
					Operator: v1.NodeSelectorOpIn,
					Values:   []string{"amd64"},
				},
				{
					Key:      "instance",
					Operator: v1.NodeSelectorOpIn,
					Values:   []string{"large"},
				},
				{
					Key:      "region",
					Operator: v1.NodeSelectorOpIn,
					Values:   []string{"us-east-1"},
				},
				{
					Key:      "zone",
					Operator: v1.NodeSelectorOpIn,
					Values:   []string{"us-east-1a"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToNodeSelectorRequirements(tt.selectors)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToNodeSelectorRequirements() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("deterministic_order_over_repeated_calls", func(t *testing.T) {
		multi := map[string]string{
			"z": "1",
			"a": "2",
			"m": "3",
			"b": "4",
			"k": "5",
		}
		expected := []v1.NodeSelectorRequirement{
			{Key: "a", Operator: v1.NodeSelectorOpIn, Values: []string{"2"}},
			{Key: "b", Operator: v1.NodeSelectorOpIn, Values: []string{"4"}},
			{Key: "k", Operator: v1.NodeSelectorOpIn, Values: []string{"5"}},
			{Key: "m", Operator: v1.NodeSelectorOpIn, Values: []string{"3"}},
			{Key: "z", Operator: v1.NodeSelectorOpIn, Values: []string{"1"}},
		}
		for i := 0; i < 100; i++ {
			got := ToNodeSelectorRequirements(multi)
			if !reflect.DeepEqual(got, expected) {
				t.Fatalf("iteration %d: ToNodeSelectorRequirements produced non-deterministic output: got %v, expected %v", i, got, expected)
			}
		}
	})
}
