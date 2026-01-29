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

package nodeaffinitywithcache

import (
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("NodeAffinityWithCache Plugin", func() {
	var (
		tieredLocality = `
preferred:
- name: fluid.io/node
  weight: 100
required:
- fluid.io/node
`
		alluxioRuntime *datav1alpha1.AlluxioRuntime
	)

	BeforeEach(func() {
		alluxioRuntime = &datav1alpha1.AlluxioRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "alluxio-runtime",
				Namespace: "fluid-test",
			},
		}
	})

	It("should create plugin and return correct name", func() {
		var cl client.Client
		plugin, err := NewPlugin(cl, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(plugin.GetName()).To(Equal(Name))
	})

	Describe("getPreferredSchedulingTerm", func() {
		It("should return correct PreferredSchedulingTerm with and without selector", func() {
			runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio")
			Expect(err).NotTo(HaveOccurred())

			runtimeInfo.SetFuseNodeSelector(map[string]string{"test1": "test1"})
			term := getPreferredSchedulingTerm(100, runtimeInfo.GetCommonLabelName())

			expectTerm := corev1.PreferredSchedulingTerm{
				Weight: 100,
				Preference: corev1.NodeSelectorTerm{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      runtimeInfo.GetCommonLabelName(),
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
				},
			}
			Expect(term).To(Equal(expectTerm))

			runtimeInfo.SetFuseNodeSelector(map[string]string{})
			term = getPreferredSchedulingTerm(100, runtimeInfo.GetCommonLabelName())
			Expect(term).To(Equal(expectTerm))
		})
	})

	Describe("MutateOnlyRequired", func() {
		var (
			schema   *runtime.Scheme
			client   client.Client
			schedPod *corev1.Pod
		)

		BeforeEach(func() {
			schema = runtime.NewScheme()
			Expect(datav1alpha1.AddToScheme(schema)).To(Succeed())
			Expect(corev1.AddToScheme(schema)).To(Succeed())
			client = fake.NewFakeClientWithScheme(schema, alluxioRuntime)
			schedPod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						"fluid.io/dataset.test10-ds.sched": "required",
					},
				},
			}
		})

		It("should mutate only required", func() {
			plugin, err := NewPlugin(client, tieredLocality)
			Expect(err).NotTo(HaveOccurred())
			runtimeInfo, err := base.BuildRuntimeInfo(alluxioRuntime.Name, alluxioRuntime.Namespace, "alluxio")
			runtimeInfo.SetFuseNodeSelector(map[string]string{})
			Expect(err).NotTo(HaveOccurred())

			_, err = plugin.Mutate(schedPod, map[string]base.RuntimeInfoInterface{"pvcName": runtimeInfo})
			Expect(err).NotTo(HaveOccurred())
			// reset injected scheduling terms
			schedPod.Spec = corev1.PodSpec{}

			_, err = plugin.Mutate(schedPod, map[string]base.RuntimeInfoInterface{"test10-ds": nil})
			Expect(err).NotTo(HaveOccurred())
			Expect(schedPod.Spec.Affinity).To(BeNil())

			// reset injected scheduling terms
			schedPod.Spec = corev1.PodSpec{}

			_, err = plugin.Mutate(schedPod, map[string]base.RuntimeInfoInterface{"test10-ds": runtimeInfo})
			Expect(err).NotTo(HaveOccurred())
			Expect(schedPod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms).To(HaveLen(1))
			Expect(schedPod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution).To(BeNil())
		})
	})

	Describe("MutateOnlyPrefer", func() {
		var (
			schema *runtime.Scheme
			client client.Client
			pod    *corev1.Pod
		)

		BeforeEach(func() {
			schema = runtime.NewScheme()
			Expect(datav1alpha1.AddToScheme(schema)).To(Succeed())
			Expect(corev1.AddToScheme(schema)).To(Succeed())
			client = fake.NewFakeClientWithScheme(schema, alluxioRuntime)
			pod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
			}
		})

		It("should mutate only prefer", func() {
			plugin, err := NewPlugin(client, tieredLocality)
			Expect(err).NotTo(HaveOccurred())
			Expect(plugin.GetName()).To(Equal(Name))

			runtimeInfo, err := base.BuildRuntimeInfo(alluxioRuntime.Name, alluxioRuntime.Namespace, "alluxio")
			runtimeInfo.SetFuseNodeSelector(map[string]string{})
			Expect(err).NotTo(HaveOccurred())

			shouldStop, err := plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"pvcName": runtimeInfo})
			Expect(err).NotTo(HaveOccurred())
			Expect(shouldStop).To(BeFalse())

			_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{})
			Expect(err).NotTo(HaveOccurred())

			_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"pvcName": nil})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("MutateBothRequiredAndPrefer", func() {
		var (
			schema   *runtime.Scheme
			client   client.Client
			schedPod *corev1.Pod
		)

		BeforeEach(func() {
			schema = runtime.NewScheme()
			Expect(datav1alpha1.AddToScheme(schema)).To(Succeed())
			Expect(corev1.AddToScheme(schema)).To(Succeed())
			client = fake.NewFakeClientWithScheme(schema, alluxioRuntime)
			schedPod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						"fluid.io/dataset." + alluxioRuntime.Name + ".sched": "required",
						"fluid.io/dataset.no_exist.sched":                    "required",
					},
				},
			}
		})

		It("should mutate both required and prefer", func() {
			plugin, err := NewPlugin(client, tieredLocality)
			Expect(err).NotTo(HaveOccurred())
			runtimeInfo, err := base.BuildRuntimeInfo(alluxioRuntime.Name, alluxioRuntime.Namespace, "alluxio")
			runtimeInfo.SetFuseNodeSelector(map[string]string{})
			Expect(err).NotTo(HaveOccurred())

			runtimeInfos := map[string]base.RuntimeInfoInterface{
				alluxioRuntime.Name:   runtimeInfo,
				"prefer_dataset_name": runtimeInfo,
			}
			_, err = plugin.Mutate(schedPod, runtimeInfos)
			Expect(err).NotTo(HaveOccurred())
			Expect(schedPod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms).To(HaveLen(1))
			Expect(schedPod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution).To(HaveLen(1))
			Expect(runtimeInfos).To(HaveLen(2))
		})
	})

	Describe("TieredLocality", func() {
		var (
			customizedTieredLocality string
			schema                   *runtime.Scheme
			client                   client.Client
			runtimeInfo              base.RuntimeInfoInterface
		)

		BeforeEach(func() {
			customizedTieredLocality = `
preferred:
- name: fluid.io/fuse
  weight: 100
- name: fluid.io/node
  weight: 100
- name: topology.kubernetes.io/rack
  weight: 50
- name: topology.kubernetes.io/zone
  weight: 10
required:
- fluid.io/node
`
			alluxioRuntime = &datav1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "alluxio-runtime",
					Namespace: "fluid-test",
				},
				Status: datav1alpha1.RuntimeStatus{
					CacheAffinity: &corev1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      "topology.kubernetes.io/rack",
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{"rack-a"},
										},
										{
											Key:      "topology.kubernetes.io/zone",
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{"zone-a"},
										},
									},
								},
							},
						},
					},
				},
			}
			schema = runtime.NewScheme()
			Expect(corev1.AddToScheme(schema)).To(Succeed())
			Expect(datav1alpha1.AddToScheme(schema)).To(Succeed())
			client = fake.NewFakeClientWithScheme(schema, alluxioRuntime)
			runtimeInfo, _ = base.BuildRuntimeInfo(alluxioRuntime.Name, alluxioRuntime.Namespace, "alluxio")
			runtimeInfo.SetFuseNodeSelector(map[string]string{})
		})

		It("should mutate pod with dataset sched", func() {
			plugin, err := NewPlugin(client, customizedTieredLocality)
			Expect(err).NotTo(HaveOccurred())
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						"fluid.io/dataset." + alluxioRuntime.Name + ".sched": "required",
					},
				},
			}
			_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{
				alluxioRuntime.Name: runtimeInfo,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms).To(HaveLen(1))
		})

		It("should mutate pod with tiered locality", func() {
			plugin, err := NewPlugin(client, customizedTieredLocality)
			Expect(err).NotTo(HaveOccurred())
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
			}
			_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{
				alluxioRuntime.Name: runtimeInfo,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution).To(HaveLen(4))
		})

		It("should not mutate pod if pod already has customized preferred", func() {
			plugin, err := NewPlugin(client, customizedTieredLocality)
			Expect(err).NotTo(HaveOccurred())
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
								{
									Weight: 100,
									Preference: corev1.NodeSelectorTerm{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "topology.kubernetes.io/rack",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"rack-a"},
											},
										},
									},
								},
							},
						},
					},
				},
			}
			_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{
				alluxioRuntime.Name: runtimeInfo,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution).To(HaveLen(1))
		})

		It("should not mutate pod if pluginArg is empty", func() {
			plugin, err := NewPlugin(client, "")
			Expect(err).NotTo(HaveOccurred())
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: corev1.PodSpec{},
			}
			_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{
				alluxioRuntime.Name: runtimeInfo,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(pod.Spec.Affinity).To(BeNil())
		})

	})
})
