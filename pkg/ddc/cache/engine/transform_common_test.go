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
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
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
			It("should resolve to empty resources so the workload is cleared", func() {
				for _, definition := range []*datav1alpha1.RuntimeComponentDefinition{
					workerDefinition(corev1.ResourceRequirements{}),
					nil,
				} {
					desired := desiredComponentResources(corev1.ResourceRequirements{}, definition)
					Expect(desired).NotTo(BeNil())
					Expect(desired.Requests).To(BeEmpty())
					Expect(desired.Limits).To(BeEmpty())
					Expect(desired.Claims).To(BeEmpty())
				}
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

	Describe("validateResourceRequirements", func() {
		It("should accept a request equal to its limit, and a request with no limit", func() {
			Expect(validateResourceRequirements(corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("2Gi"),
				},
				Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")},
			})).To(Succeed())
			Expect(validateResourceRequirements(corev1.ResourceRequirements{})).To(Succeed())
		})

		It("should name every key whose request exceeds its limit", func() {
			// The template's requests plus a CacheRuntime that lowers both limits.
			merged := mergeResourceRequirements(corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("2Gi"),
				},
			}, corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
			})

			Expect(validateResourceRequirements(merged)).To(
				MatchError("cpu request 1 exceeds limit 500m; memory request 2Gi exceeds limit 1Gi"))
		})

		It("should accept a native request below its limit, a limit alone and one value spelled two ways", func() {
			Expect(validateResourceRequirements(corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")},
				Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")},
			})).To(Succeed())
			Expect(validateResourceRequirements(corev1.ResourceRequirements{
				Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
			})).To(Succeed())
			Expect(validateResourceRequirements(corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1000m")},
				Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")},
			})).To(Succeed())
		})

		Context("for a resource that cannot be overcommitted", func() {
			const gpu corev1.ResourceName = "nvidia.com/gpu"
			const hugePages corev1.ResourceName = "hugepages-2Mi"

			It("should accept a request equal to its limit", func() {
				Expect(validateResourceRequirements(corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						gpu:       resource.MustParse("1"),
						hugePages: resource.MustParse("1024Mi"),
					},
					Limits: corev1.ResourceList{
						gpu:       resource.MustParse("1"),
						hugePages: resource.MustParse("1Gi"),
					},
				})).To(Succeed())
			})

			It("should accept a limit alone, which the API server copies into the request", func() {
				Expect(validateResourceRequirements(corev1.ResourceRequirements{
					Limits: corev1.ResourceList{gpu: resource.MustParse("1")},
				})).To(Succeed())
			})

			It("should reject a merged huge pages request that differs from its limit", func() {
				// A template pinning 1Gi plus a CacheRuntime raising only the limit.
				merged := mergeResourceRequirements(corev1.ResourceRequirements{
					Requests: corev1.ResourceList{hugePages: resource.MustParse("1Gi")},
					Limits:   corev1.ResourceList{hugePages: resource.MustParse("1Gi")},
				}, corev1.ResourceRequirements{
					Limits: corev1.ResourceList{hugePages: resource.MustParse("2Gi")},
				})

				Expect(validateResourceRequirements(merged)).To(
					MatchError("hugepages-2Mi request 1Gi must equal limit 2Gi"))
			})

			It("should reject a merged extended resource request that differs from its limit", func() {
				merged := mergeResourceRequirements(corev1.ResourceRequirements{
					Requests: corev1.ResourceList{gpu: resource.MustParse("1")},
					Limits:   corev1.ResourceList{gpu: resource.MustParse("1")},
				}, corev1.ResourceRequirements{
					Limits: corev1.ResourceList{gpu: resource.MustParse("2")},
				})

				Expect(validateResourceRequirements(merged)).To(
					MatchError("nvidia.com/gpu request 1 must equal limit 2"))
			})

			It("should report a request above its limit as not equal", func() {
				Expect(validateResourceRequirements(corev1.ResourceRequirements{
					Requests: corev1.ResourceList{gpu: resource.MustParse("2")},
					Limits:   corev1.ResourceList{gpu: resource.MustParse("1")},
				})).To(MatchError("nvidia.com/gpu request 2 must equal limit 1"))
			})

			It("should reject a request with no limit", func() {
				Expect(validateResourceRequirements(corev1.ResourceRequirements{
					Requests: corev1.ResourceList{gpu: resource.MustParse("1")},
				})).To(MatchError("nvidia.com/gpu request 1 has no limit"))
			})

			It("should name every conflict alongside the native ones", func() {
				Expect(validateResourceRequirements(corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("1"),
						gpu:                resource.MustParse("1"),
					},
					Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")},
				})).To(MatchError("cpu request 1 exceeds limit 500m; nvidia.com/gpu request 1 has no limit"))
			})

			It("should apply to any extended resource and huge pages size, not only the ones above", func() {
				Expect(validateResourceRequirements(corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"example.com/foo": resource.MustParse("1"),
						"hugepages-1Gi":   resource.MustParse("1Gi"),
					},
					Limits: corev1.ResourceList{"example.com/foo": resource.MustParse("2")},
				})).To(MatchError("example.com/foo request 1 must equal limit 2; hugepages-1Gi request 1Gi has no limit"))
			})
		})

		It("should let a kubernetes.io-prefixed native resource set a request below its limit", func() {
			Expect(validateResourceRequirements(corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"example.kubernetes.io/foo": resource.MustParse("1")},
				Limits:   corev1.ResourceList{"example.kubernetes.io/foo": resource.MustParse("2")},
			})).To(Succeed())
		})
	})

	Describe("isOvercommitAllowed", func() {
		It("should allow native resources other than huge pages", func() {
			Expect(isOvercommitAllowed(corev1.ResourceCPU)).To(BeTrue())
			Expect(isOvercommitAllowed(corev1.ResourceMemory)).To(BeTrue())
			Expect(isOvercommitAllowed(corev1.ResourceEphemeralStorage)).To(BeTrue())
			Expect(isOvercommitAllowed("kubernetes.io/batteries")).To(BeTrue())
			Expect(isOvercommitAllowed("example.kubernetes.io/foo")).To(BeTrue())
		})

		It("should not allow huge pages or extended resources", func() {
			Expect(isOvercommitAllowed("hugepages-2Mi")).To(BeFalse())
			Expect(isOvercommitAllowed("hugepages-1Gi")).To(BeFalse())
			Expect(isOvercommitAllowed("nvidia.com/gpu")).To(BeFalse())
			Expect(isOvercommitAllowed("example.com/foo")).To(BeFalse())
		})
	})

	Describe("splitDeclaredClaims", func() {
		podSpec := &corev1.PodSpec{ResourceClaims: []corev1.PodResourceClaim{{Name: "gpu"}}}

		It("should split nothing, even without a pod spec", func() {
			declared, undeclared := splitDeclaredClaims(nil, nil)
			Expect(declared).To(BeEmpty())
			Expect(undeclared).To(BeEmpty())
		})

		It("should keep the claims the pod spec declares and name the others in order", func() {
			declared, undeclared := splitDeclaredClaims(
				[]corev1.ResourceClaim{{Name: "nic"}, {Name: "gpu"}, {Name: "fpga"}}, podSpec)
			Expect(declared).To(Equal([]corev1.ResourceClaim{{Name: "gpu"}}))
			Expect(undeclared).To(Equal([]string{"fpga", "nic"}))
		})

		It("should keep a claim the container already carries, even if the pod spec does not declare it", func() {
			carried := &corev1.PodSpec{Containers: []corev1.Container{{
				Resources: corev1.ResourceRequirements{Claims: []corev1.ResourceClaim{{Name: "nic"}}},
			}}}
			declared, undeclared := splitDeclaredClaims([]corev1.ResourceClaim{{Name: "nic"}, {Name: "gpu"}}, carried)
			Expect(declared).To(Equal([]corev1.ResourceClaim{{Name: "nic"}}))
			Expect(undeclared).To(Equal([]string{"gpu"}))
		})

		It("should treat every claim as undeclared without a pod spec", func() {
			declared, undeclared := splitDeclaredClaims([]corev1.ResourceClaim{{Name: "gpu"}}, nil)
			Expect(declared).To(BeEmpty())
			Expect(undeclared).To(Equal([]string{"gpu"}))
		})
	})

	Describe("validateComponentResources", func() {
		It("should name the component whose first container conflicts", func() {
			value := &common.CacheRuntimeComponentValue{
				ComponentType: common.ComponentTypeWorker,
				PodTemplateSpec: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{Containers: []corev1.Container{{
						Name: "worker",
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")},
							Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")},
						},
					}}},
				},
			}
			Expect(validateComponentResources(value)).To(
				MatchError("invalid worker resources: cpu request 1 exceeds limit 500m"))
		})

		It("should reject a non-overcommittable conflict on the creation path too", func() {
			value := &common.CacheRuntimeComponentValue{
				ComponentType: common.ComponentTypeMaster,
				PodTemplateSpec: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{Containers: []corev1.Container{{
						Name: "master",
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{"nvidia.com/gpu": resource.MustParse("1")},
							Limits:   corev1.ResourceList{"nvidia.com/gpu": resource.MustParse("2")},
						},
					}}},
				},
			}
			Expect(validateComponentResources(value)).To(
				MatchError("invalid master resources: nvidia.com/gpu request 1 must equal limit 2"))
		})

		It("should accept a component without containers", func() {
			value := &common.CacheRuntimeComponentValue{ComponentType: common.ComponentTypeWorker}
			Expect(validateComponentResources(value)).To(Succeed())
		})
	})
})

