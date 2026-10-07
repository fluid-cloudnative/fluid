/*
Copyright 2026 The Fluid Authors.

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

package engine

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("CacheEngine Transform Common Tests", Label("pkg.ddc.cache.engine.transform_common_test.go"), func() {
	var (
		engine  *CacheEngine
		dataset *datav1alpha1.Dataset
	)

	// componentValueWith builds a component value whose PodTemplateSpec stands in for the
	// CacheRuntimeClass template, which is the first of the three layers.
	componentValueWith := func(template corev1.PodTemplateSpec) *common.CacheRuntimeComponentValue {
		if len(template.Spec.Containers) == 0 {
			template.Spec.Containers = []corev1.Container{{Name: "worker", Image: "fluid/cache:v1"}}
		}
		return &common.CacheRuntimeComponentValue{
			Name:            "test-worker",
			Namespace:       "default",
			ComponentType:   common.ComponentTypeWorker,
			PodTemplateSpec: template,
		}
	}

	BeforeEach(func() {
		engine = &CacheEngine{name: "test", namespace: "default"}
		dataset = &datav1alpha1.Dataset{ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"}}
	})

	Describe("transformComponentPodTemplate pod metadata", func() {
		It("keeps the template labels and annotations when neither spec sets any", func() {
			value := componentValueWith(corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{"from": "template"},
					Annotations: map[string]string{"owner": "template"},
				},
			})

			engine.transformComponentPodTemplate(datav1alpha1.CacheRuntimeSpec{},
				datav1alpha1.RuntimeComponentCommonSpec{}, dataset, value)

			Expect(value.PodTemplateSpec.Labels).To(HaveKeyWithValue("from", "template"))
			Expect(value.PodTemplateSpec.Annotations).To(HaveKeyWithValue("owner", "template"))
		})

		It("applies the runtime-level pod metadata", func() {
			value := componentValueWith(corev1.PodTemplateSpec{})

			engine.transformComponentPodTemplate(datav1alpha1.CacheRuntimeSpec{
				PodMetadata: datav1alpha1.PodMetadata{
					Labels:      map[string]string{"from": "runtime"},
					Annotations: map[string]string{"owner": "runtime"},
				},
			}, datav1alpha1.RuntimeComponentCommonSpec{}, dataset, value)

			Expect(value.PodTemplateSpec.Labels).To(HaveKeyWithValue("from", "runtime"))
			Expect(value.PodTemplateSpec.Annotations).To(HaveKeyWithValue("owner", "runtime"))
		})

		It("applies the component-level pod metadata", func() {
			value := componentValueWith(corev1.PodTemplateSpec{})

			engine.transformComponentPodTemplate(datav1alpha1.CacheRuntimeSpec{},
				datav1alpha1.RuntimeComponentCommonSpec{
					PodMetadata: datav1alpha1.PodMetadata{
						Labels: map[string]string{"from": "component"},
					},
				}, dataset, value)

			Expect(value.PodTemplateSpec.Labels).To(HaveKeyWithValue("from", "component"))
		})

		It("layers template, runtime level and component level, with the later layer winning a key", func() {
			value := componentValueWith(corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"only-template": "yes",
						"shared":        "template",
					},
				},
			})

			engine.transformComponentPodTemplate(datav1alpha1.CacheRuntimeSpec{
				PodMetadata: datav1alpha1.PodMetadata{
					Labels: map[string]string{
						"only-runtime": "yes",
						"shared":       "runtime",
					},
				},
			}, datav1alpha1.RuntimeComponentCommonSpec{
				PodMetadata: datav1alpha1.PodMetadata{
					Labels: map[string]string{
						"only-component": "yes",
						"shared":         "component",
					},
				},
			}, dataset, value)

			Expect(value.PodTemplateSpec.Labels).To(HaveKeyWithValue("only-template", "yes"))
			Expect(value.PodTemplateSpec.Labels).To(HaveKeyWithValue("only-runtime", "yes"))
			Expect(value.PodTemplateSpec.Labels).To(HaveKeyWithValue("only-component", "yes"))
			Expect(value.PodTemplateSpec.Labels).To(HaveKeyWithValue("shared", "component"))
		})
	})

	Describe("transformComponentPodTemplate image pull secrets", func() {
		It("keeps the template secrets when the runtime names none", func() {
			value := componentValueWith(corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ImagePullSecrets: []corev1.LocalObjectReference{{Name: "from-template"}},
				},
			})

			engine.transformComponentPodTemplate(datav1alpha1.CacheRuntimeSpec{},
				datav1alpha1.RuntimeComponentCommonSpec{}, dataset, value)

			Expect(value.PodTemplateSpec.Spec.ImagePullSecrets).To(ConsistOf(
				corev1.LocalObjectReference{Name: "from-template"}))
		})

		It("applies the runtime-level secrets", func() {
			value := componentValueWith(corev1.PodTemplateSpec{})

			engine.transformComponentPodTemplate(datav1alpha1.CacheRuntimeSpec{
				ImagePullSecrets: []corev1.LocalObjectReference{{Name: "from-runtime"}},
			}, datav1alpha1.RuntimeComponentCommonSpec{}, dataset, value)

			Expect(value.PodTemplateSpec.Spec.ImagePullSecrets).To(ConsistOf(
				corev1.LocalObjectReference{Name: "from-runtime"}))
		})

		It("merges both layers and carries a secret named twice only once", func() {
			value := componentValueWith(corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: "from-template"},
						{Name: "shared"},
					},
				},
			})

			engine.transformComponentPodTemplate(datav1alpha1.CacheRuntimeSpec{
				ImagePullSecrets: []corev1.LocalObjectReference{
					{Name: "shared"},
					{Name: "from-runtime"},
				},
			}, datav1alpha1.RuntimeComponentCommonSpec{}, dataset, value)

			Expect(value.PodTemplateSpec.Spec.ImagePullSecrets).To(ConsistOf(
				corev1.LocalObjectReference{Name: "from-template"},
				corev1.LocalObjectReference{Name: "shared"},
				corev1.LocalObjectReference{Name: "from-runtime"},
			))
		})
	})
})

var _ = Describe("CacheEngine component resources Tests", Label("pkg.ddc.cache.engine.transform_common_test.go"), func() {
	// The full set of requirements a CacheRuntimeClass template typically declares.
	templateResources := func() corev1.ResourceRequirements {
		return corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("4Gi"),
			},
		}
	}

	workerDefinition := func(resources corev1.ResourceRequirements) *datav1alpha1.RuntimeComponentDefinition {
		return &datav1alpha1.RuntimeComponentDefinition{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "worker", Resources: resources}},
				},
			},
		}
	}

	Describe("mergeResourceRequirements", func() {
		Context("when the overlay only names one key", func() {
			It("should move that key and keep the rest of the baseline", func() {
				merged := mergeResourceRequirements(templateResources(), corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("8Gi")},
				})

				Expect(merged.Limits).To(HaveLen(2))
				Expect(merged.Limits.Memory().String()).To(Equal("8Gi"))
				Expect(merged.Limits.Cpu().String()).To(Equal("2"))
				Expect(merged.Requests).To(HaveLen(2))
				Expect(merged.Requests.Cpu().String()).To(Equal("1"))
				Expect(merged.Requests.Memory().String()).To(Equal("2Gi"))
			})
		})

		Context("when the overlay names a key the baseline does not declare", func() {
			It("should add it", func() {
				merged := mergeResourceRequirements(corev1.ResourceRequirements{
					Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")},
				}, corev1.ResourceRequirements{
					Requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("2Gi")},
					Limits:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("4Gi")},
				})

				Expect(merged.Requests.Cpu().String()).To(Equal("1"))
				Expect(merged.Requests.Memory().String()).To(Equal("2Gi"))
				Expect(merged.Limits.Memory().String()).To(Equal("4Gi"))
			})
		})

		Context("when neither side declares anything", func() {
			It("should leave the lists nil", func() {
				merged := mergeResourceRequirements(corev1.ResourceRequirements{}, corev1.ResourceRequirements{})

				Expect(merged.Requests).To(BeNil())
				Expect(merged.Limits).To(BeNil())
				Expect(merged.Claims).To(BeNil())
			})
		})

		It("should not mutate the baseline it was given", func() {
			base := templateResources()
			mergeResourceRequirements(base, corev1.ResourceRequirements{
				Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("8Gi")},
			})

			Expect(base.Limits.Memory().String()).To(Equal("4Gi"))
		})

		Context("with resource claims", func() {
			It("should keep the baseline claims and add the ones the overlay names", func() {
				merged := mergeResourceRequirements(corev1.ResourceRequirements{
					Claims: []corev1.ResourceClaim{{Name: "gpu"}, {Name: "nic"}},
				}, corev1.ResourceRequirements{
					Claims: []corev1.ResourceClaim{{Name: "nic"}, {Name: "fpga"}},
				})

				Expect(merged.Claims).To(Equal([]corev1.ResourceClaim{
					{Name: "gpu"},
					{Name: "nic"},
					{Name: "fpga"},
				}))
			})

			It("should keep the baseline claims when the overlay names none", func() {
				merged := mergeResourceRequirements(corev1.ResourceRequirements{
					Claims: []corev1.ResourceClaim{{Name: "gpu"}},
				}, corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("8Gi")},
				})

				Expect(merged.Claims).To(Equal([]corev1.ResourceClaim{{Name: "gpu"}}))
			})
		})
	})

	Describe("desiredComponentResources", func() {
		Context("when the CacheRuntime sets no resources", func() {
			It("should resolve to the template values", func() {
				desired := desiredComponentResources(corev1.ResourceRequirements{}, workerDefinition(templateResources()))

				Expect(desired).NotTo(BeNil())
				Expect(*desired).To(Equal(templateResources()))
			})
		})

		Context("when the CacheRuntime only raises the memory limit", func() {
			// Issue #6173: the other three values used to be dropped from the workload,
			// leaving the container with no CPU request and no CPU limit at all.
			It("should keep the template's other requirements", func() {
				desired := desiredComponentResources(corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("8Gi")},
				}, workerDefinition(templateResources()))

				Expect(desired).NotTo(BeNil())
				Expect(desired.Limits.Memory().String()).To(Equal("8Gi"))
				Expect(desired.Limits.Cpu().String()).To(Equal("2"))
				Expect(desired.Requests.Cpu().String()).To(Equal("1"))
				Expect(desired.Requests.Memory().String()).To(Equal("2Gi"))
			})
		})

		Context("when neither the CacheRuntime nor the template declares resources", func() {
			It("should return nil so the workload is left untouched", func() {
				Expect(desiredComponentResources(corev1.ResourceRequirements{}, workerDefinition(corev1.ResourceRequirements{}))).To(BeNil())
				Expect(desiredComponentResources(corev1.ResourceRequirements{}, nil)).To(BeNil())
			})
		})

		Context("when only the CacheRuntime declares resources", func() {
			It("should resolve to them", func() {
				runtimeResources := corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("8Gi")},
				}
				desired := desiredComponentResources(runtimeResources, workerDefinition(corev1.ResourceRequirements{}))

				Expect(desired).NotTo(BeNil())
				Expect(*desired).To(Equal(runtimeResources))
			})
		})

		It("should not mutate the CacheRuntimeClass template", func() {
			definition := workerDefinition(templateResources())
			desiredComponentResources(corev1.ResourceRequirements{
				Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("8Gi")},
			}, definition)

			Expect(definition.Template.Spec.Containers[0].Resources.Limits.Memory().String()).To(Equal("4Gi"))
		})
	})
})
