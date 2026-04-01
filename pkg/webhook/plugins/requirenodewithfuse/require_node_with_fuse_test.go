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

			// runtimeInfo is nil
			_, err = getRequiredSchedulingTerm(nil)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Mutate", func() {
		var (
			cl  client.Client
			pod *corev1.Pod
		)

		BeforeEach(func() {
			cl = nil
			pod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
			}
		})

		It("should create plugin and mutate pod correctly", func() {
			plugin, err := NewPlugin(cl, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(plugin.GetName()).To(Equal(Name))

			runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio")
			Expect(err).NotTo(HaveOccurred())

			shouldStop, err := plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"pvcName": runtimeInfo})
			Expect(err).NotTo(HaveOccurred())
			Expect(shouldStop).To(BeFalse())

			_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{})
			Expect(err).NotTo(HaveOccurred())

			_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"pvcName": nil})
			Expect(err).To(HaveOccurred())
		})

		It("should inject node selector terms when runtimeInfo has fuse node selectors", func() {
			plugin, err := NewPlugin(cl, "")
			Expect(err).NotTo(HaveOccurred())

			runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio")
			Expect(err).NotTo(HaveOccurred())
			runtimeInfo.SetFuseNodeSelector(map[string]string{"fluid.io/fuse": "true"})

			shouldStop, err := plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"pvcName": runtimeInfo})
			Expect(err).NotTo(HaveOccurred())
			Expect(shouldStop).To(BeFalse())
			Expect(pod.Spec.Affinity).NotTo(BeNil())
			Expect(pod.Spec.Affinity.NodeAffinity).NotTo(BeNil())
			terms := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
			Expect(terms).To(HaveLen(1))
			Expect(terms[0].MatchExpressions).To(HaveLen(1))
			Expect(terms[0].MatchExpressions[0].Key).To(Equal("fluid.io/fuse"))
			Expect(terms[0].MatchExpressions[0].Operator).To(Equal(corev1.NodeSelectorOpIn))
			Expect(terms[0].MatchExpressions[0].Values).To(ConsistOf("true"))
		})

		// InjectNodeSelectorTerms appends fuse MatchExpressions only into NodeSelectorTerms[0].
		// A pod with pre-existing terms (A) OR (B) becomes (A AND fuse) OR (B) after injection —
		// this is the known upstream semantic: term B can still match a node without fuse.
		It("should append fuse match expression into the first existing node selector term when pod already has multiple required node affinity terms", func() {
			const fuseKey = "fluid.io/fuse"
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

			plugin, err := NewPlugin(cl, "")
			Expect(err).NotTo(HaveOccurred())

			runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio")
			Expect(err).NotTo(HaveOccurred())
			runtimeInfo.SetFuseNodeSelector(map[string]string{fuseKey: "true"})

			shouldStop, err := plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"pvcName": runtimeInfo})
			Expect(err).NotTo(HaveOccurred())
			Expect(shouldStop).To(BeFalse())

			terms := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
			// The two original terms must be preserved; no new term is added.
			Expect(terms).To(HaveLen(2))
			// Term[0] gains the fuse expression appended alongside the original zone expression.
			Expect(terms[0].MatchExpressions).To(HaveLen(2))
			fuseKeys := make([]string, 0, len(terms[0].MatchExpressions))
			for _, me := range terms[0].MatchExpressions {
				fuseKeys = append(fuseKeys, me.Key)
			}
			Expect(fuseKeys).To(ContainElement(fuseKey))
			Expect(fuseKeys).To(ContainElement(termAKey))
			// Term[1] is left unmodified — it does NOT receive the fuse expression.
			Expect(terms[1].MatchExpressions).To(HaveLen(1))
			Expect(terms[1].MatchExpressions[0].Key).To(Equal(termBKey))
		})
	})
})
