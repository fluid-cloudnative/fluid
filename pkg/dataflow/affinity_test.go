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

package dataflow

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"testing"
)

func TestInjectAffinityByRunAfterOp(t *testing.T) {

	type args struct {
		runAfter        *datav1alpha1.OperationRef
		opNamespace     string
		objects         []runtime.Object
		currentAffinity *v1.Affinity
	}
	tests := []struct {
		name    string
		args    args
		want    *v1.Affinity
		wantErr bool
	}{
		{
			name: "default policy",
			args: args{
				runAfter: &datav1alpha1.OperationRef{
					Kind: "DataLoad",
					Name: "test-op",
				},
				objects: []runtime.Object{
					&datav1alpha1.DataLoad{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-op",
							Namespace: "default",
						},
						Status: datav1alpha1.OperationStatus{},
					},
				},
				opNamespace:     "default",
				currentAffinity: nil,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "no preceding op, error",
			args: args{
				runAfter: &datav1alpha1.OperationRef{
					Kind: "DataLoad",
					Name: "test-op",
					AffinityStrategy: datav1alpha1.AffinityStrategy{
						Policy: datav1alpha1.PreferAffinityStrategy,
					},
				},
				objects: []runtime.Object{
					&datav1alpha1.DataLoad{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-op2",
							Namespace: "default",
						},
						Status: datav1alpha1.OperationStatus{},
					},
				},
				opNamespace:     "default",
				currentAffinity: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "require policy, use node",
			args: args{
				runAfter: &datav1alpha1.OperationRef{
					Kind: "DataLoad",
					Name: "test-op",
					AffinityStrategy: datav1alpha1.AffinityStrategy{
						Policy: datav1alpha1.RequireAffinityStrategy,
					},
				},
				objects: []runtime.Object{
					&datav1alpha1.DataLoad{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-op",
							Namespace: "default",
						},
						Status: datav1alpha1.OperationStatus{
							NodeAffinity: &v1.NodeAffinity{
								RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
									NodeSelectorTerms: []v1.NodeSelectorTerm{
										{
											MatchExpressions: []v1.NodeSelectorRequirement{
												{
													Key:      common.K8sNodeNameLabelKey,
													Operator: v1.NodeSelectorOpIn,
													Values:   []string{"node01"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				opNamespace:     "default",
				currentAffinity: nil,
			},
			want: &v1.Affinity{
				NodeAffinity: &v1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
						NodeSelectorTerms: []v1.NodeSelectorTerm{
							{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      common.K8sNodeNameLabelKey,
										Operator: v1.NodeSelectorOpIn,
										Values:   []string{"node01"},
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "require policy, customized",
			args: args{
				runAfter: &datav1alpha1.OperationRef{
					Kind: "DataLoad",
					Name: "test-op",
					AffinityStrategy: datav1alpha1.AffinityStrategy{
						Policy: datav1alpha1.RequireAffinityStrategy,
						Requires: []datav1alpha1.Require{
							{
								Name: "k8s.rack",
							},
						},
					},
				},
				objects: []runtime.Object{
					&datav1alpha1.DataLoad{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-op",
							Namespace: "default",
						},
						Status: datav1alpha1.OperationStatus{
							NodeAffinity: &v1.NodeAffinity{
								RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
									NodeSelectorTerms: []v1.NodeSelectorTerm{
										{
											MatchExpressions: []v1.NodeSelectorRequirement{
												{
													Key:      "k8s.rack",
													Operator: v1.NodeSelectorOpIn,
													Values:   []string{"rack01"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				opNamespace:     "default",
				currentAffinity: nil,
			},
			want: &v1.Affinity{
				NodeAffinity: &v1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
						NodeSelectorTerms: []v1.NodeSelectorTerm{
							{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      "k8s.rack",
										Operator: v1.NodeSelectorOpIn,
										Values:   []string{"rack01"},
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "prefer policy, use zone",
			args: args{
				runAfter: &datav1alpha1.OperationRef{
					Kind:      "DataLoad",
					Name:      "test-op",
					Namespace: "test",
					AffinityStrategy: datav1alpha1.AffinityStrategy{
						Policy: datav1alpha1.PreferAffinityStrategy,
						Prefers: []datav1alpha1.Prefer{
							{
								Weight: 10,
								Name:   common.K8sZoneLabelKey,
							},
						},
					},
				},
				objects: []runtime.Object{
					&datav1alpha1.DataLoad{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-op",
							Namespace: "test",
						},
						Status: datav1alpha1.OperationStatus{
							NodeAffinity: &v1.NodeAffinity{
								RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
									NodeSelectorTerms: []v1.NodeSelectorTerm{
										{
											MatchExpressions: []v1.NodeSelectorRequirement{
												{
													Key:      common.K8sNodeNameLabelKey,
													Operator: v1.NodeSelectorOpIn,
													Values:   []string{"node01"},
												},
												{
													Key:      common.K8sZoneLabelKey,
													Operator: v1.NodeSelectorOpIn,
													Values:   []string{"zone01"},
												},
												{
													Key:      common.K8sRegionLabelKey,
													Operator: v1.NodeSelectorOpIn,
													Values:   []string{"region01"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				opNamespace:     "default",
				currentAffinity: nil,
			},
			want: &v1.Affinity{
				NodeAffinity: &v1.NodeAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
						{
							Weight: 10,
							Preference: v1.NodeSelectorTerm{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      common.K8sZoneLabelKey,
										Operator: v1.NodeSelectorOpIn,
										Values:   []string{"zone01"},
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	testScheme := runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
	_ = datav1alpha1.AddToScheme(testScheme)
	_ = appsv1.AddToScheme(testScheme)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fake.NewFakeClientWithScheme(testScheme, tt.args.objects...)

			got, err := InjectAffinityByRunAfterOp(c, tt.args.runAfter, tt.args.opNamespace, tt.args.currentAffinity)
			if (err != nil) != tt.wantErr {
				t.Errorf("InjectAffinityByRunAfterOp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InjectAffinityByRunAfterOp() got = %v, want %v", got, tt.want)
			}
		})
	}
}
