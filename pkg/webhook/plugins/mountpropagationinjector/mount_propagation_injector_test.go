/*

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

package mountpropagationinjector

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("MountPropagationInjector Plugin", func() {
	var (
		cl     client.Client
		pod    *corev1.Pod
		plugin api.MutatingHandler
		err    error
	)

	BeforeEach(func() {
		cl = nil
		plugin, err = NewPlugin(cl, "")
	})

	It("should create plugin without error", func() {
		Expect(err).NotTo(HaveOccurred())
		Expect(plugin.GetName()).To(Equal(Name))
	})

	Context("Mutate method", func() {
		var runtimeInfo base.RuntimeInfoInterface

		BeforeEach(func() {
			runtimeInfo, err = base.BuildRuntimeInfo("test", "fluid", "alluxio")
			Expect(err).NotTo(HaveOccurred())
			pod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
			}
		})

		It("should mutate pod with valid runtimeInfo and not stop", func() {
			shouldStop, err := plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"test": runtimeInfo})
			Expect(err).NotTo(HaveOccurred())
			Expect(shouldStop).To(BeFalse())
		})

		It("should mutate pod with empty runtimeInfos", func() {
			shouldStop, err := plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{})
			Expect(err).NotTo(HaveOccurred())
			Expect(shouldStop).To(BeFalse())
		})

		It("should return error when runtimeInfo is nil", func() {
			shouldStop, err := plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"test": nil})
			Expect(err).To(HaveOccurred())
			Expect(shouldStop).To(BeTrue())
		})

		It("should mutate pod with multiple runtimeInfos", func() {
			runtimeInfo2, err := base.BuildRuntimeInfo("test2", "fluid", "juicefs")
			Expect(err).NotTo(HaveOccurred())
			runtimeInfo3, err := base.BuildRuntimeInfo("test3", "fluid", "jindo")
			Expect(err).NotTo(HaveOccurred())

			pod.Spec.Volumes = []corev1.Volume{
				{
					Name: "vol-test",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "test",
						},
					},
				},
				{
					Name: "vol-test3",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "test3",
						},
					},
				},
			}
			pod.Spec.Containers = []corev1.Container{
				{
					Name: "app",
					VolumeMounts: []corev1.VolumeMount{
						{Name: "vol-test", MountPath: "/mnt/test"},
						{Name: "vol-test3", MountPath: "/mnt/test3"},
					},
				},
			}

			runtimeMap := map[string]base.RuntimeInfoInterface{
				"test3": runtimeInfo3,
				"test":  runtimeInfo,
				"test2": runtimeInfo2,
			}

			for i := 0; i < 20; i++ {
				testPod := pod.DeepCopy()
				shouldStop, err := plugin.Mutate(testPod, runtimeMap)
				Expect(err).NotTo(HaveOccurred())
				Expect(shouldStop).To(BeFalse())

				hostToContainer := corev1.MountPropagationHostToContainer
				Expect(testPod.Spec.Containers[0].VolumeMounts[0].MountPropagation).To(Equal(&hostToContainer))
				Expect(testPod.Spec.Containers[0].VolumeMounts[1].MountPropagation).To(Equal(&hostToContainer))
			}
		})
	})
})
