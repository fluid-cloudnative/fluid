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

package fusesidecar

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/api"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("FuseSidecar Plugin", func() {
	var (
		plugin api.MutatingHandler
		err    error
	)

	BeforeEach(func() {
		s := runtime.NewScheme()
		Expect(corev1.AddToScheme(s)).To(Succeed())

		c := fake.NewClientBuilder().
			WithScheme(s).
			Build()

		plugin, err = NewPlugin(c, "")
	})

	It("creates plugin successfully", func() {
		Expect(err).NotTo(HaveOccurred())
		Expect(plugin.GetName()).To(Equal(Name))
	})

	Context("when mutating a pod", func() {
		var pod *corev1.Pod

		BeforeEach(func() {
			pod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
			}
		})

		It("does not stop mutation when runtimeInfo is present", func() {
			runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio")
			Expect(err).NotTo(HaveOccurred())

			shouldStop, err := plugin.Mutate(
				pod,
				map[string]base.RuntimeInfoInterface{"test": runtimeInfo},
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(shouldStop).To(BeFalse())
		})

		It("does not error when runtimeInfos is empty", func() {
			_, err := plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{})
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not error when runtimeInfo is nil", func() {
			_, err := plugin.Mutate(
				pod,
				map[string]base.RuntimeInfoInterface{"test": nil},
			)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func TestFuseSidecar(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FuseSidecar Plugin Suite")
}
