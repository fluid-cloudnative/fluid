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
)

// Constants for test values
const (
	testSecretName1  = "test-secret-1"
	testSecretName2  = "test-secret-2"
	testSecretKey    = "access-key"
	testMountName    = "test-mount"
	testMountPoint   = "s3://test-bucket"
	nativeMountPoint = "local:///mnt/test"
)

var _ = Describe("CacheEngine Transform Volumes Tests", Label("pkg.ddc.cache.engine.transform_volumes_test.go"), func() {
	var (
		engine *CacheEngine
		value  *common.CacheRuntimeValue
	)

	BeforeEach(func() {
		engine = &CacheEngine{}
		value = &common.CacheRuntimeValue{
			Master: &common.CacheRuntimeComponentValue{
				Enabled: true,
				PodTemplateSpec: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "master",
							},
						},
					},
				},
			},
		}
	})

	Describe("transformEncryptOptionsToMasterVolumes", func() {
		Context("when dataset has shared encrypt options", func() {
			It("should correctly transform shared encrypt options to master volumes", func() {
				dataset := &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: testMountPoint,
								Name:       testMountName,
							},
						},
						SharedEncryptOptions: []datav1alpha1.EncryptOption{
							{
								Name: "aws-access-key-id",
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: testSecretName1,
										Key:  testSecretKey,
									},
								},
							},
						},
					},
				}

				engine.transformEncryptOptionsToComponentVolumes(dataset, value.Master)

				Expect(value.Master.PodTemplateSpec.Spec.Volumes).To(HaveLen(1))
				Expect(value.Master.PodTemplateSpec.Spec.Volumes[0].Name).To(Equal(secretVolumeNamePrefix + testSecretName1))
				Expect(value.Master.PodTemplateSpec.Spec.Volumes[0].Secret.SecretName).To(Equal(testSecretName1))

				Expect(value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts).To(HaveLen(1))
				Expect(value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts[0].Name).To(Equal(secretVolumeNamePrefix + testSecretName1))
				Expect(value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts[0].MountPath).To(Equal("/etc/fluid/secrets/" + testSecretName1))
				Expect(value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts[0].ReadOnly).To(BeTrue())
			})
		})

		Context("when dataset has mount-specific encrypt options", func() {
			It("should correctly transform mount encrypt options to master volumes", func() {
				dataset := &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: testMountPoint,
								Name:       testMountName,
								EncryptOptions: []datav1alpha1.EncryptOption{
									{
										Name: "aws-secret-access-key",
										ValueFrom: datav1alpha1.EncryptOptionSource{
											SecretKeyRef: datav1alpha1.SecretKeySelector{
												Name: testSecretName2,
												Key:  testSecretKey,
											},
										},
									},
								},
							},
						},
					},
				}

				engine.transformEncryptOptionsToComponentVolumes(dataset, value.Master)

				Expect(value.Master.PodTemplateSpec.Spec.Volumes).To(HaveLen(1))
				Expect(value.Master.PodTemplateSpec.Spec.Volumes[0].Name).To(Equal(secretVolumeNamePrefix + testSecretName2))
				Expect(value.Master.PodTemplateSpec.Spec.Volumes[0].Secret.SecretName).To(Equal(testSecretName2))

				Expect(value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts).To(HaveLen(1))
				Expect(value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts[0].Name).To(Equal(secretVolumeNamePrefix + testSecretName2))
				Expect(value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts[0].MountPath).To(Equal("/etc/fluid/secrets/" + testSecretName2))
				Expect(value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts[0].ReadOnly).To(BeTrue())
			})
		})

		Context("when dataset has both shared and mount-specific encrypt options", func() {
			It("should correctly transform all encrypt options to master volumes", func() {
				dataset := &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: testMountPoint,
								Name:       testMountName,
								EncryptOptions: []datav1alpha1.EncryptOption{
									{
										Name: "aws-secret-access-key",
										ValueFrom: datav1alpha1.EncryptOptionSource{
											SecretKeyRef: datav1alpha1.SecretKeySelector{
												Name: testSecretName2,
												Key:  testSecretKey,
											},
										},
									},
								},
							},
						},
						SharedEncryptOptions: []datav1alpha1.EncryptOption{
							{
								Name: "aws-access-key-id",
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: testSecretName1,
										Key:  testSecretKey,
									},
								},
							},
						},
					},
				}

				engine.transformEncryptOptionsToComponentVolumes(dataset, value.Master)

				Expect(value.Master.PodTemplateSpec.Spec.Volumes).To(HaveLen(2))
				volumeNames := []string{
					value.Master.PodTemplateSpec.Spec.Volumes[0].Name,
					value.Master.PodTemplateSpec.Spec.Volumes[1].Name,
				}
				Expect(volumeNames).To(ContainElements(
					secretVolumeNamePrefix+testSecretName1,
					secretVolumeNamePrefix+testSecretName2,
				))

				Expect(value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts).To(HaveLen(2))
				mountNames := []string{
					value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts[0].Name,
					value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts[1].Name,
				}
				Expect(mountNames).To(ContainElements(
					secretVolumeNamePrefix+testSecretName1,
					secretVolumeNamePrefix+testSecretName2,
				))
			})
		})

		Context("when dataset has native fluid scheme mount", func() {
			It("should skip native fluid scheme mounts", func() {
				dataset := &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: nativeMountPoint,
								Name:       testMountName,
								EncryptOptions: []datav1alpha1.EncryptOption{
									{
										Name: "some-option",
										ValueFrom: datav1alpha1.EncryptOptionSource{
											SecretKeyRef: datav1alpha1.SecretKeySelector{
												Name: testSecretName1,
												Key:  testSecretKey,
											},
										},
									},
								},
							},
						},
					},
				}

				engine.transformEncryptOptionsToComponentVolumes(dataset, value.Master)

				Expect(value.Master.PodTemplateSpec.Spec.Volumes).To(BeEmpty())
				Expect(value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts).To(BeEmpty())
			})
		})

		Context("when master is disabled", func() {
			It("should not add any volumes", func() {
				value.Master.Enabled = false
				dataset := &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: testMountPoint,
								Name:       testMountName,
								EncryptOptions: []datav1alpha1.EncryptOption{
									{
										Name: "some-option",
										ValueFrom: datav1alpha1.EncryptOptionSource{
											SecretKeyRef: datav1alpha1.SecretKeySelector{
												Name: testSecretName1,
												Key:  testSecretKey,
											},
										},
									},
								},
							},
						},
					},
				}

				engine.transformEncryptOptionsToComponentVolumes(dataset, value.Master)

				Expect(value.Master.PodTemplateSpec.Spec.Volumes).To(BeEmpty())
				Expect(value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts).To(BeEmpty())
			})
		})

		Context("when master is nil", func() {
			It("should not panic", func() {
				value.Master = nil
				dataset := &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: testMountPoint,
								Name:       testMountName,
							},
						},
					},
				}

				// Should not panic
				engine.transformEncryptOptionsToComponentVolumes(dataset, value.Master)
			})
		})

		Context("when same secret is used multiple times", func() {
			It("should override existing volume and volume mount", func() {
				dataset := &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: testMountPoint,
								Name:       testMountName,
								EncryptOptions: []datav1alpha1.EncryptOption{
									{
										Name: "option1",
										ValueFrom: datav1alpha1.EncryptOptionSource{
											SecretKeyRef: datav1alpha1.SecretKeySelector{
												Name: testSecretName1,
												Key:  testSecretKey,
											},
										},
									},
								},
							},
							{
								MountPoint: "s3://another-bucket",
								Name:       "another-mount",
								EncryptOptions: []datav1alpha1.EncryptOption{
									{
										Name: "option2",
										ValueFrom: datav1alpha1.EncryptOptionSource{
											SecretKeyRef: datav1alpha1.SecretKeySelector{
												Name: testSecretName1,
												Key:  testSecretKey,
											},
										},
									},
								},
							},
						},
					},
				}

				engine.transformEncryptOptionsToComponentVolumes(dataset, value.Master)

				// Should only have one volume for the same secret
				Expect(value.Master.PodTemplateSpec.Spec.Volumes).To(HaveLen(1))
				Expect(value.Master.PodTemplateSpec.Spec.Volumes[0].Name).To(Equal(secretVolumeNamePrefix + testSecretName1))

				// Should only have one volume mount for the same secret
				Expect(value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts).To(HaveLen(1))
				Expect(value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts[0].Name).To(Equal(secretVolumeNamePrefix + testSecretName1))
			})
		})

		Context("when encrypt option has empty secret name", func() {
			It("should skip encrypt options with empty secret name in shared encrypt options", func() {
				dataset := &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: testMountPoint,
								Name:       testMountName,
							},
						},
						SharedEncryptOptions: []datav1alpha1.EncryptOption{
							{
								Name: "aws-access-key-id",
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: "", // Empty secret name
										Key:  testSecretKey,
									},
								},
							},
						},
					},
				}

				engine.transformEncryptOptionsToComponentVolumes(dataset, value.Master)

				// Should not add any volumes for empty secret name
				Expect(value.Master.PodTemplateSpec.Spec.Volumes).To(BeEmpty())
				Expect(value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts).To(BeEmpty())
			})

			It("should skip encrypt options with empty secret name in mount encrypt options", func() {
				dataset := &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: testMountPoint,
								Name:       testMountName,
								EncryptOptions: []datav1alpha1.EncryptOption{
									{
										Name: "aws-secret-access-key",
										ValueFrom: datav1alpha1.EncryptOptionSource{
											SecretKeyRef: datav1alpha1.SecretKeySelector{
												Name: "", // Empty secret name
												Key:  testSecretKey,
											},
										},
									},
								},
							},
						},
					},
				}

				engine.transformEncryptOptionsToComponentVolumes(dataset, value.Master)

				// Should not add any volumes for empty secret name
				Expect(value.Master.PodTemplateSpec.Spec.Volumes).To(BeEmpty())
				Expect(value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts).To(BeEmpty())
			})

			It("should skip empty secret names but process valid ones", func() {
				dataset := &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: testMountPoint,
								Name:       testMountName,
								EncryptOptions: []datav1alpha1.EncryptOption{
									{
										Name: "invalid-option",
										ValueFrom: datav1alpha1.EncryptOptionSource{
											SecretKeyRef: datav1alpha1.SecretKeySelector{
												Name: "", // Empty secret name - should be skipped
												Key:  testSecretKey,
											},
										},
									},
									{
										Name: "valid-option",
										ValueFrom: datav1alpha1.EncryptOptionSource{
											SecretKeyRef: datav1alpha1.SecretKeySelector{
												Name: testSecretName1, // Valid secret name
												Key:  testSecretKey,
											},
										},
									},
								},
							},
						},
						SharedEncryptOptions: []datav1alpha1.EncryptOption{
							{
								Name: "another-invalid-option",
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: "", // Empty secret name - should be skipped
										Key:  testSecretKey,
									},
								},
							},
						},
					},
				}

				engine.transformEncryptOptionsToComponentVolumes(dataset, value.Master)

				// Should only add volume for the valid secret name
				Expect(value.Master.PodTemplateSpec.Spec.Volumes).To(HaveLen(1))
				Expect(value.Master.PodTemplateSpec.Spec.Volumes[0].Name).To(Equal(secretVolumeNamePrefix + testSecretName1))
				Expect(value.Master.PodTemplateSpec.Spec.Volumes[0].Secret.SecretName).To(Equal(testSecretName1))

				Expect(value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts).To(HaveLen(1))
				Expect(value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts[0].Name).To(Equal(secretVolumeNamePrefix + testSecretName1))
				Expect(value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts[0].MountPath).To(Equal("/etc/fluid/secrets/" + testSecretName1))
			})
		})
	})

	Describe("transformEncryptOptionsToComponentVolumes for Worker", func() {
		BeforeEach(func() {
			value.Worker = &common.CacheRuntimeComponentValue{
				Enabled: true,
				PodTemplateSpec: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "worker",
							},
						},
					},
				},
			}
		})

		Context("when dataset has shared encrypt options for worker", func() {
			It("should correctly transform shared encrypt options to worker volumes", func() {
				dataset := &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: testMountPoint,
								Name:       testMountName,
							},
						},
						SharedEncryptOptions: []datav1alpha1.EncryptOption{
							{
								Name: "aws-access-key-id",
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: testSecretName1,
										Key:  testSecretKey,
									},
								},
							},
						},
					},
				}

				engine.transformEncryptOptionsToComponentVolumes(dataset, value.Worker)

				Expect(value.Worker.PodTemplateSpec.Spec.Volumes).To(HaveLen(1))
				Expect(value.Worker.PodTemplateSpec.Spec.Volumes[0].Name).To(Equal(secretVolumeNamePrefix + testSecretName1))
				Expect(value.Worker.PodTemplateSpec.Spec.Volumes[0].Secret.SecretName).To(Equal(testSecretName1))

				Expect(value.Worker.PodTemplateSpec.Spec.Containers[0].VolumeMounts).To(HaveLen(1))
				Expect(value.Worker.PodTemplateSpec.Spec.Containers[0].VolumeMounts[0].Name).To(Equal(secretVolumeNamePrefix + testSecretName1))
				Expect(value.Worker.PodTemplateSpec.Spec.Containers[0].VolumeMounts[0].MountPath).To(Equal("/etc/fluid/secrets/" + testSecretName1))
				Expect(value.Worker.PodTemplateSpec.Spec.Containers[0].VolumeMounts[0].ReadOnly).To(BeTrue())
			})
		})

		Context("when worker is disabled", func() {
			It("should not add any volumes", func() {
				value.Worker.Enabled = false
				dataset := &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: testMountPoint,
								Name:       testMountName,
								EncryptOptions: []datav1alpha1.EncryptOption{
									{
										Name: "some-option",
										ValueFrom: datav1alpha1.EncryptOptionSource{
											SecretKeyRef: datav1alpha1.SecretKeySelector{
												Name: testSecretName1,
												Key:  testSecretKey,
											},
										},
									},
								},
							},
						},
					},
				}

				engine.transformEncryptOptionsToComponentVolumes(dataset, value.Worker)

				Expect(value.Worker.PodTemplateSpec.Spec.Volumes).To(BeEmpty())
				Expect(value.Worker.PodTemplateSpec.Spec.Containers[0].VolumeMounts).To(BeEmpty())
			})
		})

		Context("when encrypt option has empty secret name for worker", func() {
			It("should skip encrypt options with empty secret name", func() {
				dataset := &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: testMountPoint,
								Name:       testMountName,
								EncryptOptions: []datav1alpha1.EncryptOption{
									{
										Name: "aws-secret-access-key",
										ValueFrom: datav1alpha1.EncryptOptionSource{
											SecretKeyRef: datav1alpha1.SecretKeySelector{
												Name: "", // Empty secret name
												Key:  testSecretKey,
											},
										},
									},
								},
							},
						},
						SharedEncryptOptions: []datav1alpha1.EncryptOption{
							{
								Name: "aws-access-key-id",
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: "", // Empty secret name
										Key:  testSecretKey,
									},
								},
							},
						},
					},
				}

				engine.transformEncryptOptionsToComponentVolumes(dataset, value.Worker)

				// Should not add any volumes for empty secret name
				Expect(value.Worker.PodTemplateSpec.Spec.Volumes).To(BeEmpty())
				Expect(value.Worker.PodTemplateSpec.Spec.Containers[0].VolumeMounts).To(BeEmpty())
			})
		})
	})

	Describe("shouldMountSecrets helper function", func() {
		Context("when SecretMount config is nil", func() {
			It("should return defaultEnabled value", func() {
				// Test with defaultEnabled = true (for Master/Worker)
				Expect(shouldMountSecrets(nil, true)).To(BeTrue())

				// Test with defaultEnabled = false (for Client)
				Expect(shouldMountSecrets(nil, false)).To(BeFalse())
			})
		})

		Context("when SecretMount config is provided", func() {
			It("should return the configured Enabled value", func() {
				// Test with Enabled = true
				config := &datav1alpha1.SecretMountComponentDependency{
					Enabled: true,
				}
				Expect(shouldMountSecrets(config, false)).To(BeTrue())

				// Test with Enabled = false
				config.Enabled = false
				Expect(shouldMountSecrets(config, true)).To(BeFalse())
			})
		})
	})

	Describe("Client component secret mount behavior", func() {
		var (
			clientValue *common.CacheRuntimeComponentValue
			dataset     *datav1alpha1.Dataset
		)

		BeforeEach(func() {
			clientValue = &common.CacheRuntimeComponentValue{
				Enabled: true,
				PodTemplateSpec: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "client",
							},
						},
					},
				},
			}
			dataset = &datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					SharedEncryptOptions: []datav1alpha1.EncryptOption{
						{
							Name: "test-secret",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: testSecretName1,
									Key:  testSecretKey,
								},
							},
						},
					},
				},
			}
		})

		Context("when Client has no SecretMount configuration (default disabled)", func() {
			It("should not mount secrets to client pod", func() {
				// Simulate Client with nil SecretMount (default behavior)
				if shouldMountSecrets(nil, false) {
					engine.transformEncryptOptionsToComponentVolumes(dataset, clientValue)
				}

				// Should not add any volumes for Client by default
				Expect(clientValue.PodTemplateSpec.Spec.Volumes).To(BeEmpty())
				Expect(clientValue.PodTemplateSpec.Spec.Containers[0].VolumeMounts).To(BeEmpty())
			})
		})

		Context("when Client has SecretMount explicitly enabled", func() {
			It("should mount secrets to client pod", func() {
				// Simulate Client with SecretMount enabled
				secretMountConfig := &datav1alpha1.SecretMountComponentDependency{
					Enabled: true,
				}
				if shouldMountSecrets(secretMountConfig, false) {
					engine.transformEncryptOptionsToComponentVolumes(dataset, clientValue)
				}

				// Should add volumes for Client when explicitly enabled
				Expect(clientValue.PodTemplateSpec.Spec.Volumes).To(HaveLen(1))
				Expect(clientValue.PodTemplateSpec.Spec.Volumes[0].Name).To(Equal(secretVolumeNamePrefix + testSecretName1))
				Expect(clientValue.PodTemplateSpec.Spec.Containers[0].VolumeMounts).To(HaveLen(1))
			})
		})

		Context("when Client has SecretMount explicitly disabled", func() {
			It("should not mount secrets to client pod", func() {
				// Simulate Client with SecretMount explicitly disabled
				secretMountConfig := &datav1alpha1.SecretMountComponentDependency{
					Enabled: false,
				}
				if shouldMountSecrets(secretMountConfig, false) {
					engine.transformEncryptOptionsToComponentVolumes(dataset, clientValue)
				}

				// Should not add any volumes for Client
				Expect(clientValue.PodTemplateSpec.Spec.Volumes).To(BeEmpty())
				Expect(clientValue.PodTemplateSpec.Spec.Containers[0].VolumeMounts).To(BeEmpty())
			})
		})
	})
})