var _ = Describe("CacheEngine desiredComponentVersion Tests", Label("pkg.ddc.cache.engine.transform_common_test.go"), func() {
	const templateImage = "btxu/mooncake:v3"

	Describe("desiredComponentVersion", func() {
		It("should resolve a version naming neither half to the template image", func() {
			desired := desiredComponentVersion(datav1alpha1.VersionSpec{}, templateImage)

			Expect(desired.Image).To(Equal("btxu/mooncake"))
			Expect(desired.ImageTag).To(Equal("v3"))
		})

		It("should leave a complete version untouched", func() {
			desired := desiredComponentVersion(datav1alpha1.VersionSpec{
				Image:    "other/mooncake",
				ImageTag: "v9",
			}, templateImage)

			Expect(desired.Image).To(Equal("other/mooncake"))
			Expect(desired.ImageTag).To(Equal("v9"))
		})

		It("should take the repository from the template when only imageTag is named", func() {
			desired := desiredComponentVersion(datav1alpha1.VersionSpec{ImageTag: "v9"}, templateImage)

			Expect(desired.Image).To(Equal("btxu/mooncake"))
			Expect(desired.ImageTag).To(Equal("v9"))
		})

		It("should take the tag from the template when only image is named", func() {
			desired := desiredComponentVersion(datav1alpha1.VersionSpec{Image: "other/mooncake"}, templateImage)

			Expect(desired.Image).To(Equal("other/mooncake"))
			Expect(desired.ImageTag).To(Equal("v3"))
		})

		It("should keep the runtime imagePullPolicy through the completion", func() {
			desired := desiredComponentVersion(datav1alpha1.VersionSpec{
				ImageTag:        "v9",
				ImagePullPolicy: "Always",
			}, templateImage)

			Expect(desired.ImagePullPolicy).To(Equal("Always"))
		})

		It("should stay incomplete when the template declares no image", func() {
			desired := desiredComponentVersion(datav1alpha1.VersionSpec{ImageTag: "v9"}, "")

			Expect(desired.Image).To(BeEmpty())
			Expect(desired.ImageTag).To(Equal("v9"))
		})

		It("should stay incomplete when the template pins a digest", func() {
			desired := desiredComponentVersion(datav1alpha1.VersionSpec{ImageTag: "v9"},
				"btxu/mooncake@sha256:067614b70d25b496e3edc3480747d558ee8a364ef47a67f669f5d96ca5098552")

			Expect(desired.Image).To(BeEmpty())
			Expect(desired.ImageTag).To(Equal("v9"))
		})

		It("should stay incomplete when the runtime image pins a digest", func() {
			digestImage := "btxu/mooncake@sha256:067614b70d25b496e3edc3480747d558ee8a364ef47a67f669f5d96ca5098552"
			desired := desiredComponentVersion(datav1alpha1.VersionSpec{Image: digestImage}, templateImage)

			Expect(desired.Image).To(Equal(digestImage))
			Expect(desired.ImageTag).To(BeEmpty())
		})
	})

	Describe("reportIncompleteVersion", func() {
		const digestImage = "btxu/mooncake@sha256:067614b70d25b496e3edc3480747d558ee8a364ef47a67f669f5d96ca5098552"

		// logsOf runs reportIncompleteVersion against an engine logging at the default
		// verbosity, which is all the shipped controller enables, and returns every line it wrote.
		logsOf := func(runtimeVersion datav1alpha1.VersionSpec, templateImage string) []string {
			var lines []string
			e := &CacheEngine{Log: funcr.New(func(prefix, args string) {
				lines = append(lines, args)
			}, funcr.Options{})}

			e.reportIncompleteVersion(nil, common.ComponentTypeWorker, runtimeVersion, templateImage)
			return lines
		}

		It("should log a partial version the template cannot complete", func() {
			lines := logsOf(datav1alpha1.VersionSpec{ImageTag: "v9"}, digestImage)

			Expect(lines).To(HaveLen(1))
			Expect(lines[0]).To(ContainSubstring(`"component"="worker"`))
			Expect(lines[0]).To(ContainSubstring(`"imageTag"="v9"`))
		})

		It("should log a digest-pinned runtime image left without a tag", func() {
			Expect(logsOf(datav1alpha1.VersionSpec{Image: digestImage}, templateImage)).To(HaveLen(1))
		})

		It("should stay quiet for a version naming neither half", func() {
			Expect(logsOf(datav1alpha1.VersionSpec{}, digestImage)).To(BeEmpty())
		})

		It("should stay quiet for a partial version the template completes", func() {
			Expect(logsOf(datav1alpha1.VersionSpec{ImageTag: "v9"}, templateImage)).To(BeEmpty())
		})

		It("should emit a Warning event on the runtime it is given", func() {
			recorder := record.NewFakeRecorder(10)
			e := &CacheEngine{Log: logr.Discard(), Recorder: recorder}

			e.reportIncompleteVersion(&datav1alpha1.CacheRuntime{}, common.ComponentTypeWorker,
				datav1alpha1.VersionSpec{ImageTag: "v9"}, digestImage)

			Expect(recorder.Events).To(Receive(And(
				HavePrefix(corev1.EventTypeWarning+" "+common.RuntimeVersionIncomplete+" "),
				ContainSubstring(`imageTag "v9"`))))
			Expect(recorder.Events).To(BeEmpty())
		})

		It("should emit no event for a partial version the template completes", func() {
			recorder := record.NewFakeRecorder(10)
			e := &CacheEngine{Log: logr.Discard(), Recorder: recorder}

			e.reportIncompleteVersion(&datav1alpha1.CacheRuntime{}, common.ComponentTypeWorker,
				datav1alpha1.VersionSpec{ImageTag: "v9"}, templateImage)

			Expect(recorder.Events).To(BeEmpty())
		})
	})

	Describe("splitImageReference", func() {
		It("should split a plain tagged reference", func() {
			repository, tag := splitImageReference("btxu/mooncake:v3")

			Expect(repository).To(Equal("btxu/mooncake"))
			Expect(tag).To(Equal("v3"))
		})

		It("should report no tag rather than inventing one", func() {
			repository, tag := splitImageReference("nginx")

			Expect(repository).To(Equal("nginx"))
			Expect(tag).To(BeEmpty())
		})

		It("should not mistake a registry port for a tag", func() {
			repository, tag := splitImageReference("registry.local:5000/mooncake")

			Expect(repository).To(Equal("registry.local:5000/mooncake"))
			Expect(tag).To(BeEmpty())
		})

		It("should split a tag off a reference that also carries a registry port", func() {
			repository, tag := splitImageReference("registry.local:5000/mooncake:v3")

			Expect(repository).To(Equal("registry.local:5000/mooncake"))
			Expect(tag).To(Equal("v3"))
		})

		It("should refuse to split a digest reference", func() {
			repository, tag := splitImageReference("btxu/mooncake@sha256:0676")

			Expect(repository).To(BeEmpty())
			Expect(tag).To(BeEmpty())
		})
	})

	Describe("componentTemplateImage", func() {
		It("should return an empty image for a nil component definition", func() {
			Expect(componentTemplateImage(nil)).To(BeEmpty())
		})

		It("should return an empty image for a template with no containers", func() {
			Expect(componentTemplateImage(&datav1alpha1.RuntimeComponentDefinition{})).To(BeEmpty())
		})

		It("should return the first container's image", func() {
			Expect(componentTemplateImage(&datav1alpha1.RuntimeComponentDefinition{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: "worker", Image: templateImage}},
					},
				},
			})).To(Equal(templateImage))
		})
	})

	Describe("the creation and the sync path agreeing", func() {
		It("should resolve an imageTag-only version to the same image on both paths", func() {
			// creation resolves against the image already on the template copy
			creationVersion := desiredComponentVersion(datav1alpha1.VersionSpec{ImageTag: "v9"}, templateImage)

			// sync resolves against the CacheRuntimeClass component definition
			syncVersion := desiredComponentVersion(datav1alpha1.VersionSpec{ImageTag: "v9"},
				componentTemplateImage(&datav1alpha1.RuntimeComponentDefinition{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "worker", Image: templateImage}},
						},
					},
				}))

			Expect(creationVersion).To(Equal(syncVersion))
			Expect(creationVersion.Image + ":" + creationVersion.ImageTag).To(Equal("btxu/mooncake:v9"))
		})
	})
})
