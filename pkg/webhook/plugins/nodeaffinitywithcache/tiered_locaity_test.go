/*
Copyright 2023 The Fluid Authors.

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

package nodeaffinitywithcache

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = ginkgo.Describe("TieredLocality.hasRepeatedLocality", func() {
	var tieredLocality *TieredLocality

	ginkgo.BeforeEach(func() {
		tieredLocality = &TieredLocality{
			Preferred: []Preferred{
				{
					Name:   "label.a",
					Weight: 1,
				},
				{
					Name:   "label.b",
					Weight: 2,
				},
			},
			Required: []string{"label.a"},
		}
	})

	ginkgo.DescribeTable("hasRepeatedLocality cases",
		func(pod *corev1.Pod, want bool) {
			got := tieredLocality.hasRepeatedLocality(pod)
			gomega.Expect(got).To(gomega.Equal(want))
		},
		ginkgo.Entry("empty affinity and selector",
			&corev1.Pod{
				Spec: corev1.PodSpec{},
			},
			false,
		),
		ginkgo.Entry("affinity and empty selector, has same label",
			&corev1.Pod{
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "label.b",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"b.value"},
											},
										},
									},
								},
							},
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
								{
									Weight: 10,
									Preference: corev1.NodeSelectorTerm{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "label.b",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"b.value"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			true,
		),
		ginkgo.Entry("node selector with same label",
			&corev1.Pod{
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						"label.a": "a-value",
					},
				},
			},
			true,
		),
		ginkgo.Entry("node selector without same label",
			&corev1.Pod{
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						"label.c": "a-value",
					},
				},
			},
			false,
		),
	)
})
