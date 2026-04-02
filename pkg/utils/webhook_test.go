package utils

import (
	"math/rand"
	"reflect"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
)

func TestInjectPreferredSchedulingTerms(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var (
		pod                         corev1.Pod
		lenNodePrefer               int
		lenNodeRequire              int
		lenPodPrefer                int
		lenPodAntiPrefer            int
		lenPodRequire               int
		lenPodAntiRequire           int
		lenPreferredSchedulingTerms int
	)
	for i := 0; i < 3; i++ {
		lenNodePrefer = r.Intn(3) + 1
		lenNodeRequire = r.Intn(3) + 1
		lenPodPrefer = r.Intn(3) + 1
		lenPodAntiPrefer = r.Intn(3) + 1
		lenPodRequire = r.Intn(3) + 1
		lenPodAntiRequire = r.Intn(3) + 1
		lenPreferredSchedulingTerms = r.Intn(3) + 1

		pod.Spec.Affinity = &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: make([]corev1.NodeSelectorTerm, lenNodeRequire),
				},
				PreferredDuringSchedulingIgnoredDuringExecution: make([]corev1.PreferredSchedulingTerm, lenNodePrefer),
			},
			PodAffinity: &corev1.PodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution:  make([]corev1.PodAffinityTerm, lenPodRequire),
				PreferredDuringSchedulingIgnoredDuringExecution: make([]corev1.WeightedPodAffinityTerm, lenPodPrefer),
			},
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution:  make([]corev1.PodAffinityTerm, lenPodAntiRequire),
				PreferredDuringSchedulingIgnoredDuringExecution: make([]corev1.WeightedPodAffinityTerm, lenPodAntiPrefer),
			},
		}
		var preferredSchedulingTerms = make([]corev1.PreferredSchedulingTerm, lenPreferredSchedulingTerms)

		InjectPreferredSchedulingTerms(preferredSchedulingTerms, &pod)

		if len(pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenPodPrefer ||
			len(pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != lenPodRequire ||
			len(pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != lenPodAntiRequire ||
			len(pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenPodAntiPrefer {
			t.Errorf("should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution")
		}
		if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
			t.Errorf("should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution")
		} else {
			if len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) != lenNodeRequire {
				t.Errorf("should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution")
			}
		}
		if len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenNodePrefer+lenPreferredSchedulingTerms {
			t.Errorf("the inject is not success")
		}
	}
}

