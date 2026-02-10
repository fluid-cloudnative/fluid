/*
  Copyright 2022 The Fluid Authors.

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

package juicefs

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

// Constants for test values
const (
	testJuiceVolumeName   = "test"
	testJuiceSecretName   = "test"
	testJuiceMountPath    = "/test"
	testJuiceCachePath    = "/cache"
	testJuiceWorkerCache1 = "/worker-cache1"
	testJuiceWorkerCache2 = "/worker-cache2"
	testJuiceFuseCache1   = "/fuse-cache1"
	testCacheDirPrefix    = "cache-dir-"
)

var _ = Describe("JuiceFSEngine Transform Volume Tests", Label("pkg.ddc.juicefs.transform_volume_test.go"), func() {
	var engine *JuiceFSEngine

	BeforeEach(func() {
		engine = &JuiceFSEngine{}
	})

	Describe("transformWorkerVolumes", func() {
		Context("when both volumes and volume mounts are provided", func() {
			It("should correctly transform volumes for worker", func() {
				runtime := &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						Volumes: []corev1.Volume{
							{
								Name: testJuiceVolumeName,
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName: testJuiceSecretName,
									},
								},
							},
						},
						Worker: datav1alpha1.JuiceFSCompTemplateSpec{
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      testJuiceVolumeName,
									MountPath: testJuiceMountPath,
								},
							},
						},
					},
				}

				got := &JuiceFS{}
				err := engine.transformWorkerVolumes(runtime, got)

				Expect(err).NotTo(HaveOccurred())
				Expect(got.Worker.Volumes).To(HaveLen(1))
				Expect(got.Worker.Volumes[0].Name).To(Equal(testJuiceVolumeName))
				Expect(got.Worker.Volumes[0].Secret.SecretName).To(Equal(testJuiceSecretName))
				Expect(got.Worker.VolumeMounts).To(HaveLen(1))
				Expect(got.Worker.VolumeMounts[0].Name).To(Equal(testJuiceVolumeName))
				Expect(got.Worker.VolumeMounts[0].MountPath).To(Equal(testJuiceMountPath))
			})
		})

		Context("when only volume mounts are provided without corresponding volumes", func() {
			It("should return an error", func() {
				runtime := &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						Worker: datav1alpha1.JuiceFSCompTemplateSpec{
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      testJuiceVolumeName,
									MountPath: testJuiceMountPath,
								},
							},
						},
					},
				}

				got := &JuiceFS{}
				err := engine.transformWorkerVolumes(runtime, got)

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("transformFuseVolumes", func() {
		Context("when both volumes and volume mounts are provided", func() {
			It("should correctly transform volumes for fuse", func() {
				runtime := &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						Volumes: []corev1.Volume{
							{
								Name: testJuiceVolumeName,
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName: testJuiceSecretName,
									},
								},
							},
						},
						Fuse: datav1alpha1.JuiceFSFuseSpec{
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      testJuiceVolumeName,
									MountPath: testJuiceMountPath,
								},
							},
						},
					},
				}

				got := &JuiceFS{}
				err := engine.transformFuseVolumes(runtime, got)

				Expect(err).NotTo(HaveOccurred())
				Expect(got.Fuse.Volumes).To(HaveLen(1))
				Expect(got.Fuse.Volumes[0].Name).To(Equal(testJuiceVolumeName))
				Expect(got.Fuse.Volumes[0].Secret.SecretName).To(Equal(testJuiceSecretName))
				Expect(got.Fuse.VolumeMounts).To(HaveLen(1))
				Expect(got.Fuse.VolumeMounts[0].Name).To(Equal(testJuiceVolumeName))
				Expect(got.Fuse.VolumeMounts[0].MountPath).To(Equal(testJuiceMountPath))
			})
		})

		Context("when only volume mounts are provided without corresponding volumes", func() {
			It("should return an error", func() {
				runtime := &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						Fuse: datav1alpha1.JuiceFSFuseSpec{
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      testJuiceVolumeName,
									MountPath: testJuiceMountPath,
								},
							},
						},
					},
				}

				got := &JuiceFS{}
				err := engine.transformFuseVolumes(runtime, got)

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("transformWorkerCacheVolumes", func() {
		var hostPathDir corev1.HostPathType

		BeforeEach(func() {
			hostPathDir = corev1.HostPathDirectoryOrCreate
		})

		Context("when normal cache directory is specified", func() {
			It("should create hostPath volume and mount", func() {
				runtime := &datav1alpha1.JuiceFSRuntime{}
				value := &JuiceFS{
					CacheDirs: map[string]cache{
						"1": {Path: testJuiceCachePath, Type: string(common.VolumeTypeHostPath)},
					},
				}
				options := map[string]string{"cache-dir": testJuiceCachePath}

				err := engine.transformWorkerCacheVolumes(runtime, value, options)

				Expect(err).NotTo(HaveOccurred())
				Expect(value.Worker.Volumes).To(HaveLen(1))
				Expect(value.Worker.Volumes[0].Name).To(Equal(testCacheDirPrefix + "0"))
				Expect(value.Worker.Volumes[0].HostPath.Path).To(Equal(testJuiceCachePath))
				Expect(*value.Worker.Volumes[0].HostPath.Type).To(Equal(hostPathDir))
				Expect(value.Worker.VolumeMounts).To(HaveLen(1))
				Expect(value.Worker.VolumeMounts[0].Name).To(Equal(testCacheDirPrefix + "0"))
				Expect(value.Worker.VolumeMounts[0].MountPath).To(Equal(testJuiceCachePath))
			})
		})

		Context("when multiple cache directories are specified via options", func() {
			It("should create multiple hostPath volumes and mounts", func() {
				runtime := &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						Worker: datav1alpha1.JuiceFSCompTemplateSpec{},
					},
				}
				value := &JuiceFS{
					CacheDirs: map[string]cache{
						"1": {Path: testJuiceCachePath, Type: string(common.VolumeTypeHostPath)},
					},
				}
				options := map[string]string{"cache-dir": testJuiceCachePath + ":" + testJuiceWorkerCache1 + ":" + testJuiceWorkerCache2}

				err := engine.transformWorkerCacheVolumes(runtime, value, options)

				Expect(err).NotTo(HaveOccurred())
				Expect(value.Worker.Volumes).To(HaveLen(3))
				Expect(value.Worker.VolumeMounts).To(HaveLen(3))
				// Verify volume names and paths
				volumeNames := make([]string, 0, len(value.Worker.Volumes))
				for _, v := range value.Worker.Volumes {
					volumeNames = append(volumeNames, v.Name)
					Expect(v.HostPath).NotTo(BeNil(), "expected HostPath volume type")
				}
				mountNames := make([]string, 0, len(value.Worker.VolumeMounts))
				for _, m := range value.Worker.VolumeMounts {
					mountNames = append(mountNames, m.Name)
				}
				Expect(volumeNames).To(ConsistOf(testCacheDirPrefix+"0", testCacheDirPrefix+"1", testCacheDirPrefix+"2"))
				Expect(mountNames).To(ConsistOf(testCacheDirPrefix+"0", testCacheDirPrefix+"1", testCacheDirPrefix+"2"))
			})
		})

		Context("when runtime has existing volumes that overlap with cache dirs", func() {
			It("should preserve existing volumes and add non-overlapping cache volumes", func() {
				runtime := &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						Worker: datav1alpha1.JuiceFSCompTemplateSpec{
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cache",
									MountPath: testJuiceWorkerCache2,
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "cache",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{},
								},
							},
						},
					},
				}
				value := &JuiceFS{
					CacheDirs: map[string]cache{
						"1": {Path: testJuiceCachePath, Type: string(common.VolumeTypeHostPath)},
					},
					Worker: Worker{
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "cache",
								MountPath: testJuiceWorkerCache2,
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "cache",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{},
								},
							},
						},
					},
				}
				options := map[string]string{"cache-dir": testJuiceWorkerCache1 + ":" + testJuiceWorkerCache2}

				err := engine.transformWorkerCacheVolumes(runtime, value, options)

				Expect(err).NotTo(HaveOccurred())
				Expect(value.Worker.Volumes).To(HaveLen(2))
				Expect(value.Worker.VolumeMounts).To(HaveLen(2))
				// Verify that existing "cache" volume is preserved and new cache-dir volume is added
				volumeNames := make([]string, 0, len(value.Worker.Volumes))
				for _, v := range value.Worker.Volumes {
					volumeNames = append(volumeNames, v.Name)
				}
				Expect(volumeNames).To(ContainElement("cache"), "existing cache volume should be preserved")
				Expect(volumeNames).To(ContainElement(testCacheDirPrefix+"0"), "new cache-dir volume should be added")
			})
		})

		Context("when emptyDir cache type is specified", func() {
			It("should handle emptyDir cache configuration", func() {
				runtime := &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						Worker: datav1alpha1.JuiceFSCompTemplateSpec{},
					},
				}
				value := &JuiceFS{
					CacheDirs: map[string]cache{
						"1": {
							Path: testJuiceCachePath,
							Type: string(common.VolumeTypeEmptyDir),
							VolumeSource: &datav1alpha1.VolumeSource{VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: corev1.StorageMediumMemory,
								},
							}},
						},
					},
					Worker: Worker{},
				}
				options := map[string]string{"cache-dir": testJuiceWorkerCache1}

				err := engine.transformWorkerCacheVolumes(runtime, value, options)

				Expect(err).NotTo(HaveOccurred())
				Expect(value.Worker.Volumes).To(HaveLen(1))
				Expect(value.Worker.VolumeMounts).To(HaveLen(1))
				// Verify volume name and mount path are correct
				Expect(value.Worker.Volumes[0].Name).To(Equal(testCacheDirPrefix + "0"))
				Expect(value.Worker.VolumeMounts[0].Name).To(Equal(testCacheDirPrefix + "0"))
				Expect(value.Worker.VolumeMounts[0].MountPath).To(Equal(testJuiceWorkerCache1))
			})
		})
	})

	Describe("transformFuseCacheVolumes", func() {
		var hostPathDir corev1.HostPathType

		BeforeEach(func() {
			hostPathDir = corev1.HostPathDirectoryOrCreate
		})

		Context("when normal cache directory is specified", func() {
			It("should create hostPath volume and mount for fuse", func() {
				runtime := &datav1alpha1.JuiceFSRuntime{}
				value := &JuiceFS{
					CacheDirs: map[string]cache{
						"1": {Path: testJuiceCachePath, Type: string(common.VolumeTypeHostPath)},
					},
				}
				options := map[string]string{"cache-dir": testJuiceCachePath}

				err := engine.transformFuseCacheVolumes(runtime, value, options)

				Expect(err).NotTo(HaveOccurred())
				Expect(value.Fuse.Volumes).To(HaveLen(1))
				Expect(value.Fuse.Volumes[0].Name).To(Equal(testCacheDirPrefix + "0"))
				Expect(value.Fuse.Volumes[0].HostPath.Path).To(Equal(testJuiceCachePath))
				Expect(*value.Fuse.Volumes[0].HostPath.Type).To(Equal(hostPathDir))
				Expect(value.Fuse.VolumeMounts).To(HaveLen(1))
				Expect(value.Fuse.VolumeMounts[0].Name).To(Equal(testCacheDirPrefix + "0"))
				Expect(value.Fuse.VolumeMounts[0].MountPath).To(Equal(testJuiceCachePath))
			})
		})

		Context("when runtime has existing fuse volumes that overlap with cache dirs", func() {
			It("should not add cache volumes when existing volume covers cache path", func() {
				runtime := &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						Fuse: datav1alpha1.JuiceFSFuseSpec{
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cache",
									MountPath: testJuiceCachePath,
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "cache",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{},
								},
							},
						},
					},
				}
				value := &JuiceFS{
					CacheDirs: map[string]cache{
						"1": {Path: testJuiceCachePath, Type: string(common.VolumeTypeHostPath)},
					},
				}

				err := engine.transformFuseCacheVolumes(runtime, value, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(value.Fuse.Volumes).To(BeEmpty())
				Expect(value.Fuse.VolumeMounts).To(BeEmpty())
			})
		})

		Context("when emptyDir cache type is specified for fuse", func() {
			It("should handle emptyDir cache configuration for fuse", func() {
				runtime := &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						Fuse: datav1alpha1.JuiceFSFuseSpec{},
					},
				}
				value := &JuiceFS{
					CacheDirs: map[string]cache{
						"1": {
							Path: testJuiceCachePath,
							Type: string(common.VolumeTypeEmptyDir),
							VolumeSource: &datav1alpha1.VolumeSource{VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: corev1.StorageMediumMemory,
								},
							}},
						},
					},
				}
				options := map[string]string{"cache-dir": testJuiceFuseCache1}

				err := engine.transformFuseCacheVolumes(runtime, value, options)

				Expect(err).NotTo(HaveOccurred())
				Expect(value.Fuse.Volumes).To(HaveLen(1))
				Expect(value.Fuse.VolumeMounts).To(HaveLen(1))
				// Verify volume name and mount path are correct
				Expect(value.Fuse.Volumes[0].Name).To(Equal(testCacheDirPrefix + "0"))
				Expect(value.Fuse.VolumeMounts[0].Name).To(Equal(testCacheDirPrefix + "0"))
				Expect(value.Fuse.VolumeMounts[0].MountPath).To(Equal(testJuiceFuseCache1))
			})
		})
	})
})
