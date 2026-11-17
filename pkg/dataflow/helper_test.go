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
									Key:      "k8s.rack",
									Operator: v1.NodeSelectorOpIn,
									Values:   []string{"rack01"},
								},
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
				t.Errorf("GenerateNodeAffinity() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateNodeAffinitySortsCustomizedAffinityAnnotations(t *testing.T) {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				common.AnnotationDataFlowAffinityInject:                       "true",
				common.AnnotationDataFlowCustomizedAffinityPrefix + "z-label": "z-value",
				common.AnnotationDataFlowCustomizedAffinityPrefix + "a-label": "a-value",
				common.AnnotationDataFlowCustomizedAffinityPrefix + "m-label": "m-value",
			},
		},
	}
	want := &v1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
			NodeSelectorTerms: []v1.NodeSelectorTerm{
				{
					MatchExpressions: []v1.NodeSelectorRequirement{
						{Key: "a-label", Operator: v1.NodeSelectorOpIn, Values: []string{"a-value"}},
						{Key: "m-label", Operator: v1.NodeSelectorOpIn, Values: []string{"m-value"}},
						{Key: "z-label", Operator: v1.NodeSelectorOpIn, Values: []string{"z-value"}},
					},
				},
			},
		},
	}

	for i := 0; i < 10; i++ {
		got, err := GenerateNodeAffinity(job)
		if err != nil {
			t.Fatalf("GenerateNodeAffinity() error = %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("GenerateNodeAffinity() got = %v, want %v", got, want)
		}
	}
}
