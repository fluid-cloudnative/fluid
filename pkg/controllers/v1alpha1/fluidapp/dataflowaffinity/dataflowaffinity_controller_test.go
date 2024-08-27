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

package dataflowaffinity

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"testing"
)

func TestDataOpJobReconciler_injectPodNodeLabelsToJob(t *testing.T) {
	type args struct {
		job  *batchv1.Job
		pods *v1.Pod
		node *v1.Node
	}
	tests := []struct {
		name            string
		args            args
		wantAnnotations map[string]string
		wantErr         bool
	}{
		{
			name: "job with succeed pods",
			args: args{
				job: &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-job",
						Labels: map[string]string{
							common.LabelAnnotationManagedBy: common.Fluid,
						},
					},
					Spec: batchv1.JobSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"controller-uid": "455afc34-93b1-4e75-a6fa-8e13d2c6ca06",
							},
						},
					},
				},
				pods: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-pod",
						Labels: map[string]string{
							"controller-uid": "455afc34-93b1-4e75-a6fa-8e13d2c6ca06",
						},
						Annotations: map[string]string{
							common.AnnotationDataFlowAffinityLabelsName: "k8s.gpu,,",
						},
					},
					Spec: v1.PodSpec{
						NodeName: "node01",
						Affinity: &v1.Affinity{
							NodeAffinity: &v1.NodeAffinity{
								RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
									NodeSelectorTerms: []v1.NodeSelectorTerm{
										{
											MatchExpressions: []v1.NodeSelectorRequirement{
												{
													Key:      "k8s.gpu",
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
					Status: v1.PodStatus{
						Phase: v1.PodSucceeded,
					},
				},
				node: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node01",
						Labels: map[string]string{
							common.K8sNodeNameLabelKey: "node01",
							common.K8sRegionLabelKey:   "region01",
							common.K8sZoneLabelKey:     "zone01",
							"k8s.gpu":                  "true",
						},
					},
				},
			},
			wantAnnotations: map[string]string{
				common.AnnotationDataFlowAffinityPrefix + common.K8sNodeNameLabelKey: "node01",
				common.AnnotationDataFlowAffinityPrefix + common.K8sRegionLabelKey:   "region01",
				common.AnnotationDataFlowAffinityPrefix + common.K8sZoneLabelKey:     "zone01",
				common.AnnotationDataFlowAffinityPrefix + "k8s.gpu":                  "true",
			},
			wantErr: false,
		},
		{
			name: "job with failed pods",
			args: args{
				job: &batchv1.Job{
					Spec: batchv1.JobSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"controller-uid": "455afc34-93b1-4e75-a6fa-8e13d2c6ca06",
							},
						},
					},
				},
				pods: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-pod",
						Labels: map[string]string{
							"controller-uid": "455afc34-93b1-4e75-a6fa-8e13d2c6ca06",
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodFailed,
					},
				},
				node: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node01",
						Labels: map[string]string{
							common.K8sNodeNameLabelKey: "node01",
							common.K8sRegionLabelKey:   "region01",
							common.K8sZoneLabelKey:     "zone01",
							"k8s.gpu":                  "true",
						},
					},
				},
			},
			wantErr: true,
		},
	}
	testScheme := runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
	_ = batchv1.AddToScheme(testScheme)
	_ = datav1alpha1.AddToScheme(testScheme)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c = fake.NewFakeClientWithScheme(testScheme, tt.args.job, tt.args.pods, tt.args.node)

			f := &DataOpJobReconciler{
				Client: c,
				Log:    fake.NullLogger(),
			}
			err := f.injectPodNodeLabelsToJob(tt.args.job)
			if (err != nil) != tt.wantErr {
				t.Errorf("injectPodNodeLabelsToJob() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && !reflect.DeepEqual(tt.args.job.Annotations, tt.wantAnnotations) {
				t.Errorf("injectPodNodeLabelsToJob() got = %v, want %v", tt.args.job.Labels, tt.wantAnnotations)
			}
		})
	}
}
