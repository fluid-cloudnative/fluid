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
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/tieredstore"
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
				Expect(podSpec.Volumes[0].Name).To(Equal(getMemoryTieredStoreVolumeName(0)))
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
				Expect(podSpec.Volumes[0].Name).To(Equal(getTieredStoreVolumeName(0, 0)))
				Expect(podSpec.Volumes[0].HostPath).NotTo(BeNil())
				Expect(podSpec.Volumes[0].HostPath.Path).To(Equal("/mnt/cache1"))

				// Verify volume mount is created
				Expect(podSpec.Containers[0].VolumeMounts).To(HaveLen(1))
				Expect(podSpec.Containers[0].VolumeMounts[0].Name).To(Equal(getTieredStoreVolumeName(0, 0)))
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
					Expect(podSpec.Volumes[i].Name).To(Equal(getTieredStoreVolumeName(0, i)))
					Expect(podSpec.Volumes[i].HostPath.Path).To(Equal(fmt.Sprintf("/mnt/cache%d", i+1)))
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
				Expect(podSpec.Volumes[0].Name).To(Equal(getTieredStoreVolumeName(0, 0)))
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

	// chargedTieredStoreMemoryQuota recovers from a workload the quota that
	// TransformRuntimeTieredStore charged to container memory. It must select exactly
	// the volumes that carry that quota, because the sync path adds whatever it returns
	// on top of the user's baseline.
	Describe("chargedTieredStoreMemoryQuota", func() {
		memoryVolume := func(name, size string) corev1.Volume {
			quota := resource.MustParse(size)
			return corev1.Volume{
				Name: name,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium:    corev1.StorageMediumMemory,
						SizeLimit: &quota,
					},
				},
			}
		}

		It("should return zero for a pod spec without volumes", func() {
			empty := chargedTieredStoreMemoryQuota(&corev1.PodSpec{})
			Expect(empty.IsZero()).To(BeTrue())
		})

		It("should sum the memory-backed tiered store volumes", func() {
			podSpec := &corev1.PodSpec{Volumes: []corev1.Volume{
				memoryVolume("tiered-store-level-0-memory", "8Gi"),
				memoryVolume("tiered-store-level-1-index-0", "2Gi"),
			}}

			quota := chargedTieredStoreMemoryQuota(podSpec)
			Expect(quota.String()).To(Equal("10Gi"))
		})

		It("should round-trip the quota that the transform charged", func() {
			podSpec := &corev1.PodSpec{Containers: []corev1.Container{{
				Name: "worker",
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("4Gi")},
				},
			}}}
			tieredStore := &datav1alpha1.RuntimeTieredStore{
				Levels: []datav1alpha1.RuntimeTieredStoreLevel{
					{ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{Quota: resource.MustParse("8Gi")}},
				},
			}

			Expect(engine.TransformRuntimeTieredStore(tieredStore, podSpec)).To(Succeed())

			// The container was charged 4Gi + 8Gi, and the quota is recoverable from the volume.
			limit := podSpec.Containers[0].Resources.Limits[corev1.ResourceMemory]
			Expect(limit.String()).To(Equal("12Gi"))
			recovered := chargedTieredStoreMemoryQuota(podSpec)
			Expect(recovered.String()).To(Equal("8Gi"))
		})

		It("should ignore volumes the transform never charges to memory", func() {
			hostPathQuota := resource.MustParse("100Gi")
			diskQuota := resource.MustParse("50Gi")
			podSpec := &corev1.PodSpec{Volumes: []corev1.Volume{
				memoryVolume("tiered-store-level-0-memory", "8Gi"),
				// hostPath levels are not charged to container memory
				{
					Name:         "tiered-store-level-1-index-0",
					VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/mnt/disk"}},
				},
				// a disk-backed emptyDir level is not charged either
				{
					Name: "tiered-store-level-2-index-0",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{SizeLimit: &diskQuota},
					},
				},
				// a memory volume with no size limit contributes nothing
				{
					Name: "tiered-store-level-3-memory",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{Medium: corev1.StorageMediumMemory},
					},
				},
				// a tmpfs the CacheRuntimeClass template declares itself is not a tiered
				// store level, and charging it would inflate the container on every sync
				{
					Name: "user-defined-shm",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{
							Medium:    corev1.StorageMediumMemory,
							SizeLimit: &hostPathQuota,
						},
					},
				},
			}}

			quota := chargedTieredStoreMemoryQuota(podSpec)
			Expect(quota.String()).To(Equal("8Gi"))
		})
	})
})

