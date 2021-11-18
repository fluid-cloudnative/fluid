package utils

import (
	"math/rand"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
)

func TestInjectPreferredSchedulingTerms(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

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
		lenNodePrefer = rand.Intn(3) + 1
		lenNodeRequire = rand.Intn(3) + 1
		lenPodPrefer = rand.Intn(3) + 1
		lenPodAntiPrefer = rand.Intn(3) + 1
		lenPodRequire = rand.Intn(3) + 1
		lenPodAntiRequire = rand.Intn(3) + 1
		lenPreferredSchedulingTerms = rand.Intn(3) + 1

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
		expectLen            int
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
			expectLen: 0,
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
			expectLen: 1,
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
			expectLen: 2,
		},
	}

	for k, item := range testCases {
		InjectNodeSelectorTerms(item.nodeSelectorTermList, item.pod)
		if k == "test empty nodeSelectorTermList " {
			if len(item.pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) !=
				item.expectLen {
				t.Errorf("%s InjectNodeSelectorTerms failure, want:%v, got:%v", k, item.expectLen, item.pod.Spec.Affinity.NodeAffinity)
			}
		} else {
			if len(item.pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions) !=
				item.expectLen {
				t.Errorf("%s InjectNodeSelectorTerms failure, want:%v, got:%v", k, item.expectLen, item.pod.Spec.Affinity.NodeAffinity)
			}
		}
	}
}
