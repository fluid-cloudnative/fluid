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

var _ = Describe("CacheEngine getDataOperationImage Tests", Label("pkg.ddc.cache.engine.image_test.go"), func() {
	const digestImage = "btxu/mooncake@sha256:067614b70d25b496e3edc3480747d558ee8a364ef47a67f669f5d96ca5098552"

	// classWithWorkerImage builds a CacheRuntimeClass whose worker template runs image.
	classWithWorkerImage := func(image string) *datav1alpha1.CacheRuntimeClass {
		return &datav1alpha1.CacheRuntimeClass{
			Topology: &datav1alpha1.RuntimeTopology{
				Worker: &datav1alpha1.RuntimeComponentDefinition{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "worker", Image: image}},
						},
					},
				},
			},
		}
	}

	// runtimeWithWorkerVersion builds a CacheRuntime whose worker sets version.
	runtimeWithWorkerVersion := func(version datav1alpha1.VersionSpec) *datav1alpha1.CacheRuntime {
		runtime := &datav1alpha1.CacheRuntime{}
		runtime.Name = "cache"
		runtime.Namespace = "default"
		runtime.Spec.Worker.RuntimeVersion = version
		return runtime
	}

	DescribeTable("should resolve the image the worker runs",
		func(version datav1alpha1.VersionSpec, templateImage string, want string) {
			image, err := (&CacheEngine{}).getDataOperationImage(
				runtimeWithWorkerVersion(version), classWithWorkerImage(templateImage))

			Expect(err).NotTo(HaveOccurred())
			Expect(image).To(Equal(want))
		},
		Entry("a full version replaces the template",
			datav1alpha1.VersionSpec{Image: "other/mooncake", ImageTag: "v9"}, "btxu/mooncake:v3", "other/mooncake:v9"),
		Entry("an imageTag-only version takes the image from the template",
			datav1alpha1.VersionSpec{ImageTag: "v9"}, "btxu/mooncake:v3", "btxu/mooncake:v9"),
		Entry("an image-only version takes the tag from the template",
			datav1alpha1.VersionSpec{Image: "other/mooncake"}, "btxu/mooncake:v3", "other/mooncake:v3"),
		Entry("an empty version keeps the template image",
			datav1alpha1.VersionSpec{}, "btxu/mooncake:v3", "btxu/mooncake:v3"),
		Entry("a partial version against a digest-pinned template keeps the template image",
			datav1alpha1.VersionSpec{ImageTag: "v9"}, digestImage, digestImage),
	)

	It("should return an error when neither the runtime nor the class names a worker image", func() {
		runtimeClass := &datav1alpha1.CacheRuntimeClass{Topology: &datav1alpha1.RuntimeTopology{}}

		_, err := (&CacheEngine{}).getDataOperationImage(
			runtimeWithWorkerVersion(datav1alpha1.VersionSpec{ImageTag: "v9"}), runtimeClass)

		Expect(err).To(HaveOccurred())
	})
})
