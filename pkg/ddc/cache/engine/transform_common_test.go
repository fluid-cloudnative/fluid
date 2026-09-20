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
