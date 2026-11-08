/*
Copyright 2021 The Fluid Authors.

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

package requirenodewithfuse

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("RequireNodeWithFuse Plugin", func() {
	Describe("getRequiredSchedulingTerm", func() {
		It("should return correct NodeSelectorTerm with selector enabled and disabled", func() {
			runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio")
			Expect(err).NotTo(HaveOccurred())

			// Global fuse with selector enable
			runtimeInfo.SetFuseNodeSelector(map[string]string{"test1": "test1"})
			terms, err := getRequiredSchedulingTerm(runtimeInfo)
			Expect(err).NotTo(HaveOccurred())
			expectTerms := corev1.NodeSelectorTerm{
				MatchExpressions: []corev1.NodeSelectorRequirement{
					{
						Key:      "test1",
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{"test1"},
					},
				},
			}
			Expect(terms).To(Equal(expectTerms))

			// Global fuse with selector disable
			runtimeInfo.SetFuseNodeSelector(map[string]string{})
			terms, err = getRequiredSchedulingTerm(runtimeInfo)
			Expect(err).NotTo(HaveOccurred())
			expectTerms = corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{}}
			Expect(terms).To(Equal(expectTerms))

			// Global fuse with multi-label selector (sorted deterministically)
			runtimeInfo.SetFuseNodeSelector(map[string]string{
				"zone":   "us-east-1a",
				"arch":   "amd64",
				"region": "us-east-1",
			})
			for i := 0; i < 50; i++ {
				terms, err = getRequiredSchedulingTerm(runtimeInfo)
				Expect(err).NotTo(HaveOccurred())
				Expect(terms.MatchExpressions).To(Equal([]corev1.NodeSelectorRequirement{
					{
						Key:      "arch",
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{"amd64"},
					},
					{
						Key:      "region",
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{"us-east-1"},
					},
					{
						Key:      "zone",
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{"us-east-1a"},
					},
				}))
			}

			// runtimeInfo is nil
			_, err = getRequiredSchedulingTerm(nil)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Mutate", func() {
		const fuseKey = "fluid.io/fuse"

		var (
			cl          client.Client
			plugin      api.MutatingHandler
			pod         *corev1.Pod
			runtimeInfo base.RuntimeInfoInterface
		)

		BeforeEach(func() {
			cl = nil
			var err error
			plugin, err = NewPlugin(cl, "")
			Expect(err).NotTo(HaveOccurred())

			runtimeInfo, err = base.BuildRuntimeInfo("test", "fluid", "alluxio")
			Expect(err).NotTo(HaveOccurred())

			pod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
			}
		})

		It("should create plugin and mutate pod correctly", func() {
			Expect(plugin.GetName()).To(Equal(Name))

			shouldStop, err := plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"pvcName": runtimeInfo})
			Expect(err).NotTo(HaveOccurred())
			Expect(shouldStop).To(BeFalse())

			_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{})
			Expect(err).NotTo(HaveOccurred())

			_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"pvcName": nil})
			Expect(err).To(HaveOccurred())
		})

		It("should inject node selector terms when runtimeInfo has fuse node selectors", func() {
			runtimeInfo.SetFuseNodeSelector(map[string]string{fuseKey: "true"})

			shouldStop, err := plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"pvcName": runtimeInfo})
			Expect(err).NotTo(HaveOccurred())
			Expect(shouldStop).To(BeFalse())
			Expect(pod.Spec.Affinity).NotTo(BeNil())
			Expect(pod.Spec.Affinity.NodeAffinity).NotTo(BeNil())
			terms := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
			Expect(terms).To(HaveLen(1))
			Expect(terms[0].MatchExpressions).To(HaveLen(1))
			Expect(terms[0].MatchExpressions[0]).To(Equal(corev1.NodeSelectorRequirement{
				Key:      fuseKey,
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{"true"},
			}))
		})

		It("should inject fuse match expression into every existing required node affinity branch", func() {
			const termAKey = "zone"
			const termBKey = "region"

			pod.Spec.Affinity = &corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{Key: termAKey, Operator: corev1.NodeSelectorOpIn, Values: []string{"us-east-1a"}},
								},
							},
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{Key: termBKey, Operator: corev1.NodeSelectorOpIn, Values: []string{"us-east-1"}},
								},
							},
						},
					},
				},
			}

			runtimeInfo.SetFuseNodeSelector(map[string]string{fuseKey: "true"})

			shouldStop, err := plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"pvcName": runtimeInfo})
			Expect(err).NotTo(HaveOccurred())
			Expect(shouldStop).To(BeFalse())

			terms := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
			Expect(terms).To(HaveLen(2))
			Expect(terms[0].MatchExpressions).To(ConsistOf(
				corev1.NodeSelectorRequirement{
					Key:      termAKey,
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"us-east-1a"},
				},
				corev1.NodeSelectorRequirement{
					Key:      fuseKey,
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"true"},
				},
			))
			Expect(terms[1].MatchExpressions).To(ConsistOf(
				corev1.NodeSelectorRequirement{
					Key:      termBKey,
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"us-east-1"},
				},
				corev1.NodeSelectorRequirement{
					Key:      fuseKey,
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"true"},
				},
			))
		})

		It("should ignore empty existing required node affinity branches when injecting fuse requirements", func() {
			pod.Spec.Affinity = &corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{},
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{Key: "region", Operator: corev1.NodeSelectorOpIn, Values: []string{"us-east-1"}},
								},
							},
						},
					},
				},
			}

			runtimeInfo.SetFuseNodeSelector(map[string]string{fuseKey: "true"})

			shouldStop, err := plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"pvcName": runtimeInfo})
			Expect(err).NotTo(HaveOccurred())
			Expect(shouldStop).To(BeFalse())

			terms := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
			Expect(terms).To(HaveLen(1))
			Expect(terms[0].MatchExpressions).To(ConsistOf(
				corev1.NodeSelectorRequirement{Key: "region", Operator: corev1.NodeSelectorOpIn, Values: []string{"us-east-1"}},
				corev1.NodeSelectorRequirement{Key: fuseKey, Operator: corev1.NodeSelectorOpIn, Values: []string{"true"}},
			))
		})

		It("should produce deterministically ordered node selector terms across multiple runtimes", func() {
			runtimeA, err := base.BuildRuntimeInfo("ds-a", "fluid", "alluxio")
			Expect(err).NotTo(HaveOccurred())
			runtimeA.SetFuseNodeSelector(map[string]string{"z-label": "1"})

			runtimeM, err := base.BuildRuntimeInfo("ds-m", "fluid", "alluxio")
			Expect(err).NotTo(HaveOccurred())
			runtimeM.SetFuseNodeSelector(map[string]string{"m-label": "2"})

			runtimeZ, err := base.BuildRuntimeInfo("ds-z", "fluid", "alluxio")
			Expect(err).NotTo(HaveOccurred())
			runtimeZ.SetFuseNodeSelector(map[string]string{"a-label": "3"})

			infos := map[string]base.RuntimeInfoInterface{
				"runtime-z": runtimeZ,
				"runtime-a": runtimeA,
				"runtime-m": runtimeM,
			}

			expectedTerms := []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{Key: "z-label", Operator: corev1.NodeSelectorOpIn, Values: []string{"1"}},
					},
				},
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{Key: "m-label", Operator: corev1.NodeSelectorOpIn, Values: []string{"2"}},
					},
				},
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{Key: "a-label", Operator: corev1.NodeSelectorOpIn, Values: []string{"3"}},
					},
				},
			}

			for i := 0; i < 50; i++ {
				testPod := pod.DeepCopy()
				shouldStop, err := plugin.Mutate(testPod, infos)
				Expect(err).NotTo(HaveOccurred())
				Expect(shouldStop).To(BeFalse())
				terms := testPod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
				Expect(terms).To(Equal(expectedTerms))
			}
		})
	})
})
