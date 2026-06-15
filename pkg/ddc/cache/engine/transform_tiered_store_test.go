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
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("CacheEngine TransformRuntimeTieredStore Tests", Label("pkg.ddc.cache.engine.transform_tiered_store_test.go"), func() {
	var (
		engine  *CacheEngine
		podSpec *corev1.PodSpec
	)

	BeforeEach(func() {
		engine = &CacheEngine{}

		// Initialize a basic pod spec with one container
		podSpec = &corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "worker",
					Image: "test-image:latest",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{},
						Limits:   corev1.ResourceList{},
					},
				},
			},
		}
	})

	Describe("TransformRuntimeTieredStore", func() {
		Context("when tiered store has no levels", func() {
			It("should return nil without error", func() {
				tieredStore := &datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{},
				}

				err := engine.TransformRuntimeTieredStore(tieredStore, podSpec)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when pod spec has no containers", func() {
			It("should return error", func() {
				emptyPodSpec := &corev1.PodSpec{
					Containers: []corev1.Container{},
				}

				tieredStore := &datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{
							EmptyDir: &datav1alpha1.EmptyDirMediumSource{
								Quota: resource.MustParse("1Gi"),
							},
						},
					},
				}

				err := engine.TransformRuntimeTieredStore(tieredStore, emptyPodSpec)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no containers found"))
			})
		})

		Context("when using ProcessMemory medium", func() {
			It("should add memory resources to container", func() {
				tieredStore := &datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{
							ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{
								Quota: resource.MustParse("4Gi"),
							},
							High: "0.9",
							Low:  "0.7",
						},
					},
				}

				// Set initial memory resources
				podSpec.Containers[0].Resources.Requests[corev1.ResourceMemory] = resource.MustParse("2Gi")
				podSpec.Containers[0].Resources.Limits[corev1.ResourceMemory] = resource.MustParse("4Gi")

				err := engine.TransformRuntimeTieredStore(tieredStore, podSpec)
				Expect(err).NotTo(HaveOccurred())

				// Verify memory request is increased
				memRequest := podSpec.Containers[0].Resources.Requests[corev1.ResourceMemory]
				expectedRequest := resource.MustParse("6Gi") // 2Gi + 4Gi
				Expect(memRequest.Cmp(expectedRequest)).To(Equal(0))

				// Verify memory limit is increased
				memLimit := podSpec.Containers[0].Resources.Limits[corev1.ResourceMemory]
				expectedLimit := resource.MustParse("8Gi") // 4Gi + 4Gi
				Expect(memLimit.Cmp(expectedLimit)).To(Equal(0))

				// Verify volume and volume mount are created for ProcessMemory
				Expect(podSpec.Volumes).To(HaveLen(1))
				Expect(podSpec.Volumes[0].Name).To(Equal("tiered-store-level-0-memory"))
				Expect(podSpec.Volumes[0].EmptyDir).NotTo(BeNil())
				Expect(podSpec.Volumes[0].EmptyDir.Medium).To(Equal(corev1.StorageMediumMemory))

				// Verify volume mount path is set to /dev/shm (from GetMemoryTieredStoreMountPath)
				Expect(podSpec.Containers[0].VolumeMounts).To(HaveLen(1))
				Expect(podSpec.Containers[0].VolumeMounts[0].MountPath).To(Equal("/dev/shm"))
			})

			It("should return error when quota is zero", func() {
				tieredStore := &datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{
							ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{
								Quota: resource.MustParse("0"),
							},
						},
					},
				}

				err := engine.TransformRuntimeTieredStore(tieredStore, podSpec)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("quota cannot be zero"))
			})

			It("should not modify resources when container has no memory constraints", func() {
				tieredStore := &datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{
							ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{
								Quota: resource.MustParse("4Gi"),
							},
						},
					},
				}

				// Container has no memory resources set
				podSpec.Containers[0].Resources.Requests = nil
				podSpec.Containers[0].Resources.Limits = nil

				err := engine.TransformRuntimeTieredStore(tieredStore, podSpec)
				Expect(err).NotTo(HaveOccurred())

				// Resources should remain nil
				Expect(podSpec.Containers[0].Resources.Requests).To(BeNil())
				Expect(podSpec.Containers[0].Resources.Limits).To(BeNil())
			})

			It("should return error when multiple processMemory levels are specified", func() {
				tieredStore := &datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{
							ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{
								Quota: resource.MustParse("4Gi"),
							},
						},
						{
							ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{
								Quota: resource.MustParse("2Gi"),
							},
						},
					},
				}

				err := engine.TransformRuntimeTieredStore(tieredStore, podSpec)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("only one ProcessMemoryMediumSource"))
			})
		})

		Context("when using HostPath medium", func() {
			It("should create volumes and volume mounts for single path", func() {
				tieredStore := &datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{
							HostPath: &datav1alpha1.HostPathMediumSource{
								Paths:  []string{"/mnt/cache1"},
								Quotas: []resource.Quantity{resource.MustParse("100Gi")},
								Type:   nil,
							},
							High: "0.9",
							Low:  "0.7",
						},
					},
				}

				err := engine.TransformRuntimeTieredStore(tieredStore, podSpec)
				Expect(err).NotTo(HaveOccurred())

				// Verify volume is created
				Expect(podSpec.Volumes).To(HaveLen(1))
				Expect(podSpec.Volumes[0].Name).To(Equal("tiered-store-level-0-index-0"))
				Expect(podSpec.Volumes[0].HostPath).NotTo(BeNil())
				Expect(podSpec.Volumes[0].HostPath.Path).To(Equal("/mnt/cache1"))

				// Verify volume mount is created
				Expect(podSpec.Containers[0].VolumeMounts).To(HaveLen(1))
				Expect(podSpec.Containers[0].VolumeMounts[0].Name).To(Equal("tiered-store-level-0-index-0"))
				Expect(podSpec.Containers[0].VolumeMounts[0].MountPath).To(ContainSubstring("tiered-store"))
			})

			It("should create multiple volumes and mounts for multiple paths", func() {
				tieredStore := &datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{
							HostPath: &datav1alpha1.HostPathMediumSource{
								Paths: []string{"/mnt/cache1", "/mnt/cache2", "/mnt/cache3"},
								Quotas: []resource.Quantity{
									resource.MustParse("100Gi"),
									resource.MustParse("200Gi"),
									resource.MustParse("300Gi"),
								},
							},
						},
					},
				}

				err := engine.TransformRuntimeTieredStore(tieredStore, podSpec)
				Expect(err).NotTo(HaveOccurred())

				// Verify 3 volumes are created
				Expect(podSpec.Volumes).To(HaveLen(3))
				for i := 0; i < 3; i++ {
					Expect(podSpec.Volumes[i].Name).To(Equal("tiered-store-level-0-index-" + string(rune('0'+i))))
					Expect(podSpec.Volumes[i].HostPath.Path).To(Equal("/mnt/cache" + string(rune('1'+i))))
				}

				// Verify 3 volume mounts are created
				Expect(podSpec.Containers[0].VolumeMounts).To(HaveLen(3))
			})

			It("should return error when paths and quotas count mismatch", func() {
				tieredStore := &datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{
							HostPath: &datav1alpha1.HostPathMediumSource{
								Paths:  []string{"/mnt/cache1", "/mnt/cache2"},
								Quotas: []resource.Quantity{resource.MustParse("100Gi")},
							},
						},
					},
				}

				err := engine.TransformRuntimeTieredStore(tieredStore, podSpec)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("number of paths and quotas must be equal"))
			})
		})

		Context("when using EmptyDir medium", func() {
			It("should create volume and mount for disk-based EmptyDir", func() {
				tieredStore := &datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{
							EmptyDir: &datav1alpha1.EmptyDirMediumSource{
								Quota:  resource.MustParse("50Gi"),
								Medium: corev1.StorageMediumDefault,
							},
							High: "0.85",
							Low:  "0.65",
						},
					},
				}

				err := engine.TransformRuntimeTieredStore(tieredStore, podSpec)
				Expect(err).NotTo(HaveOccurred())

				// Verify volume is created
				Expect(podSpec.Volumes).To(HaveLen(1))
				Expect(podSpec.Volumes[0].Name).To(Equal("tiered-store-level-0-index-0"))
				Expect(podSpec.Volumes[0].EmptyDir).NotTo(BeNil())
				Expect(podSpec.Volumes[0].EmptyDir.Medium).To(Equal(corev1.StorageMediumDefault))

				// Verify SizeLimit is set
				Expect(podSpec.Volumes[0].EmptyDir.SizeLimit).NotTo(BeNil())
				Expect(podSpec.Volumes[0].EmptyDir.SizeLimit.String()).To(Equal("50Gi"))

				// Verify volume mount is created
				Expect(podSpec.Containers[0].VolumeMounts).To(HaveLen(1))
			})

			It("should add memory resources for Memory-backed EmptyDir", func() {
				tieredStore := &datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{
							EmptyDir: &datav1alpha1.EmptyDirMediumSource{
								Quota:  resource.MustParse("2Gi"),
								Medium: corev1.StorageMediumMemory,
							},
						},
					},
				}

				// Set initial memory resources
				podSpec.Containers[0].Resources.Requests[corev1.ResourceMemory] = resource.MustParse("1Gi")
				podSpec.Containers[0].Resources.Limits[corev1.ResourceMemory] = resource.MustParse("2Gi")

				err := engine.TransformRuntimeTieredStore(tieredStore, podSpec)
				Expect(err).NotTo(HaveOccurred())

				// Verify memory request is increased
				memRequest := podSpec.Containers[0].Resources.Requests[corev1.ResourceMemory]
				expectedRequest := resource.MustParse("3Gi") // 1Gi + 2Gi
				Expect(memRequest.Cmp(expectedRequest)).To(Equal(0))

				// Verify memory limit is increased
				memLimit := podSpec.Containers[0].Resources.Limits[corev1.ResourceMemory]
				expectedLimit := resource.MustParse("4Gi") // 2Gi + 2Gi
				Expect(memLimit.Cmp(expectedLimit)).To(Equal(0))
			})

			It("should return error when quota is zero", func() {
				tieredStore := &datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{
							EmptyDir: &datav1alpha1.EmptyDirMediumSource{
								Quota: resource.MustParse("0"),
							},
						},
					},
				}

				err := engine.TransformRuntimeTieredStore(tieredStore, podSpec)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("quota cannot be zero"))
			})
		})

		Context("when using multiple tier levels", func() {
			It("should process all levels correctly", func() {
				tieredStore := &datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						// Level 0: ProcessMemory
						{
							ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{
								Quota: resource.MustParse("4Gi"),
							},
						},
						// Level 1: HostPath
						{
							HostPath: &datav1alpha1.HostPathMediumSource{
								Paths:  []string{"/mnt/ssd1"},
								Quotas: []resource.Quantity{resource.MustParse("100Gi")},
							},
						},
						// Level 2: EmptyDir
						{
							EmptyDir: &datav1alpha1.EmptyDirMediumSource{
								Quota: resource.MustParse("50Gi"),
							},
						},
					},
				}

				// Set initial memory resources
				podSpec.Containers[0].Resources.Requests[corev1.ResourceMemory] = resource.MustParse("2Gi")
				podSpec.Containers[0].Resources.Limits[corev1.ResourceMemory] = resource.MustParse("4Gi")

				err := engine.TransformRuntimeTieredStore(tieredStore, podSpec)
				Expect(err).NotTo(HaveOccurred())

				// Verify 3 volumes are created (ProcessMemory's EmptyDir + HostPath + EmptyDir)
				Expect(podSpec.Volumes).To(HaveLen(3))

				// Verify 3 volume mounts are created
				Expect(podSpec.Containers[0].VolumeMounts).To(HaveLen(3))

				// Verify memory resources include ProcessMemory quota
				memRequest := podSpec.Containers[0].Resources.Requests[corev1.ResourceMemory]
				expectedRequest := resource.MustParse("6Gi") // 2Gi + 4Gi
				Expect(memRequest.Cmp(expectedRequest)).To(Equal(0))
			})

			It("should return error when multiple media types in same level", func() {
				tieredStore := &datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{
							// Both ProcessMemory and EmptyDir specified - invalid
							ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{
								Quota: resource.MustParse("4Gi"),
							},
							EmptyDir: &datav1alpha1.EmptyDirMediumSource{
								Quota: resource.MustParse("50Gi"),
							},
						},
					},
				}

				err := engine.TransformRuntimeTieredStore(tieredStore, podSpec)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("only one storage medium can be specified"))
			})
		})

		Context("when handling high/low watermarks", func() {
			It("should preserve watermark configuration in level", func() {
				tieredStore := &datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{
							EmptyDir: &datav1alpha1.EmptyDirMediumSource{
								Quota: resource.MustParse("100Gi"),
							},
							High: "0.9",
							Low:  "0.7",
						},
					},
				}

				err := engine.TransformRuntimeTieredStore(tieredStore, podSpec)
				Expect(err).NotTo(HaveOccurred())

				// Watermarks are stored in the tieredStore config, not in pod spec
				// This test verifies that processing doesn't fail with watermarks present
				Expect(tieredStore.Levels[0].High).To(Equal("0.9"))
				Expect(tieredStore.Levels[0].Low).To(Equal("0.7"))
			})
		})
	})
})
