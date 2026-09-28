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
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("CacheEngine desiredComponentVersion Tests", Label("pkg.ddc.cache.engine.transform_common_test.go"), func() {
	const templateImage = "btxu/mooncake:v3"

	Describe("desiredComponentVersion", func() {
		It("should leave a version naming neither half untouched", func() {
			desired := desiredComponentVersion(datav1alpha1.VersionSpec{}, templateImage)

			Expect(desired.Image).To(BeEmpty())
			Expect(desired.ImageTag).To(BeEmpty())
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

		It("should not carry imagePullPolicy into the completion", func() {
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
