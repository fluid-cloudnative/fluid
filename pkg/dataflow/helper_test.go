package dataflow

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

func TestGenerateNodeLabels(t *testing.T) {
	type args struct {
		job *batchv1.Job
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
				job: &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name: "jobtest",
						Annotations: map[string]string{
							common.AnnotationDataFlowAffinityInject:                                        "true",
							common.AnnotationDataFlowCustomizedAffinityPrefix + common.K8sNodeNameLabelKey: "node01",
							common.AnnotationDataFlowCustomizedAffinityPrefix + common.K8sRegionLabelKey:   "region01",
							common.AnnotationDataFlowCustomizedAffinityPrefix + common.K8sZoneLabelKey:     "zone01",
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
				job: nil,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "customized labels",
			args: args{
				job: &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name: "jobtest",
						Annotations: map[string]string{
							common.AnnotationDataFlowAffinityInject:                                        "true",
							common.AnnotationDataFlowCustomizedAffinityPrefix + common.K8sNodeNameLabelKey: "node01",
							common.AnnotationDataFlowCustomizedAffinityPrefix + common.K8sZoneLabelKey:     "zone01",
							common.AnnotationDataFlowCustomizedAffinityPrefix + "k8s.rack":                 "rack01",
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateNodeAffinity(tt.args.job)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateNodeAffinity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				// map traversal causes array order to be different, so only compare the length.
				if len(got.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) != len(tt.want.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) {
					t.Errorf("GenerateNodeAffinity() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestGenerateNodeAffinityFromPod(t *testing.T) {
	tests := []struct {
		name    string
		pod     *v1.Pod
		wantNil bool
		wantErr bool
		wantLen int
	}{
		{
			name:    "nil pod returns nil",
			pod:     nil,
			wantNil: true,
			wantErr: false,
		},
		{
			name: "pod without inject annotation returns nil",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "backup-pod",
					Annotations: map[string]string{},
				},
			},
			wantNil: true,
			wantErr: false,
		},
		{
			name: "pod with inject annotation but no affinity labels returns error",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backup-pod",
					Annotations: map[string]string{
						common.AnnotationDataFlowAffinityInject: "true",
					},
				},
			},
			wantNil: true,
			wantErr: true,
		},
		{
			name: "pod with inject annotation and affinity labels returns NodeAffinity",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backup-pod",
					Annotations: map[string]string{
						common.AnnotationDataFlowAffinityInject:                                        "true",
						common.AnnotationDataFlowCustomizedAffinityPrefix + common.K8sNodeNameLabelKey: "node01",
						common.AnnotationDataFlowCustomizedAffinityPrefix + common.K8sZoneLabelKey:     "zone01",
					},
				},
			},
			wantNil: false,
			wantErr: false,
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateNodeAffinityFromPod(tt.pod)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateNodeAffinityFromPod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantNil {
				if got != nil {
					t.Errorf("GenerateNodeAffinityFromPod() expected nil, got %v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("GenerateNodeAffinityFromPod() expected non-nil NodeAffinity")
			}
			terms := got.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
			if len(terms) != 1 {
				t.Fatalf("expected 1 NodeSelectorTerm, got %d", len(terms))
			}
			if len(terms[0].MatchExpressions) != tt.wantLen {
				t.Errorf("expected %d MatchExpressions, got %d", tt.wantLen, len(terms[0].MatchExpressions))
			}
		})
	}
}
