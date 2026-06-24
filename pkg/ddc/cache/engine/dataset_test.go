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
	"time"

	"github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// -----------------------------------------------------------------------------
// GetCacheStates tests
// -----------------------------------------------------------------------------

var _ = Describe("GetCacheStates Tests", Label("pkg.ddc.cache.engine.dataset_test.go"), func() {
	var (
		patches *gomonkey.Patches
		engine  *CacheEngine
	)

	BeforeEach(func() {
		engine = &CacheEngine{
			name:      "test-runtime",
			namespace: "default",
			Log:       logr.Discard(),
		}
	})

	AfterEach(func() {
		if patches != nil {
			patches.Reset()
			patches = nil
		}
	})

	buildMasterWorkerRuntimeClass := func(reportSummary *datav1alpha1.ExecutionCommonEntry) *datav1alpha1.CacheRuntimeClass {
		return &datav1alpha1.CacheRuntimeClass{
			ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-class"},
			Topology: &datav1alpha1.RuntimeTopology{
				Master: &datav1alpha1.RuntimeComponentDefinition{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "master"}},
						},
					},
					ExecutionEntries: &datav1alpha1.ExecutionEntries{
						ReportSummary: reportSummary,
					},
				},
			},
		}
	}

	buildWorkersOnlyRuntimeClass := func(reportSummary *datav1alpha1.ExecutionCommonEntry) *datav1alpha1.CacheRuntimeClass {
		return &datav1alpha1.CacheRuntimeClass{
			ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-class-workers-only"},
			Topology: &datav1alpha1.RuntimeTopology{
				Worker: &datav1alpha1.RuntimeComponentDefinition{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "worker"}},
						},
					},
					ExecutionEntries: &datav1alpha1.ExecutionEntries{
						ReportSummary: reportSummary,
					},
				},
			},
		}
	}

	newMasterWorkerRuntime := func() *datav1alpha1.CacheRuntime {
		return &datav1alpha1.CacheRuntime{
			Spec: datav1alpha1.CacheRuntimeSpec{
				RuntimeClassName: "test-runtime-class",
				Master:           datav1alpha1.CacheRuntimeMasterSpec{Replicas: 1},
			},
		}
	}

	newWorkersOnlyRuntime := func() *datav1alpha1.CacheRuntime {
		return &datav1alpha1.CacheRuntime{
			Spec: datav1alpha1.CacheRuntimeSpec{
				RuntimeClassName: "test-runtime-class-workers-only",
				Master: datav1alpha1.CacheRuntimeMasterSpec{
					RuntimeComponentCommonSpec: datav1alpha1.RuntimeComponentCommonSpec{
						Disabled: true,
					},
					Replicas: 0,
				},
			},
		}
	}

	Describe("GetCacheStates with Master-Worker architecture", func() {
		Context("when ReportSummary returns valid JSON", func() {
			It("should parse JSON and return the correct CacheStateList", func() {
				reportSummary := &datav1alpha1.ExecutionCommonEntry{
					Command:        []string{"/bin/sh", "-c", "report-summary"},
					TimeoutSeconds: 30,
				}
				runtimeClass := buildMasterWorkerRuntimeClass(reportSummary)
				runtime := newMasterWorkerRuntime()

				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, command []string, timeout time.Duration) (stdout string, stderr string, err error) {
						return `{"cached":"1048576","cachedPercentage":"50","cacheCapacity":"2097152","cacheHitRatio":"0.85"}`, "", nil
					})

				cacheStates, err := engine.GetCacheStates(runtime, runtimeClass)

				Expect(err).NotTo(HaveOccurred())
				Expect(cacheStates).NotTo(BeNil())
				Expect(cacheStates[common.Cached]).To(Equal("1048576"))
				Expect(cacheStates[common.CachedPercentage]).To(Equal("50"))
				Expect(cacheStates[common.CacheCapacity]).To(Equal("2097152"))
				Expect(cacheStates[common.CacheHitRatio]).To(Equal("0.85"))
			})

			It("should return valid CacheStateList with partial fields when JSON has partial data", func() {
				reportSummary := &datav1alpha1.ExecutionCommonEntry{
					Command:        []string{"/report-summary"},
					TimeoutSeconds: 30,
				}
				runtimeClass := buildMasterWorkerRuntimeClass(reportSummary)
				runtime := newMasterWorkerRuntime()

				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, command []string, timeout time.Duration) (stdout string, stderr string, err error) {
						return `{"cached":"524288","cacheCapacity":"1048576"}`, "", nil
					})

				cacheStates, err := engine.GetCacheStates(runtime, runtimeClass)

				Expect(err).NotTo(HaveOccurred())
				Expect(cacheStates).NotTo(BeNil())
				Expect(cacheStates[common.Cached]).To(Equal("524288"))
				Expect(cacheStates[common.CacheCapacity]).To(Equal("1048576"))
				Expect(cacheStates[common.CachedPercentage]).To(Equal(""))
				Expect(cacheStates[common.CacheHitRatio]).To(Equal(""))
			})

			It("should clamp timeout to MinExecutionTimeoutSeconds when configured timeout is too small", func() {
				reportSummary := &datav1alpha1.ExecutionCommonEntry{
					Command:        []string{"/report-summary"},
					TimeoutSeconds: 5,
				}
				runtimeClass := buildMasterWorkerRuntimeClass(reportSummary)
				runtime := newMasterWorkerRuntime()

				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, command []string, timeout time.Duration) (stdout string, stderr string, err error) {
						Expect(timeout).To(Equal(time.Duration(common.MinExecutionTimeoutSeconds) * time.Second))
						return `{"cached":"100","cachedPercentage":"10","cacheCapacity":"1000","cacheHitRatio":"0.5"}`, "", nil
					})

				cacheStates, err := engine.GetCacheStates(runtime, runtimeClass)

				Expect(err).NotTo(HaveOccurred())
				Expect(cacheStates).NotTo(BeNil())
			})

			It("should respect timeout that is larger than MinExecutionTimeoutSeconds", func() {
				reportSummary := &datav1alpha1.ExecutionCommonEntry{
					Command:        []string{"/report-summary"},
					TimeoutSeconds: 60,
				}
				runtimeClass := buildMasterWorkerRuntimeClass(reportSummary)
				runtime := newMasterWorkerRuntime()

				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, command []string, timeout time.Duration) (stdout string, stderr string, err error) {
						Expect(timeout).To(Equal(60 * time.Second))
						return `{"cached":"100","cachedPercentage":"10","cacheCapacity":"1000","cacheHitRatio":"0.5"}`, "", nil
					})

				cacheStates, err := engine.GetCacheStates(runtime, runtimeClass)

				Expect(err).NotTo(HaveOccurred())
				Expect(cacheStates).NotTo(BeNil())
			})
		})

		Context("when ReportSummary returns malformed JSON", func() {
			It("should return an error when stdout is not valid JSON", func() {
				reportSummary := &datav1alpha1.ExecutionCommonEntry{
					Command:        []string{"/report-summary"},
					TimeoutSeconds: 30,
				}
				runtimeClass := buildMasterWorkerRuntimeClass(reportSummary)
				runtime := newMasterWorkerRuntime()

				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, command []string, timeout time.Duration) (stdout string, stderr string, err error) {
						return "this is not json {invalid}", "", nil
					})

				cacheStates, err := engine.GetCacheStates(runtime, runtimeClass)

				Expect(err).To(HaveOccurred())
				Expect(cacheStates).To(BeNil())
			})

			It("should return an error when stdout is empty", func() {
				reportSummary := &datav1alpha1.ExecutionCommonEntry{
					Command:        []string{"/report-summary"},
					TimeoutSeconds: 30,
				}
				runtimeClass := buildMasterWorkerRuntimeClass(reportSummary)
				runtime := newMasterWorkerRuntime()

				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, command []string, timeout time.Duration) (stdout string, stderr string, err error) {
						return "", "", nil
					})

				cacheStates, err := engine.GetCacheStates(runtime, runtimeClass)

				Expect(err).To(HaveOccurred())
				Expect(cacheStates).To(BeNil())
			})
		})

		Context("when pod exec fails", func() {
			It("should return an error when ExecCommandInContainerWithTimeout fails", func() {
				reportSummary := &datav1alpha1.ExecutionCommonEntry{
					Command:        []string{"/report-summary"},
					TimeoutSeconds: 30,
				}
				runtimeClass := buildMasterWorkerRuntimeClass(reportSummary)
				runtime := newMasterWorkerRuntime()

				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, command []string, timeout time.Duration) (stdout string, stderr string, err error) {
						return "", "pod not running", errors.New("pod exec failed: pod not found")
					})

				cacheStates, err := engine.GetCacheStates(runtime, runtimeClass)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("error when executing command"))
				Expect(cacheStates).To(BeNil())
			})

			It("should return an error when exec command times out", func() {
				reportSummary := &datav1alpha1.ExecutionCommonEntry{
					Command:        []string{"/report-summary"},
					TimeoutSeconds: 30,
				}
				runtimeClass := buildMasterWorkerRuntimeClass(reportSummary)
				runtime := newMasterWorkerRuntime()

				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, command []string, timeout time.Duration) (stdout string, stderr string, err error) {
						return "", "", errors.New("command timed out")
					})

				cacheStates, err := engine.GetCacheStates(runtime, runtimeClass)

				Expect(err).To(HaveOccurred())
				Expect(cacheStates).To(BeNil())
			})
		})

		Context("when ReportSummary entry is not configured", func() {
			It("should return an error when ReportSummary is nil in ExecutionEntries", func() {
				runtimeClass := &datav1alpha1.CacheRuntimeClass{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-class-no-report"},
					Topology: &datav1alpha1.RuntimeTopology{
						Master: &datav1alpha1.RuntimeComponentDefinition{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{{Name: "master"}},
								},
							},
							ExecutionEntries: &datav1alpha1.ExecutionEntries{
								ReportSummary: nil,
							},
						},
					},
				}
				runtime := &datav1alpha1.CacheRuntime{
					Spec: datav1alpha1.CacheRuntimeSpec{
						RuntimeClassName: "test-runtime-class-no-report",
						Master:           datav1alpha1.CacheRuntimeMasterSpec{Replicas: 1},
					},
				}

				cacheStates, err := engine.GetCacheStates(runtime, runtimeClass)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("ReportSummary command is empty or not configured"))
				Expect(cacheStates).To(BeNil())
			})

			It("should return an error when ExecutionEntries itself is nil", func() {
				runtimeClass := &datav1alpha1.CacheRuntimeClass{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-class-no-exec-entries"},
					Topology: &datav1alpha1.RuntimeTopology{
						Master: &datav1alpha1.RuntimeComponentDefinition{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{{Name: "master"}},
								},
							},
							ExecutionEntries: nil,
						},
					},
				}
				runtime := &datav1alpha1.CacheRuntime{
					Spec: datav1alpha1.CacheRuntimeSpec{
						RuntimeClassName: "test-runtime-class-no-exec-entries",
						Master:           datav1alpha1.CacheRuntimeMasterSpec{Replicas: 1},
					},
				}

				cacheStates, err := engine.GetCacheStates(runtime, runtimeClass)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("ReportSummary command is empty or not configured"))
				Expect(cacheStates).To(BeNil())
			})
		})

		Context("when pod info resolution fails", func() {
			It("should return an error when master component has no containers", func() {
				runtimeClass := &datav1alpha1.CacheRuntimeClass{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-class-no-containers"},
					Topology: &datav1alpha1.RuntimeTopology{
						Master: &datav1alpha1.RuntimeComponentDefinition{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{},
								},
							},
							ExecutionEntries: &datav1alpha1.ExecutionEntries{
								ReportSummary: &datav1alpha1.ExecutionCommonEntry{
									Command: []string{"/report-summary"},
								},
							},
						},
					},
				}
				runtime := &datav1alpha1.CacheRuntime{
					Spec: datav1alpha1.CacheRuntimeSpec{
						RuntimeClassName: "test-runtime-class-no-containers",
						Master:           datav1alpha1.CacheRuntimeMasterSpec{Replicas: 1},
					},
				}

				cacheStates, err := engine.GetCacheStates(runtime, runtimeClass)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no container in master pod template"))
				Expect(cacheStates).To(BeNil())
			})
		})
	})

	Describe("GetCacheStates with Workers-Only architecture", func() {
		Context("when ReportSummary returns valid JSON through worker pod", func() {
			It("should parse JSON and return correct CacheStateList via workers-only handler", func() {
				reportSummary := &datav1alpha1.ExecutionCommonEntry{
					Command:        []string{"/report-summary"},
					TimeoutSeconds: 30,
				}
				runtimeClass := buildWorkersOnlyRuntimeClass(reportSummary)
				runtime := newWorkersOnlyRuntime()

				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, command []string, timeout time.Duration) (stdout string, stderr string, err error) {
						Expect(podName).To(Equal(common.GetCacheComponentName("test-runtime", common.ComponentTypeWorker) + "-0"))
						Expect(containerName).To(Equal("worker"))
						return `{"cached":"5242880","cachedPercentage":"25","cacheCapacity":"20971520","cacheHitRatio":"0.75"}`, "", nil
					})

				cacheStates, err := engine.GetCacheStates(runtime, runtimeClass)

				Expect(err).NotTo(HaveOccurred())
				Expect(cacheStates).NotTo(BeNil())
				Expect(cacheStates[common.Cached]).To(Equal("5242880"))
				Expect(cacheStates[common.CachedPercentage]).To(Equal("25"))
				Expect(cacheStates[common.CacheCapacity]).To(Equal("20971520"))
				Expect(cacheStates[common.CacheHitRatio]).To(Equal("0.75"))
			})
		})

		Context("when worker component pod info resolution fails", func() {
			It("should return an error when worker component has no containers", func() {
				runtimeClass := &datav1alpha1.CacheRuntimeClass{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-class-worker-no-containers"},
					Topology: &datav1alpha1.RuntimeTopology{
						Worker: &datav1alpha1.RuntimeComponentDefinition{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{},
								},
							},
							ExecutionEntries: &datav1alpha1.ExecutionEntries{
								ReportSummary: &datav1alpha1.ExecutionCommonEntry{
									Command: []string{"/report-summary"},
								},
							},
						},
					},
				}
				runtime := newWorkersOnlyRuntime()

				cacheStates, err := engine.GetCacheStates(runtime, runtimeClass)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no container in worker pod template"))
				Expect(cacheStates).To(BeNil())
			})
		})

		Context("when runtimeClass topology is nil (workers-only default)", func() {
			It("should return an error when topology is nil", func() {
				runtimeClass := &datav1alpha1.CacheRuntimeClass{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-class-no-topology"},
				}
				runtime := &datav1alpha1.CacheRuntime{
					Spec: datav1alpha1.CacheRuntimeSpec{
						RuntimeClassName: "test-runtime-class-no-topology",
						Master:           datav1alpha1.CacheRuntimeMasterSpec{Replicas: 1},
					},
				}

				cacheStates, err := engine.GetCacheStates(runtime, runtimeClass)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("ReportSummary command is empty or not configured"))
				Expect(cacheStates).To(BeNil())
			})

			It("should return an error when runtimeClass itself is nil", func() {
				runtime := &datav1alpha1.CacheRuntime{
					Spec: datav1alpha1.CacheRuntimeSpec{
						RuntimeClassName: "test-runtime-class-nil",
						Master:           datav1alpha1.CacheRuntimeMasterSpec{Replicas: 1},
					},
				}

				cacheStates, err := engine.GetCacheStates(runtime, nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("ReportSummary command is empty or not configured"))
				Expect(cacheStates).To(BeNil())
			})
		})

		Context("when ReportSummary exec fails on workers-only architecture", func() {
			It("should propagate the exec error from worker pod", func() {
				reportSummary := &datav1alpha1.ExecutionCommonEntry{
					Command:        []string{"/report-summary"},
					TimeoutSeconds: 30,
				}
				runtimeClass := buildWorkersOnlyRuntimeClass(reportSummary)
				runtime := newWorkersOnlyRuntime()

				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, command []string, timeout time.Duration) (stdout string, stderr string, err error) {
						return "", "worker pod unavailable", errors.New("connection refused")
					})

				cacheStates, err := engine.GetCacheStates(runtime, runtimeClass)

				Expect(err).To(HaveOccurred())
				Expect(cacheStates).To(BeNil())
			})
		})

		Context("when malformed JSON returned from worker pod", func() {
			It("should return JSON unmarshal error", func() {
				reportSummary := &datav1alpha1.ExecutionCommonEntry{
					Command:        []string{"/report-summary"},
					TimeoutSeconds: 30,
				}
				runtimeClass := buildWorkersOnlyRuntimeClass(reportSummary)
				runtime := newWorkersOnlyRuntime()

				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, command []string, timeout time.Duration) (stdout string, stderr string, err error) {
						return "{invalid json}", "", nil
					})

				cacheStates, err := engine.GetCacheStates(runtime, runtimeClass)

				Expect(err).To(HaveOccurred())
				Expect(cacheStates).To(BeNil())
			})
		})
	})
})