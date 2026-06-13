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

	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("CacheEngine Transform Volumes Tests", Label("pkg.ddc.cache.engine.transform_volumes_test.go"), func() {
	var (
		engine  *CacheEngine
		podSpec *corev1.PodSpec
	)

	BeforeEach(func() {
		engine = &CacheEngine{}
		podSpec = &corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: "test-image:latest",
				},
			},
		}
	})

	Describe("transformRuntimeSpecVolumes", func() {
		Context("when handling volumes and volumeMounts", func() {
			It("should only add volumes that are referenced by volumeMounts", func() {
				volumes := []corev1.Volume{
					{
						Name: "used-volume",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					},
					{
						Name: "unused-volume",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					},
				}
				volumeMounts := []corev1.VolumeMount{
					{
						Name:      "used-volume",
						MountPath: "/mnt/used",
					},
				}

				err := engine.transformRuntimeSpecVolumes(volumes, volumeMounts, podSpec)
				Expect(err).NotTo(HaveOccurred())

				// Should only have the used volume
				Expect(podSpec.Volumes).To(HaveLen(1))
				Expect(podSpec.Volumes[0].Name).To(Equal("used-volume"))

				// Should have the volume mount
				Expect(podSpec.Containers[0].VolumeMounts).To(HaveLen(1))
				Expect(podSpec.Containers[0].VolumeMounts[0].Name).To(Equal("used-volume"))
			})

			It("should return error when volumeMount references non-existent volume", func() {
				volumes := []corev1.Volume{
					{
						Name: "existing-volume",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					},
				}
				volumeMounts := []corev1.VolumeMount{
					{
						Name:      "non-existent-volume",
						MountPath: "/mnt/nonexistent",
					},
				}

				err := engine.transformRuntimeSpecVolumes(volumes, volumeMounts, podSpec)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("volume not found for volumeMount non-existent-volume"))
			})

			It("should not add any volumes when volumeMounts is empty", func() {
				volumes := []corev1.Volume{
					{
						Name: "some-volume",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					},
				}
				volumeMounts := []corev1.VolumeMount{}

				err := engine.transformRuntimeSpecVolumes(volumes, volumeMounts, podSpec)
				Expect(err).NotTo(HaveOccurred())

				// No volumes should be added
				Expect(podSpec.Volumes).To(BeEmpty())
				Expect(podSpec.Containers[0].VolumeMounts).To(BeEmpty())
			})

			It("should not add duplicate volumes that already exist in podSpec", func() {
				// Pre-add a volume to podSpec
				podSpec.Volumes = []corev1.Volume{
					{
						Name: "existing-volume",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					},
				}

				volumes := []corev1.Volume{
					{
						Name: "existing-volume",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{},
						},
					},
					{
						Name: "new-volume",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					},
				}
				volumeMounts := []corev1.VolumeMount{
					{
						Name:      "existing-volume",
						MountPath: "/mnt/existing",
					},
					{
						Name:      "new-volume",
						MountPath: "/mnt/new",
					},
				}

				err := engine.transformRuntimeSpecVolumes(volumes, volumeMounts, podSpec)
				Expect(err).NotTo(HaveOccurred())

				// Should have 2 volumes total (1 existing + 1 new)
				Expect(podSpec.Volumes).To(HaveLen(2))
				volumeNames := []string{podSpec.Volumes[0].Name, podSpec.Volumes[1].Name}
				Expect(volumeNames).To(ContainElements("existing-volume", "new-volume"))

				// Should have 2 volume mounts
				Expect(podSpec.Containers[0].VolumeMounts).To(HaveLen(2))
			})

			It("should handle mixed scenario with some volumes existing and some missing", func() {
				// Pre-add one volume to podSpec
				podSpec.Volumes = []corev1.Volume{
					{
						Name: "pre-existing-volume",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					},
				}

				volumes := []corev1.Volume{
					{
						Name: "pre-existing-volume",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{},
						},
					},
					{
						Name: "new-volume-1",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					},
					{
						Name: "new-volume-2",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{},
						},
					},
				}
				volumeMounts := []corev1.VolumeMount{
					{
						Name:      "pre-existing-volume",
						MountPath: "/mnt/pre",
					},
					{
						Name:      "new-volume-1",
						MountPath: "/mnt/new1",
					},
					{
						Name:      "new-volume-2",
						MountPath: "/mnt/new2",
					},
				}

				err := engine.transformRuntimeSpecVolumes(volumes, volumeMounts, podSpec)
				Expect(err).NotTo(HaveOccurred())

				// Should have 3 volumes total (1 pre-existing + 2 new)
				Expect(podSpec.Volumes).To(HaveLen(3))

				// Should have 3 volume mounts
				Expect(podSpec.Containers[0].VolumeMounts).To(HaveLen(3))
			})

			It("should return error when containers is empty", func() {
				emptyPodSpec := &corev1.PodSpec{
					Containers: []corev1.Container{},
				}

				volumes := []corev1.Volume{
					{
						Name: "test-volume",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					},
				}
				volumeMounts := []corev1.VolumeMount{
					{
						Name:      "test-volume",
						MountPath: "/mnt/test",
					},
				}

				err := engine.transformRuntimeSpecVolumes(volumes, volumeMounts, emptyPodSpec)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("podTemplateSpec does not have any containers"))
			})

			It("should handle empty volumes and empty volumeMounts", func() {
				volumes := []corev1.Volume{}
				volumeMounts := []corev1.VolumeMount{}

				err := engine.transformRuntimeSpecVolumes(volumes, volumeMounts, podSpec)
				Expect(err).NotTo(HaveOccurred())

				// No changes should be made
				Expect(podSpec.Volumes).To(BeEmpty())
				Expect(podSpec.Containers[0].VolumeMounts).To(BeEmpty())
			})

			It("should preserve existing volumes in podSpec", func() {
				// Pre-add volumes to podSpec
				podSpec.Volumes = []corev1.Volume{
					{
						Name: "runtime-config",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{},
						},
					},
				}

				volumes := []corev1.Volume{
					{
						Name: "app-volume",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					},
				}
				volumeMounts := []corev1.VolumeMount{
					{
						Name:      "app-volume",
						MountPath: "/mnt/app",
					},
				}

				err := engine.transformRuntimeSpecVolumes(volumes, volumeMounts, podSpec)
				Expect(err).NotTo(HaveOccurred())

				// Should have 2 volumes (1 existing + 1 new)
				Expect(podSpec.Volumes).To(HaveLen(2))
				volumeNames := []string{podSpec.Volumes[0].Name, podSpec.Volumes[1].Name}
				Expect(volumeNames).To(ContainElements("runtime-config", "app-volume"))
			})

			It("should return error when multiple volumeMounts reference non-existent volumes", func() {
				volumes := []corev1.Volume{
					{
						Name: "valid-volume",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					},
				}
				volumeMounts := []corev1.VolumeMount{
					{
						Name:      "valid-volume",
						MountPath: "/mnt/valid",
					},
					{
						Name:      "missing-volume-1",
						MountPath: "/mnt/missing1",
					},
					{
						Name:      "missing-volume-2",
						MountPath: "/mnt/missing2",
					},
				}

				err := engine.transformRuntimeSpecVolumes(volumes, volumeMounts, podSpec)
				Expect(err).To(HaveOccurred())
				// Should report one of the missing volumes
				Expect(err.Error()).To(Or(
					ContainSubstring("volume not found for volumeMount missing-volume-1"),
					ContainSubstring("volume not found for volumeMount missing-volume-2"),
				))
			})
		})
	})
})
