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

package ctrl

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
)

var _ = Describe("CheckWorkerAffinity", func() {
	var (
		s           *runtime.Scheme
		name        string
		namespace   string
		runtimeInfo base.RuntimeInfoInterface
		h           *Helper
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		name = "check-worker-affinity"
		namespace = "big-data"
		runtimeObjs := []runtime.Object{}
		mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		var err error
		runtimeInfo, err = base.BuildRuntimeInfo(name, namespace, common.JindoRuntime)
		Expect(err).NotTo(HaveOccurred())

		h = BuildHelper(runtimeInfo, mockClient, fake.NullLogger())
	})

	Context("when worker has no affinity", func() {
		It("should return false", func() {
			worker := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-affinity-worker",
					Namespace: namespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](1),
				},
			}

			result := h.checkWorkerAffinity(worker)
			Expect(result).To(BeFalse())
		})
	})

	Context("when worker has no node affinity", func() {
		It("should return false", func() {
			worker := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-node-affinity-worker",
					Namespace: namespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](1),
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Affinity: &v1.Affinity{},
						},
					},
				},
			}

			result := h.checkWorkerAffinity(worker)
			Expect(result).To(BeFalse())
		})
	})

	Context("when worker has other affinity without matching fluid preference", func() {
		It("should return false", func() {
			worker := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name + "-worker",
					Namespace: namespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](1),
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Affinity: &v1.Affinity{
								PodAntiAffinity: &v1.PodAntiAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
										{
											LabelSelector: &metav1.LabelSelector{
												MatchExpressions: []metav1.LabelSelectorRequirement{
													{
														Key:      "fluid.io/dataset",
														Operator: metav1.LabelSelectorOpExists,
													},
												},
											},
											TopologyKey: "kubernetes.io/hostname",
										},
									},
								},
								NodeAffinity: &v1.NodeAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
										NodeSelectorTerms: []v1.NodeSelectorTerm{
											{
												MatchExpressions: []v1.NodeSelectorRequirement{
													{
														Key:      "nodeA",
														Operator: v1.NodeSelectorOpIn,
														Values:   []string{"true"},
													},
												},
											},
										},
									},
									PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
										{
											Weight: 100,
											Preference: v1.NodeSelectorTerm{
												MatchExpressions: []v1.NodeSelectorRequirement{
													{
														Key:      "fluid.io/f-big-data-test3",
														Operator: v1.NodeSelectorOpIn,
														Values:   []string{"true"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			result := h.checkWorkerAffinity(worker)
			Expect(result).To(BeFalse())
		})
	})

	Context("when worker has matching fluid preference affinity", func() {
		It("should return true", func() {
			worker := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name + "-worker",
					Namespace: namespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](1),
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Affinity: &v1.Affinity{
								PodAntiAffinity: &v1.PodAntiAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
										{
											LabelSelector: &metav1.LabelSelector{
												MatchExpressions: []metav1.LabelSelectorRequirement{
													{
														Key:      "fluid.io/dataset",
														Operator: metav1.LabelSelectorOpExists,
													},
												},
											},
											TopologyKey: "kubernetes.io/hostname",
										},
									},
								},
								NodeAffinity: &v1.NodeAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
										NodeSelectorTerms: []v1.NodeSelectorTerm{
											{
												MatchExpressions: []v1.NodeSelectorRequirement{
													{
														Key:      "nodeA",
														Operator: v1.NodeSelectorOpIn,
														Values:   []string{"true"},
													},
												},
											},
										},
									},
									PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
										{
											Weight: 100,
											Preference: v1.NodeSelectorTerm{
												MatchExpressions: []v1.NodeSelectorRequirement{
													{
														Key:      "fluid.io/f-big-data-" + name,
														Operator: v1.NodeSelectorOpIn,
														Values:   []string{"true"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			result := h.checkWorkerAffinity(worker)
			Expect(result).To(BeTrue())
		})
	})

	Context("edge cases", func() {
		It("should handle worker with nil replicas", func() {
			worker := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nil-replicas-worker",
					Namespace: namespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: nil,
				},
			}

			result := h.checkWorkerAffinity(worker)
			Expect(result).To(BeFalse())
		})

		It("should handle worker with empty pod template", func() {
			worker := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "empty-template-worker",
					Namespace: namespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](1),
					Template: v1.PodTemplateSpec{},
				},
			}

			result := h.checkWorkerAffinity(worker)
			Expect(result).To(BeFalse())
		})

		It("should handle worker with only pod anti-affinity", func() {
			worker := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "only-pod-anti-affinity-worker",
					Namespace: namespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](1),
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Affinity: &v1.Affinity{
								PodAntiAffinity: &v1.PodAntiAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
										{
											LabelSelector: &metav1.LabelSelector{
												MatchExpressions: []metav1.LabelSelectorRequirement{
													{
														Key:      "fluid.io/dataset",
														Operator: metav1.LabelSelectorOpExists,
													},
												},
											},
											TopologyKey: "kubernetes.io/hostname",
										},
									},
								},
							},
						},
					},
				},
			}

			result := h.checkWorkerAffinity(worker)
			Expect(result).To(BeFalse())
		})

		It("should handle worker with node affinity but no preferred terms", func() {
			worker := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-preferred-terms-worker",
					Namespace: namespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](1),
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Affinity: &v1.Affinity{
								NodeAffinity: &v1.NodeAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
										NodeSelectorTerms: []v1.NodeSelectorTerm{
											{
												MatchExpressions: []v1.NodeSelectorRequirement{
													{
														Key:      "nodeA",
														Operator: v1.NodeSelectorOpIn,
														Values:   []string{"true"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			result := h.checkWorkerAffinity(worker)
			Expect(result).To(BeFalse())
		})

		It("should handle worker with empty preferred terms", func() {
			worker := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "empty-preferred-terms-worker",
					Namespace: namespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](1),
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Affinity: &v1.Affinity{
								NodeAffinity: &v1.NodeAffinity{
									PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{},
								},
							},
						},
					},
				},
			}

			result := h.checkWorkerAffinity(worker)
			Expect(result).To(BeFalse())
		})
	})

	Context("with multiple preferred scheduling terms", func() {
		It("should return true when matching term is found among multiple terms", func() {
			worker := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name + "-worker",
					Namespace: namespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](1),
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Affinity: &v1.Affinity{
								NodeAffinity: &v1.NodeAffinity{
									PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
										{
											Weight: 50,
											Preference: v1.NodeSelectorTerm{
												MatchExpressions: []v1.NodeSelectorRequirement{
													{
														Key:      "other-label",
														Operator: v1.NodeSelectorOpIn,
														Values:   []string{"value"},
													},
												},
											},
										},
										{
											Weight: 100,
											Preference: v1.NodeSelectorTerm{
												MatchExpressions: []v1.NodeSelectorRequirement{
													{
														Key:      "fluid.io/f-big-data-" + name,
														Operator: v1.NodeSelectorOpIn,
														Values:   []string{"true"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			result := h.checkWorkerAffinity(worker)
			Expect(result).To(BeTrue())
		})

		It("should return false when no matching term is found among multiple terms", func() {
			worker := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name + "-worker",
					Namespace: namespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](1),
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Affinity: &v1.Affinity{
								NodeAffinity: &v1.NodeAffinity{
									PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
										{
											Weight: 50,
											Preference: v1.NodeSelectorTerm{
												MatchExpressions: []v1.NodeSelectorRequirement{
													{
														Key:      "other-label-1",
														Operator: v1.NodeSelectorOpIn,
														Values:   []string{"value"},
													},
												},
											},
										},
										{
											Weight: 100,
											Preference: v1.NodeSelectorTerm{
												MatchExpressions: []v1.NodeSelectorRequirement{
													{
														Key:      "other-label-2",
														Operator: v1.NodeSelectorOpIn,
														Values:   []string{"true"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			result := h.checkWorkerAffinity(worker)
			Expect(result).To(BeFalse())
		})
	})

	Context("with different namespace combinations", func() {
		It("should work correctly with different namespace", func() {
			differentNamespace := "different-namespace"
			worker := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name + "-worker",
					Namespace: differentNamespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](1),
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Affinity: &v1.Affinity{
								NodeAffinity: &v1.NodeAffinity{
									PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
										{
											Weight: 100,
											Preference: v1.NodeSelectorTerm{
												MatchExpressions: []v1.NodeSelectorRequirement{
													{
														Key:      "fluid.io/f-big-data-" + name,
														Operator: v1.NodeSelectorOpIn,
														Values:   []string{"true"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			result := h.checkWorkerAffinity(worker)
			Expect(result).To(BeTrue())
		})
	})

	Context("with various match expression operators", func() {
		It("should return true with NotIn operator but matching key pattern", func() {
			worker := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name + "-worker",
					Namespace: namespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](1),
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Affinity: &v1.Affinity{
								NodeAffinity: &v1.NodeAffinity{
									PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
										{
											Weight: 100,
											Preference: v1.NodeSelectorTerm{
												MatchExpressions: []v1.NodeSelectorRequirement{
													{
														Key:      "fluid.io/f-big-data-" + name,
														Operator: v1.NodeSelectorOpNotIn,
														Values:   []string{"false"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			result := h.checkWorkerAffinity(worker)
			Expect(result).To(BeTrue())
		})

		It("should return true with Exists operator and matching key pattern", func() {
			worker := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name + "-worker",
					Namespace: namespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](1),
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Affinity: &v1.Affinity{
								NodeAffinity: &v1.NodeAffinity{
									PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
										{
											Weight: 100,
											Preference: v1.NodeSelectorTerm{
												MatchExpressions: []v1.NodeSelectorRequirement{
													{
														Key:      "fluid.io/f-big-data-" + name,
														Operator: v1.NodeSelectorOpExists,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			result := h.checkWorkerAffinity(worker)
			Expect(result).To(BeTrue())
		})
	})
})
