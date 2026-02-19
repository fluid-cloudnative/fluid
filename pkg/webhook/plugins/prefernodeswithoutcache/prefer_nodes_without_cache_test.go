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

package prefernodeswithoutcache

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("PreferNodesWithoutCache Plugin", func() {
	Describe("getPreferredSchedulingTermForPodWithoutCache", func() {
		It("should return correct PreferredSchedulingTerm with selector enabled and disabled", func() {
			runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio")
			Expect(err).NotTo(HaveOccurred())

			runtimeInfo.SetFuseNodeSelector(map[string]string{"test1": "test1"})
			term := getPreferredSchedulingTermForPodWithoutCache()

			expectTerm := corev1.PreferredSchedulingTerm{
				Weight: 100,
				Preference: corev1.NodeSelectorTerm{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      common.GetDatasetNumLabelName(),
							Operator: corev1.NodeSelectorOpDoesNotExist,
						},
					},
				},
			}
			Expect(term).To(Equal(expectTerm))

			runtimeInfo.SetFuseNodeSelector(map[string]string{})
			term = getPreferredSchedulingTermForPodWithoutCache()
			Expect(term).To(Equal(expectTerm))
		})

		It("should return correct PreferredSchedulingTerm with default mode", func() {
			runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio")
			Expect(err).NotTo(HaveOccurred())

			runtimeInfo.SetFuseNodeSelector(map[string]string{})
			term := getPreferredSchedulingTermForPodWithoutCache()

			expectTerm := corev1.PreferredSchedulingTerm{
				Weight: 100,
				Preference: corev1.NodeSelectorTerm{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      common.GetDatasetNumLabelName(),
							Operator: corev1.NodeSelectorOpDoesNotExist,
						},
					},
				},
			}
			Expect(term).To(Equal(expectTerm))
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

			shouldStop, err := plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"test": runtimeInfo})
			Expect(err).NotTo(HaveOccurred())
			Expect(shouldStop).To(BeTrue())

			_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{})
			Expect(err).NotTo(HaveOccurred())

			_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"test": nil})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