func TestInjectNodeSelectorTerms(t *testing.T) {
	testCases := map[string]struct {
		nodeSelectorTermList []corev1.NodeSelectorTerm
		pod                  *corev1.Pod
		expectTerms          []corev1.NodeSelectorTerm
	}{
		"test empty nodeSelectorTermList ": {
			nodeSelectorTermList: []corev1.NodeSelectorTerm{},
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{},
							},
						},
					},
				},
			},
			expectTerms: []corev1.NodeSelectorTerm{},
		},
		"test no empty nodeSelectorTermList ": {
			nodeSelectorTermList: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"test-label-value"},
						},
					},
				},
			},
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{},
			},
			expectTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"test-label-value"},
						},
					},
				},
			},
		},
		"test add no empty nodeSelectorTermList to pod which alredy have matchExpression": {
			nodeSelectorTermList: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"test-label-value"},
						},
					},
				},
			},
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "test",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"test-label-value2"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "test",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"test-label-value2"},
						},
						{
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"test-label-value"},
						},
					},
				},
			},
		},
		"test cross product with existing and injected or terms": {
			nodeSelectorTermList: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "disk",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"ssd"},
						},
					},
				},
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "cpu",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"amd"},
						},
					},
				},
			},
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "zone",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"z1"},
											},
										},
									},
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "region",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"r1"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "zone",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"z1"},
						},
						{
							Key:      "disk",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"ssd"},
						},
					},
				},
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "zone",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"z1"},
						},
						{
							Key:      "cpu",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"amd"},
						},
					},
				},
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "region",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"r1"},
						},
						{
							Key:      "disk",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"ssd"},
						},
					},
				},
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "region",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"r1"},
						},
						{
							Key:      "cpu",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"amd"},
						},
					},
				},
			},
		},
		"test skip empty existing term when building cross product": {
			nodeSelectorTermList: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "fluid.io/fuse",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
				},
			},
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{},
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "region",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"us-east-1"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "region",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"us-east-1"},
						},
						{
							Key:      "fluid.io/fuse",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
				},
			},
		},
		"test skip empty injected term when building cross product": {
			nodeSelectorTermList: []corev1.NodeSelectorTerm{
				{},
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "fluid.io/fuse",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
				},
			},
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "region",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"us-east-1"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "region",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"us-east-1"},
						},
						{
							Key:      "fluid.io/fuse",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
				},
			},
		},
		"test merge match fields when building cross product": {
			nodeSelectorTermList: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "fluid.io/fuse",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
					MatchFields: []corev1.NodeSelectorRequirement{
						{
							Key:      "metadata.name",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"node-b"},
						},
					},
				},
			},
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "region",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"us-east-1"},
											},
										},
										MatchFields: []corev1.NodeSelectorRequirement{
											{
												Key:      "metadata.name",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"node-a"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "region",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"us-east-1"},
						},
						{
							Key:      "fluid.io/fuse",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
					MatchFields: []corev1.NodeSelectorRequirement{
						{
							Key:      "metadata.name",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"node-a"},
						},
						{
							Key:      "metadata.name",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"node-b"},
						},
					},
				},
			},
		},
	}

	for k, item := range testCases {
		InjectNodeSelectorTerms(item.nodeSelectorTermList, item.pod)
		if item.pod.Spec.Affinity == nil ||
			item.pod.Spec.Affinity.NodeAffinity == nil ||
			item.pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
			if len(item.expectTerms) != 0 {
				t.Errorf("%s InjectNodeSelectorTerms failure, want:%v, got:nil", k, item.expectTerms)
			}
			continue
		}

		gotTerms := item.pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
		if len(gotTerms) != len(item.expectTerms) {
			t.Errorf("%s InjectNodeSelectorTerms failure, want:%v, got:%v", k, item.expectTerms, gotTerms)
			continue
		}

		for i := range gotTerms {
			if len(gotTerms[i].MatchExpressions) != len(item.expectTerms[i].MatchExpressions) {
				t.Errorf("%s InjectNodeSelectorTerms failure, want:%v, got:%v", k, item.expectTerms, gotTerms)
				break
			}

			if len(gotTerms[i].MatchFields) != len(item.expectTerms[i].MatchFields) {
				t.Errorf("%s InjectNodeSelectorTerms failure, want:%v, got:%v", k, item.expectTerms, gotTerms)
				break
			}

			for j := range gotTerms[i].MatchExpressions {
				if !reflect.DeepEqual(gotTerms[i].MatchExpressions[j], item.expectTerms[i].MatchExpressions[j]) {
					t.Errorf("%s InjectNodeSelectorTerms failure, want:%v, got:%v", k, item.expectTerms, gotTerms)
					break
				}
			}

			for j := range gotTerms[i].MatchFields {
				if !reflect.DeepEqual(gotTerms[i].MatchFields[j], item.expectTerms[i].MatchFields[j]) {
					t.Errorf("%s InjectNodeSelectorTerms failure, want:%v, got:%v", k, item.expectTerms, gotTerms)
					break
				}
			}
		}
	}
}

func TestInjectMountPropagation(t *testing.T) {
	type args struct {
		runtimeNames []string
		pod          *corev1.Pod
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test",
			args: args{
				runtimeNames: []string{"test"},
				pod: &corev1.Pod{
					Spec: corev1.PodSpec{
						Volumes: []corev1.Volume{{
							Name:         "test-volume",
							VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "test"}},
						}},
						Containers: []corev1.Container{{
							Name:         "test-cn",
							VolumeMounts: []corev1.VolumeMount{{Name: "test-volume"}},
						}},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InjectMountPropagation(tt.args.runtimeNames, tt.args.pod)
			if *tt.args.pod.Spec.Containers[0].VolumeMounts[0].MountPropagation != corev1.MountPropagationHostToContainer {
				t.Errorf("InjectMountPropagation failure, got:%v", tt.args.pod)
			}
		})
	}
}
