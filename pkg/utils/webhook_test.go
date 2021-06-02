package utils

import (
	corev1 "k8s.io/api/core/v1"
	"math/rand"
	"testing"
	"time"
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
