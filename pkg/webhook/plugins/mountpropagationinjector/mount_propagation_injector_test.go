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
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("MountPropagationInjector Plugin", func() {
	var (
		plugin api.MutatingHandler
		pod    *corev1.Pod
	)

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())

		c := fake.NewClientBuilder().
			WithScheme(scheme).
			Build()

		var err error
		plugin, err = NewPlugin(c, "")
		Expect(err).NotTo(HaveOccurred())

		pod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
		}
	})

	It("returns correct plugin name", func() {
		Expect(plugin.GetName()).To(Equal(Name))
	})

	It("injects mount propagation when PVC is mounted", func() {
		pod.Spec.Volumes = []corev1.Volume{{
			Name: "data",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: "test",
				},
			},
		}}

		pod.Spec.Containers = []corev1.Container{{
			Name: "app",
			VolumeMounts: []corev1.VolumeMount{{
				Name:      "data",
				MountPath: "/data",
			}},
		}}

		runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio")
		Expect(err).NotTo(HaveOccurred())

		shouldStop, err := plugin.Mutate(
			pod,
			map[string]base.RuntimeInfoInterface{"test": runtimeInfo},
		)

		Expect(err).NotTo(HaveOccurred())
		Expect(shouldStop).To(BeFalse())

		mp := corev1.MountPropagationHostToContainer
		Expect(pod.Spec.Containers[0].VolumeMounts[0].MountPropagation).
			To(Equal(&mp))
	})
})
