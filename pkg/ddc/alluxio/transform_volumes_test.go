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

package alluxio

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// Constants for test values
const (
	testVolumeName   = "test"
	testSecretName   = "test"
	testVolMountPath = "/test"
)

var _ = Describe("AlluxioEngine Transform Volumes Tests", Label("pkg.ddc.alluxio.transform_volumes_test.go"), func() {
	var (
		engine *AlluxioEngine
		got    *Alluxio
	)

	BeforeEach(func() {
		engine = &AlluxioEngine{}
		got = &Alluxio{}
	})

	Describe("transformMasterVolumes", func() {
		Context("when both volumes and volume mounts are provided", func() {
			It("should correctly transform volumes for master", func() {
				runtime := &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Volumes: []corev1.Volume{
							{
								Name: testVolumeName,
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName: testSecretName,
									},
								},
							},
						},
						Master: datav1alpha1.AlluxioCompTemplateSpec{
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      testVolumeName,
									MountPath: testVolMountPath,
								},
							},
						},
					},
				}
				err := engine.transformMasterVolumes(runtime, got)

				Expect(err).NotTo(HaveOccurred())
				Expect(got.Master.Volumes).To(HaveLen(1))
				Expect(got.Master.Volumes[0].Name).To(Equal(testVolumeName))
				Expect(got.Master.Volumes[0].Secret.SecretName).To(Equal(testSecretName))
				Expect(got.Master.VolumeMounts).To(HaveLen(1))
				Expect(got.Master.VolumeMounts[0].Name).To(Equal(testVolumeName))
				Expect(got.Master.VolumeMounts[0].MountPath).To(Equal(testVolMountPath))
			})
		})

		Context("when only volume mounts are provided without corresponding volumes", func() {
			It("should return an error", func() {
				runtime := &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Master: datav1alpha1.AlluxioCompTemplateSpec{
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      testVolumeName,
									MountPath: testVolMountPath,
								},
							},
						},
					},
				}
				err := engine.transformMasterVolumes(runtime, got)

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("transformWorkerVolumes", func() {
		Context("when both volumes and volume mounts are provided", func() {
			It("should correctly transform volumes for worker", func() {
				runtime := &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Volumes: []corev1.Volume{
							{
								Name: testVolumeName,
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName: testSecretName,
									},
								},
							},
						},
						Worker: datav1alpha1.AlluxioCompTemplateSpec{
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      testVolumeName,
									MountPath: testVolMountPath,
								},
							},
						},
					},
				}
				err := engine.transformWorkerVolumes(runtime, got)

				Expect(err).NotTo(HaveOccurred())
				Expect(got.Worker.Volumes).To(HaveLen(1))
				Expect(got.Worker.Volumes[0].Name).To(Equal(testVolumeName))
				Expect(got.Worker.Volumes[0].Secret.SecretName).To(Equal(testSecretName))
				Expect(got.Worker.VolumeMounts).To(HaveLen(1))
				Expect(got.Worker.VolumeMounts[0].Name).To(Equal(testVolumeName))
				Expect(got.Worker.VolumeMounts[0].MountPath).To(Equal(testVolMountPath))
			})
		})

		Context("when only volume mounts are provided without corresponding volumes", func() {
			It("should return an error", func() {
				runtime := &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Worker: datav1alpha1.AlluxioCompTemplateSpec{
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      testVolumeName,
									MountPath: testVolMountPath,
								},
							},
						},
					},
				}
				err := engine.transformWorkerVolumes(runtime, got)

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("transformFuseVolumes", func() {
		Context("when both volumes and volume mounts are provided", func() {
			It("should correctly transform volumes for fuse", func() {
				runtime := &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Volumes: []corev1.Volume{
							{
								Name: testVolumeName,
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName: testSecretName,
									},
								},
							},
						},
						Fuse: datav1alpha1.AlluxioFuseSpec{
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      testVolumeName,
									MountPath: testVolMountPath,
								},
							},
						},
					},
				}
				err := engine.transformFuseVolumes(runtime, got)

				Expect(err).NotTo(HaveOccurred())
				Expect(got.Fuse.Volumes).To(HaveLen(1))
				Expect(got.Fuse.Volumes[0].Name).To(Equal(testVolumeName))
				Expect(got.Fuse.Volumes[0].Secret.SecretName).To(Equal(testSecretName))
				Expect(got.Fuse.VolumeMounts).To(HaveLen(1))
				Expect(got.Fuse.VolumeMounts[0].Name).To(Equal(testVolumeName))
				Expect(got.Fuse.VolumeMounts[0].MountPath).To(Equal(testVolMountPath))
			})
		})

		Context("when only volume mounts are provided without corresponding volumes", func() {
			It("should return an error", func() {
				runtime := &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Fuse: datav1alpha1.AlluxioFuseSpec{
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      testVolumeName,
									MountPath: testVolMountPath,
								},
							},
						},
					},
				}
				err := engine.transformFuseVolumes(runtime, got)

				Expect(err).To(HaveOccurred())
			})
		})
	})
})
