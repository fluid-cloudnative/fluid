package dataflow

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestGenerateNodeLabels(t *testing.T) {
	type args struct {
		pod  *v1.Pod
		node *v1.Node
	}
	tests := []struct {
		name    string
		args    args
		want    *v1.NodeAffinity
		wantErr bool
	}{
		{
			name: "default labels",
			args: args{
				pod: &v1.Pod{
					Spec: v1.PodSpec{
						NodeName: "node01",
					},
				},
				node: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node01",
						Labels: map[string]string{
							common.K8sNodeNameLabelKey: "node01",
							common.K8sRegionLabelKey:   "region01",
							common.K8sZoneLabelKey:     "zone01",
						},
					},
				},
			},
			want: &v1.NodeAffinity{
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
									Key:      common.K8sRegionLabelKey,
									Operator: v1.NodeSelectorOpIn,
									Values:   []string{"region01"},
								},
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
			wantErr: false,
		},
		{
			name: "nil pod",
			args: args{
				pod: nil,
				node: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node01",
						Labels: map[string]string{
							common.K8sNodeNameLabelKey: "node01",
							common.K8sRegionLabelKey:   "region01",
							common.K8sZoneLabelKey:     "zone01",
						},
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "customized labels",
			args: args{
				pod: &v1.Pod{
					Spec: v1.PodSpec{
						NodeName: "node01",
						Affinity: &v1.Affinity{
							NodeAffinity: &v1.NodeAffinity{
								PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
									{
										Preference: v1.NodeSelectorTerm{
											MatchExpressions: []v1.NodeSelectorRequirement{
												{
													Key:      "k8s.gpu",
													Operator: v1.NodeSelectorOpIn,
													Values:   []string{"true"},
												},
											},
										},
										Weight: 10,
									},
								},
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
				node: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node01",
						Labels: map[string]string{
							common.K8sNodeNameLabelKey: "node01",
							common.K8sZoneLabelKey:     "zone01",
							"k8s.rack":                 "rack01",
							"k8s.gpu":                  "false",
						},
					},
				},
			},
			want: &v1.NodeAffinity{
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
									Key:      "k8s.gpu",
									Operator: v1.NodeSelectorOpIn,
									Values:   []string{"false"},
								},
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
			wantErr: false,
		},
	}
	testScheme := runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c client.Client
			if tt.args.pod == nil {
				c = fake.NewFakeClientWithScheme(testScheme, tt.args.node)
			} else {
				c = fake.NewFakeClientWithScheme(testScheme, tt.args.node, tt.args.pod)
			}

			got, err := GenerateNodeLabels(c, tt.args.pod)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateNodeLabels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateNodeLabels() got = %v, want %v", got, tt.want)
			}
		})
	}
}