var _ = Describe("CacheEngine convertToLegacyTieredStore Tests", Label("pkg.ddc.cache.engine.transform_tiered_store_test.go"), func() {
	Describe("convertToLegacyTieredStore", func() {
		It("should return no levels for an unset tiered store", func() {
			Expect(convertToLegacyTieredStore(datav1alpha1.RuntimeTieredStore{}).Levels).To(BeEmpty())
		})

		It("should map process memory to the MEM medium", func() {
			legacy := convertToLegacyTieredStore(datav1alpha1.RuntimeTieredStore{
				Levels: []datav1alpha1.RuntimeTieredStoreLevel{
					{
						ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{
							Quota: resource.MustParse("4Gi"),
						},
						High: "0.8",
						Low:  "0.5",
					},
				},
			})

			Expect(legacy.Levels).To(HaveLen(1))
			Expect(legacy.Levels[0].MediumType).To(Equal(common.Memory))
			Expect(legacy.Levels[0].Path).To(Equal(GetMemoryTieredStoreMountPath(0)))
			Expect(legacy.Levels[0].Quota.String()).To(Equal("4Gi"))
			Expect(legacy.Levels[0].High).To(Equal("0.8"))
			Expect(legacy.Levels[0].Low).To(Equal("0.5"))
		})

		It("should map a memory backed emptyDir to the MEM medium", func() {
			legacy := convertToLegacyTieredStore(datav1alpha1.RuntimeTieredStore{
				Levels: []datav1alpha1.RuntimeTieredStoreLevel{
					{
						EmptyDir: &datav1alpha1.EmptyDirMediumSource{
							Quota:  resource.MustParse("2Gi"),
							Medium: corev1.StorageMediumMemory,
						},
					},
				},
			})

			Expect(legacy.Levels).To(HaveLen(1))
			Expect(legacy.Levels[0].MediumType).To(Equal(common.Memory))
			Expect(legacy.Levels[0].Quota.String()).To(Equal("2Gi"))
		})

		It("should map a default emptyDir to a disk medium", func() {
			legacy := convertToLegacyTieredStore(datav1alpha1.RuntimeTieredStore{
				Levels: []datav1alpha1.RuntimeTieredStoreLevel{
					{
						EmptyDir: &datav1alpha1.EmptyDirMediumSource{
							Quota: resource.MustParse("1Gi"),
						},
					},
				},
			})

			Expect(legacy.Levels).To(HaveLen(1))
			Expect(legacy.Levels[0].MediumType).To(Equal(common.HDD))
			Expect(legacy.Levels[0].Path).To(Equal(GetEmptyDirTieredStoreMountPath(0)))
			Expect(legacy.Levels[0].Quota.String()).To(Equal("1Gi"))
		})

		It("should keep host path quotas per path instead of collapsing them into one", func() {
			legacy := convertToLegacyTieredStore(datav1alpha1.RuntimeTieredStore{
				Levels: []datav1alpha1.RuntimeTieredStoreLevel{
					{
						HostPath: &datav1alpha1.HostPathMediumSource{
							Paths:  []string{"/mnt/cache1", "/mnt/cache2"},
							Quotas: []resource.Quantity{resource.MustParse("1Gi"), resource.MustParse("3Gi")},
						},
					},
				},
			})

			Expect(legacy.Levels).To(HaveLen(1))
			Expect(legacy.Levels[0].MediumType).To(Equal(common.HDD))
			Expect(legacy.Levels[0].Path).To(Equal(
				GetHostPathTieredStoreMountPath(0, 0) + "," + GetHostPathTieredStoreMountPath(0, 1)))
			Expect(legacy.Levels[0].QuotaList).To(Equal("1Gi,3Gi"))
			Expect(legacy.Levels[0].Quota).To(BeNil())
		})

		It("should skip a host path level whose paths and quotas disagree in length", func() {
			legacy := convertToLegacyTieredStore(datav1alpha1.RuntimeTieredStore{
				Levels: []datav1alpha1.RuntimeTieredStoreLevel{
					{
						HostPath: &datav1alpha1.HostPathMediumSource{
							Paths:  []string{"/mnt/cache1", "/mnt/cache2"},
							Quotas: []resource.Quantity{resource.MustParse("1Gi")},
						},
					},
				},
			})

			Expect(legacy.Levels).To(BeEmpty())
		})

		It("should skip a level that names no medium", func() {
			legacy := convertToLegacyTieredStore(datav1alpha1.RuntimeTieredStore{
				Levels: []datav1alpha1.RuntimeTieredStoreLevel{
					{High: "0.9"},
					{EmptyDir: &datav1alpha1.EmptyDirMediumSource{Quota: resource.MustParse("1Gi")}},
				},
			})

			Expect(legacy.Levels).To(HaveLen(1))
			Expect(legacy.Levels[0].Quota.String()).To(Equal("1Gi"))
		})

		It("should keep the declared order and index paths per level", func() {
			legacy := convertToLegacyTieredStore(datav1alpha1.RuntimeTieredStore{
				Levels: []datav1alpha1.RuntimeTieredStoreLevel{
					{ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{Quota: resource.MustParse("4Gi")}},
					{EmptyDir: &datav1alpha1.EmptyDirMediumSource{Quota: resource.MustParse("1Gi")}},
				},
			})

			Expect(legacy.Levels).To(HaveLen(2))
			Expect(legacy.Levels[0].MediumType).To(Equal(common.Memory))
			Expect(legacy.Levels[1].MediumType).To(Equal(common.HDD))
			Expect(legacy.Levels[1].Path).To(Equal(GetEmptyDirTieredStoreMountPath(1)))
		})
	})

	Describe("the result feeding base.BuildRuntimeInfo", func() {
		It("should produce levels that convertToTieredstoreInfo accepts and sums correctly", func() {
			legacy := convertToLegacyTieredStore(datav1alpha1.RuntimeTieredStore{
				Levels: []datav1alpha1.RuntimeTieredStoreLevel{
					{ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{Quota: resource.MustParse("4Gi")}},
					{HostPath: &datav1alpha1.HostPathMediumSource{
						Paths:  []string{"/mnt/cache1", "/mnt/cache2"},
						Quotas: []resource.Quantity{resource.MustParse("1Gi"), resource.MustParse("3Gi")},
					}},
				},
			})

			runtimeInfo, err := base.BuildRuntimeInfo("test", "default", common.CacheRuntime,
				base.WithTieredStore(legacy))
			Expect(err).NotTo(HaveOccurred())

			storage := tieredstore.GetLevelStorageMap(runtimeInfo)
			Expect(storage[common.MemoryCacheStore].String()).To(Equal("4Gi"))
			// 1Gi and 3Gi must survive as 4Gi rather than being averaged to 2Gi each
			Expect(storage[common.DiskCacheStore].String()).To(Equal("4Gi"))
		})
	})
})
