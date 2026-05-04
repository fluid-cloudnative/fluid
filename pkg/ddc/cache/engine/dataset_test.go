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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("generateDatasetMountOptions Tests", Label("pkg.ddc.cache.engine.dataset_test.go"), func() {
	var engine *CacheEngine

	BeforeEach(func() {
		engine = &CacheEngine{}
	})

	Describe("generateDatasetMountOptions", func() {
		Context("when mount has no options and no encrypt options", func() {
			It("should return empty maps", func() {
				mount := &datav1alpha1.Mount{
					Name:       "test-mount",
					MountPoint: "s3://test-bucket",
				}
				var sharedEncryptOptions []datav1alpha1.EncryptOption
				sharedOptions := map[string]string{}

				mOptions, encryptOptions, err := engine.generateDatasetMountOptions(mount, sharedEncryptOptions, sharedOptions)

				Expect(err).NotTo(HaveOccurred())
				Expect(mOptions).To(BeEmpty())
				Expect(encryptOptions).To(BeEmpty())
			})
		})

		Context("when mount has only options", func() {
			It("should return mount options correctly", func() {
				mount := &datav1alpha1.Mount{
					Name:       "test-mount",
					MountPoint: "s3://test-bucket",
					Options: map[string]string{
						"option1": "value1",
						"option2": "value2",
					},
				}
				var sharedEncryptOptions []datav1alpha1.EncryptOption
				sharedOptions := map[string]string{}

				mOptions, encryptOptions, err := engine.generateDatasetMountOptions(mount, sharedEncryptOptions, sharedOptions)

				Expect(err).NotTo(HaveOccurred())
				Expect(mOptions).To(HaveLen(2))
				Expect(mOptions["option1"]).To(Equal("value1"))
				Expect(mOptions["option2"]).To(Equal("value2"))
				Expect(encryptOptions).To(BeEmpty())
			})
		})

		Context("when shared options exist", func() {
			It("should merge shared options with mount options", func() {
				mount := &datav1alpha1.Mount{
					Name:       "test-mount",
					MountPoint: "s3://test-bucket",
					Options: map[string]string{
						"mount-option": "mount-value",
					},
				}
				var sharedEncryptOptions []datav1alpha1.EncryptOption
				sharedOptions := map[string]string{
					"shared-option": "shared-value",
				}

				mOptions, encryptOptions, err := engine.generateDatasetMountOptions(mount, sharedEncryptOptions, sharedOptions)

				Expect(err).NotTo(HaveOccurred())
				Expect(mOptions).To(HaveLen(2))
				Expect(mOptions["shared-option"]).To(Equal("shared-value"))
				Expect(mOptions["mount-option"]).To(Equal("mount-value"))
				Expect(encryptOptions).To(BeEmpty())
			})

			It("mount options should override shared options with same key", func() {
				mount := &datav1alpha1.Mount{
					Name:       "test-mount",
					MountPoint: "s3://test-bucket",
					Options: map[string]string{
						"common-option": "mount-value",
					},
				}
				var sharedEncryptOptions []datav1alpha1.EncryptOption
				sharedOptions := map[string]string{
					"common-option": "shared-value",
				}

				mOptions, encryptOptions, err := engine.generateDatasetMountOptions(mount, sharedEncryptOptions, sharedOptions)

				Expect(err).NotTo(HaveOccurred())
				Expect(mOptions).To(HaveLen(1))
				// Mount options should override shared options
				Expect(mOptions["common-option"]).To(Equal("mount-value"))
				Expect(encryptOptions).To(BeEmpty())
			})
		})

		Context("when shared encrypt options exist", func() {
			It("should collect shared encrypt options correctly", func() {
				mount := &datav1alpha1.Mount{
					Name:       "test-mount",
					MountPoint: "s3://test-bucket",
				}
				sharedEncryptOptions := []datav1alpha1.EncryptOption{
					{
						Name: "aws-access-key-id",
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{
								Name: "aws-secret",
								Key:  "access-key",
							},
						},
					},
				}
				sharedOptions := map[string]string{}

				mOptions, encryptOptions, err := engine.generateDatasetMountOptions(mount, sharedEncryptOptions, sharedOptions)

				Expect(err).NotTo(HaveOccurred())
				Expect(mOptions).To(BeEmpty())
				Expect(encryptOptions).To(HaveLen(1))
				Expect(encryptOptions["aws-access-key-id"]).To(Equal("/etc/fluid/secrets/aws-secret/access-key"))
			})
		})

		Context("when mount has encrypt options", func() {
			It("should collect mount encrypt options correctly", func() {
				mount := &datav1alpha1.Mount{
					Name:       "test-mount",
					MountPoint: "s3://test-bucket",
					EncryptOptions: []datav1alpha1.EncryptOption{
						{
							Name: "aws-secret-access-key",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "aws-secret",
									Key:  "secret-key",
								},
							},
						},
					},
				}
				var sharedEncryptOptions []datav1alpha1.EncryptOption
				sharedOptions := map[string]string{}

				mOptions, encryptOptions, err := engine.generateDatasetMountOptions(mount, sharedEncryptOptions, sharedOptions)

				Expect(err).NotTo(HaveOccurred())
				Expect(mOptions).To(BeEmpty())
				Expect(encryptOptions).To(HaveLen(1))
				Expect(encryptOptions["aws-secret-access-key"]).To(Equal("/etc/fluid/secrets/aws-secret/secret-key"))
			})
		})

		Context("when both shared and mount encrypt options exist", func() {
			It("should collect all encrypt options", func() {
				mount := &datav1alpha1.Mount{
					Name:       "test-mount",
					MountPoint: "s3://test-bucket",
					EncryptOptions: []datav1alpha1.EncryptOption{
						{
							Name: "mount-encrypt-option",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "mount-secret",
									Key:  "key1",
								},
							},
						},
					},
				}
				sharedEncryptOptions := []datav1alpha1.EncryptOption{
					{
						Name: "shared-encrypt-option",
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{
								Name: "shared-secret",
								Key:  "key2",
							},
						},
					},
				}
				sharedOptions := map[string]string{}

				mOptions, encryptOptions, err := engine.generateDatasetMountOptions(mount, sharedEncryptOptions, sharedOptions)

				Expect(err).NotTo(HaveOccurred())
				Expect(mOptions).To(BeEmpty())
				Expect(encryptOptions).To(HaveLen(2))
				Expect(encryptOptions["shared-encrypt-option"]).To(Equal("/etc/fluid/secrets/shared-secret/key2"))
				Expect(encryptOptions["mount-encrypt-option"]).To(Equal("/etc/fluid/secrets/mount-secret/key1"))
			})

			It("mount encrypt options should override shared encrypt options with same name", func() {
				mount := &datav1alpha1.Mount{
					Name:       "test-mount",
					MountPoint: "s3://test-bucket",
					EncryptOptions: []datav1alpha1.EncryptOption{
						{
							Name: "common-encrypt-option",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "mount-secret",
									Key:  "mount-key",
								},
							},
						},
					},
				}
				sharedEncryptOptions := []datav1alpha1.EncryptOption{
					{
						Name: "common-encrypt-option",
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{
								Name: "shared-secret",
								Key:  "shared-key",
							},
						},
					},
				}
				sharedOptions := map[string]string{}

				mOptions, encryptOptions, err := engine.generateDatasetMountOptions(mount, sharedEncryptOptions, sharedOptions)

				Expect(err).NotTo(HaveOccurred())
				Expect(mOptions).To(BeEmpty())
				Expect(encryptOptions).To(HaveLen(1))
				// Mount encrypt options should override shared encrypt options
				Expect(encryptOptions["common-encrypt-option"]).To(Equal("/etc/fluid/secrets/mount-secret/mount-key"))
			})
		})

		Context("when encrypt option has empty secret key ref", func() {
			It("should return error when secret name is empty", func() {
				mount := &datav1alpha1.Mount{
					Name:       "test-mount",
					MountPoint: "s3://test-bucket",
					EncryptOptions: []datav1alpha1.EncryptOption{
						{
							Name: "test-option",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "",
									Key:  "some-key",
								},
							},
						},
					},
				}
				var sharedEncryptOptions []datav1alpha1.EncryptOption
				sharedOptions := map[string]string{}

				mOptions, encryptOptions, err := engine.generateDatasetMountOptions(mount, sharedEncryptOptions, sharedOptions)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("empty secretKeyRef name or key"))
				Expect(mOptions).To(BeEmpty())
				Expect(encryptOptions).To(BeEmpty())
			})

			It("should return error when secret key is empty", func() {
				mount := &datav1alpha1.Mount{
					Name:       "test-mount",
					MountPoint: "s3://test-bucket",
					EncryptOptions: []datav1alpha1.EncryptOption{
						{
							Name: "test-option",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "some-secret",
									Key:  "",
								},
							},
						},
					},
				}
				var sharedEncryptOptions []datav1alpha1.EncryptOption
				sharedOptions := map[string]string{}

				mOptions, encryptOptions, err := engine.generateDatasetMountOptions(mount, sharedEncryptOptions, sharedOptions)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("empty secretKeyRef name or key"))
				Expect(mOptions).To(BeEmpty())
				Expect(encryptOptions).To(BeEmpty())
			})
		})

		Context("when complex scenario with all options", func() {
			It("should handle all types of options together", func() {
				mount := &datav1alpha1.Mount{
					Name:       "test-mount",
					MountPoint: "s3://test-bucket",
					Options: map[string]string{
						"mount-opt1": "mount-val1",
						"mount-opt2": "mount-val2",
					},
					EncryptOptions: []datav1alpha1.EncryptOption{
						{
							Name: "mount-encrypt",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "mount-secret",
									Key:  "mount-key",
								},
							},
						},
					},
				}
				sharedEncryptOptions := []datav1alpha1.EncryptOption{
					{
						Name: "shared-encrypt",
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{
								Name: "shared-secret",
								Key:  "shared-key",
							},
						},
					},
				}
				sharedOptions := map[string]string{
					"shared-opt1": "shared-val1",
					"mount-opt1":  "shared-val-override",
				}

				mOptions, encryptOptions, err := engine.generateDatasetMountOptions(mount, sharedEncryptOptions, sharedOptions)

				Expect(err).NotTo(HaveOccurred())
				// Check mount options: 3 total (2 from mount + 1 from shared, but mount-opt1 overridden by mount)
				Expect(mOptions).To(HaveLen(3))
				Expect(mOptions["shared-opt1"]).To(Equal("shared-val1"))
				Expect(mOptions["mount-opt1"]).To(Equal("mount-val1")) // Mount overrides shared
				Expect(mOptions["mount-opt2"]).To(Equal("mount-val2"))

				// Check encrypt options: 2 total
				Expect(encryptOptions).To(HaveLen(2))
				Expect(encryptOptions["shared-encrypt"]).To(Equal("/etc/fluid/secrets/shared-secret/shared-key"))
				Expect(encryptOptions["mount-encrypt"]).To(Equal("/etc/fluid/secrets/mount-secret/mount-key"))
			})
		})

		Context("when multiple encrypt options reference same secret", func() {
			It("should generate correct paths for each option", func() {
				mount := &datav1alpha1.Mount{
					Name:       "test-mount",
					MountPoint: "s3://test-bucket",
					EncryptOptions: []datav1alpha1.EncryptOption{
						{
							Name: "option1",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "same-secret",
									Key:  "key1",
								},
							},
						},
						{
							Name: "option2",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "same-secret",
									Key:  "key2",
								},
							},
						},
					},
				}
				var sharedEncryptOptions []datav1alpha1.EncryptOption
				sharedOptions := map[string]string{}

				mOptions, encryptOptions, err := engine.generateDatasetMountOptions(mount, sharedEncryptOptions, sharedOptions)

				Expect(err).NotTo(HaveOccurred())
				Expect(mOptions).To(BeEmpty())
				Expect(encryptOptions).To(HaveLen(2))
				Expect(encryptOptions["option1"]).To(Equal("/etc/fluid/secrets/same-secret/key1"))
				Expect(encryptOptions["option2"]).To(Equal("/etc/fluid/secrets/same-secret/key2"))
			})
		})
	})
})
