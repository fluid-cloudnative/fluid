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
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("FuseSidecar Plugin", func() {
	var (
		fakeClient  client.Client
		plugin      api.MutatingHandler
		runtimeInfo base.RuntimeInfoInterface
	)

	BeforeEach(func() {
		fakeClient = fake.NewFakeClient()

		var err error
		plugin, err = NewPlugin(fakeClient, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(plugin).NotTo(BeNil())

		runtimeInfo, err = base.BuildRuntimeInfo("test", "fluid", "alluxio")
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("GetName", func() {
		It("should return the correct plugin name", func() {
			Expect(plugin.GetName()).To(Equal(Name))
		})
	})

	Describe("Mutate", func() {
		var pod *corev1.Pod

		BeforeEach(func() {
			pod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
			}
		})

		Context("when runtime info is provided", func() {
			It("should mutate the pod successfully", func() {
				runtimeInfos := map[string]base.RuntimeInfoInterface{
					"test": runtimeInfo,
				}

				shouldStop, err := plugin.Mutate(pod, runtimeInfos)
				Expect(err).NotTo(HaveOccurred())
				Expect(shouldStop).To(BeFalse())
			})
		})

		Context("when runtime info map is empty", func() {
			It("should not error and return shouldStop as false", func() {
				runtimeInfos := map[string]base.RuntimeInfoInterface{}

				shouldStop, err := plugin.Mutate(pod, runtimeInfos)
				Expect(err).NotTo(HaveOccurred())
				Expect(shouldStop).To(BeFalse())
			})
		})

		Context("when runtime info is nil", func() {
			It("should handle nil runtime info without error", func() {
				runtimeInfos := map[string]base.RuntimeInfoInterface{
					"test": nil,
				}

				shouldStop, err := plugin.Mutate(pod, runtimeInfos)
				Expect(err).NotTo(HaveOccurred())
				Expect(shouldStop).To(BeFalse())
			})
		})
	})
})
