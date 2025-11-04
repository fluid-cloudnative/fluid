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

package thin

import (
	"encoding/json"
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ThinEngine FUSE Transformation Tests", Label("pkg.ddc.thin.transform_fuse_test.go"), func() {
	var (
		dataset     *datav1alpha1.Dataset
		thinruntime *datav1alpha1.ThinRuntime
		profile     *datav1alpha1.ThinRuntimeProfile
		engine      *ThinEngine
		client      client.Client
		resources   []runtime.Object

		value *ThinValue
	)

	BeforeEach(func() {
		dataset, thinruntime, profile = mockFluidObjectsForTests(types.NamespacedName{Name: "test-dataset", Namespace: "default"})
		engine = mockThinEngineForTests(dataset, thinruntime, profile)
		resources = []runtime.Object{dataset, thinruntime, profile}
		value = &ThinValue{}
	})

	JustBeforeEach(func() {
		client = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
		engine.Client = client
	})

	Describe("Test ThinEngine.transformFuse", func() {
		BeforeEach(func() {
			profile.Spec.Volumes = []corev1.Volume{
				{
					Name: "myvol",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{
							Medium: corev1.StorageMediumMemory,
						},
					},
				},
			}
			profile.Spec.Fuse = datav1alpha1.ThinFuseSpec{
				Image:            "myimage",
				ImageTag:         "v1.0.0",
				ImagePullPolicy:  "Always",
				ImagePullSecrets: []corev1.LocalObjectReference{{Name: "my-secret"}},
				PodMetadata: datav1alpha1.PodMetadata{
					Labels:      map[string]string{"label1": "value-profile", "test1": "myvalue"},
					Annotations: map[string]string{"annotation1": "value-profile", "test2": "myvalue"},
				},
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("1"),
					},
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("2Gi"),
					},
				},
				Env: []corev1.EnvVar{
					{Name: "MY_ENV", Value: "MY_VALUE"},
				},
				Command: []string{"mycommand"},
				Args:    []string{"myarg1", "myarg2"},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "myvol",
						MountPath: "/path/to/myvol",
					},
				},
				Lifecycle: &corev1.Lifecycle{
					PreStop: &corev1.LifecycleHandler{
						Exec: &corev1.ExecAction{
							Command: []string{"prestopcommand"},
						},
					},
				},
			}
		})
		When("no extra specification in thinruntime", func() {
			It("should honor specifications in thinruntime profile", func() {
				err := engine.transformFuse(thinruntime, profile, dataset, value)
				Expect(err).To(BeNil())
				Expect(value.Fuse.Image).To(Equal(profile.Spec.Fuse.Image))
				Expect(value.Fuse.ImageTag).To(Equal(profile.Spec.Fuse.ImageTag))
				Expect(value.Fuse.ImagePullPolicy).To(Equal(profile.Spec.Fuse.ImagePullPolicy))
				Expect(value.Fuse.ImagePullSecrets).To(Equal(profile.Spec.Fuse.ImagePullSecrets))
				Expect(value.Fuse.Envs).To(ContainElement(profile.Spec.Fuse.Env[0]))
				Expect(value.Fuse.Command).To(Equal(profile.Spec.Fuse.Command))
				Expect(value.Fuse.Args).To(Equal(profile.Spec.Fuse.Args))
				Expect(value.Fuse.Lifecycle.PreStop).To(Equal(profile.Spec.Fuse.Lifecycle.PreStop))
				Expect(value.Fuse.VolumeMounts).To(ContainElement(profile.Spec.Fuse.VolumeMounts[0]))

				// check pod metadata
				Expect(value.Fuse.Labels).To(HaveKeyWithValue("label1", "value-profile"))
				Expect(value.Fuse.Labels).To(HaveKeyWithValue("test1", "myvalue"))
				Expect(value.Fuse.Annotations).To(HaveKeyWithValue("annotation1", "value-profile"))
				Expect(value.Fuse.Annotations).To(HaveKeyWithValue("test2", "myvalue"))

				// check resources
				Expect(value.Fuse.Resources.Requests).To(HaveKeyWithValue(corev1.ResourceMemory, "2Gi"))
				Expect(value.Fuse.Resources.Limits).To(HaveKeyWithValue(corev1.ResourceCPU, "1"))
			})
		})

		When("thinruntime overrides specification defined in thinruntime profile", func() {
			BeforeEach(func() {
				thinruntime.Spec.Fuse = datav1alpha1.ThinFuseSpec{
					Image:            "newimage",
					ImageTag:         "v1.0.1",
					ImagePullPolicy:  "IfNotPresent",
					ImagePullSecrets: []corev1.LocalObjectReference{{Name: "my-new-secret"}},
					PodMetadata: datav1alpha1.PodMetadata{
						Labels:      map[string]string{"label1": "value-runtime", "xxx": "yyy"},
						Annotations: map[string]string{"annotation1": "value-runtime", "yyy": "zzz"},
					},
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							// corev1.ResourceCPU:    resource.MustParse("2"),
							corev1.ResourceMemory: resource.MustParse("4Gi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("4Gi"),
						},
					},
					Env: []corev1.EnvVar{
						{Name: "MY_ENV2", Value: "MY_VALUE2"},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "myvol2",
							MountPath: "/path/to/myvol2",
						},
					},
				}
				thinruntime.Spec.Volumes = []corev1.Volume{
					{
						Name: "myvol2",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{
								Medium: corev1.StorageMediumMemory,
							},
						},
					},
				}
			})

			It("should honor specification in thinruntime", func() {
				err := engine.transformFuse(thinruntime, profile, dataset, value)
				Expect(err).To(BeNil())
				// specifications to be overridden
				Expect(value.Fuse.Image).To(Equal(thinruntime.Spec.Fuse.Image))
				Expect(value.Fuse.ImageTag).To(Equal(thinruntime.Spec.Fuse.ImageTag))
				Expect(value.Fuse.ImagePullPolicy).To(Equal(thinruntime.Spec.Fuse.ImagePullPolicy))
				Expect(value.Fuse.ImagePullSecrets).To(Equal(thinruntime.Spec.Fuse.ImagePullSecrets))

				// specifications to be inherited
				Expect(value.Fuse.Command).To(Equal(profile.Spec.Fuse.Command))
				Expect(value.Fuse.Args).To(Equal(profile.Spec.Fuse.Args))
				Expect(value.Fuse.Lifecycle.PreStop).To(Equal(profile.Spec.Fuse.Lifecycle.PreStop))

				// specifications to be merged
				Expect(value.Fuse.Envs).To(ContainElements(profile.Spec.Fuse.Env[0], thinruntime.Spec.Fuse.Env[0]))
				Expect(value.Fuse.VolumeMounts).To(ContainElements(profile.Spec.Fuse.VolumeMounts[0], thinruntime.Spec.Fuse.VolumeMounts[0]))
				Expect(value.Fuse.Labels).To(Equal(map[string]string{"label1": "value-runtime", "xxx": "yyy", "test1": "myvalue"}))
				Expect(value.Fuse.Annotations).To(Equal(map[string]string{"annotation1": "value-runtime", "yyy": "zzz", "test2": "myvalue"}))

				// check resources
				Expect(value.Fuse.Resources.Limits).To(HaveKeyWithValue(corev1.ResourceCPU, "1"))
				Expect(value.Fuse.Resources.Limits).To(HaveKeyWithValue(corev1.ResourceMemory, "4Gi"))
				Expect(value.Fuse.Resources.Requests).To(HaveKeyWithValue(corev1.ResourceCPU, "1"))
				Expect(value.Fuse.Resources.Requests).To(HaveKeyWithValue(corev1.ResourceMemory, "4Gi"))
			})
		})

		When("parsing network mode", func() {
			It("should default to use host network", func() {
				thinruntime.Spec.Fuse.NetworkMode = datav1alpha1.DefaultNetworkMode
				profile.Spec.Fuse.NetworkMode = datav1alpha1.DefaultNetworkMode
				err := engine.transformFuse(thinruntime, profile, dataset, value)
				Expect(err).NotTo(HaveOccurred())

				Expect(value.Fuse.HostNetwork).To(BeTrue())
			})

			It("should use host network when thinruntime specifies host network", func() {
				thinruntime.Spec.Fuse.NetworkMode = datav1alpha1.HostNetworkMode
				profile.Spec.Fuse.NetworkMode = datav1alpha1.ContainerNetworkMode
				err := engine.transformFuse(thinruntime, profile, dataset, value)
				Expect(err).NotTo(HaveOccurred())

				Expect(value.Fuse.HostNetwork).To(BeTrue())
			})

			It("should use container network when thinruntime specifies container network", func() {
				thinruntime.Spec.Fuse.NetworkMode = datav1alpha1.ContainerNetworkMode
				profile.Spec.Fuse.NetworkMode = datav1alpha1.HostNetworkMode
				err := engine.transformFuse(thinruntime, profile, dataset, value)
				Expect(err).NotTo(HaveOccurred())

				Expect(value.Fuse.HostNetwork).To(BeFalse())
			})

			It("should follow network mode in profile when thinruntime does not set network mode", func() {
				thinruntime.Spec.Fuse.NetworkMode = datav1alpha1.DefaultNetworkMode
				profile.Spec.Fuse.NetworkMode = datav1alpha1.ContainerNetworkMode
				err := engine.transformFuse(thinruntime, profile, dataset, value)
				Expect(err).NotTo(HaveOccurred())

				Expect(value.Fuse.HostNetwork).To(BeFalse())
			})
		})
	})

	Describe("Test ThinEngine.parseFuseImage", func() {
		When("image info set in thinruntime", func() {
			BeforeEach(func() {
				thinruntime.Spec.Fuse.Image = "myimage"
				thinruntime.Spec.Fuse.ImageTag = "latest"
				thinruntime.Spec.Fuse.ImagePullPolicy = "Always"
				thinruntime.Spec.Fuse.ImagePullSecrets = []corev1.LocalObjectReference{{Name: "my-secret"}}
			})

			It("should honor image and image tag in thinruntime", func() {
				engine.parseFuseImage(thinruntime, value)
				Expect(value.Fuse.Image).To(Equal("myimage"))
				Expect(value.Fuse.ImageTag).To(Equal("latest"))
				Expect(value.Fuse.ImagePullPolicy).To(Equal("Always"))
				Expect(value.Fuse.ImagePullSecrets).To(Equal([]corev1.LocalObjectReference{{Name: "my-secret"}}))
			})
		})

		When("image info is not set in thinruntime", func() {
			BeforeEach(func() {
				value.Fuse.Image = "image-from-profile"
				value.Fuse.ImageTag = "tag-from-profile"
			})
			It("should not change existing image info in values", func() {
				engine.parseFuseImage(thinruntime, value)
				Expect(value.Fuse.Image).To(Equal("image-from-profile"))
				Expect(value.Fuse.ImageTag).To(Equal("tag-from-profile"))
			})
		})
	})

	Describe("Test ThinEngine.parseFuseOptions", func() {
		BeforeEach(func() {
			profile.Spec.Fuse.Options = map[string]string{
				"option1": "value1",
				"option2": "value2",
			}
		})

		When("options are defined in both runtime and profile", func() {
			It("should parse options from runtime and profile correctly", func() {
				thinruntime.Spec.Fuse.Options = map[string]string{
					"option2": "overridden-value2",
					"option3": "value3",
				}

				options, err := engine.parseFuseOptions(thinruntime, profile, dataset)
				Expect(err).NotTo(HaveOccurred())
				Expect(options).To(ContainSubstring("option1=value1"))
				Expect(options).To(ContainSubstring("option2=overridden-value2"))
				Expect(options).To(ContainSubstring("option3=value3"))
			})
		})

		When("no options are defined", func() {
			It("should handle empty options correctly", func() {
				thinruntime.Spec.Fuse.Options = map[string]string{}
				profile.Spec.Fuse.Options = nil

				options, err := engine.parseFuseOptions(thinruntime, profile, dataset)
				Expect(err).NotTo(HaveOccurred())
				Expect(options).To(Equal("ro"))
			})
		})
	})

	Describe("Test ThinEngine.parseLifecycle", func() {
		When("no custom lifecycle is defined", func() {
			It("should set default lifecycle with umount command", func() {
				err := engine.parseLifecycle(thinruntime, profile, value)
				Expect(err).NotTo(HaveOccurred())
				Expect(value.Fuse.Lifecycle).NotTo(BeNil())
				Expect(value.Fuse.Lifecycle.PreStop).NotTo(BeNil())
				Expect(value.Fuse.Lifecycle.PreStop.Exec).NotTo(BeNil())
				Expect(value.Fuse.Lifecycle.PreStop.Exec.Command).To(ContainElement("umount"))
				Expect(value.Fuse.Lifecycle.PreStop.Exec.Command).To(ContainElement("/thin/default/test-dataset/thin-fuse"))
			})
		})

		When("custom lifecycle is defined in profile", func() {
			It("should use custom lifecycle from profile", func() {
				profile.Spec.Fuse.Lifecycle = &corev1.Lifecycle{
					PreStop: &corev1.LifecycleHandler{
						Exec: &corev1.ExecAction{
							Command: []string{"custom-command"},
						},
					},
				}

				err := engine.parseLifecycle(thinruntime, profile, value)
				Expect(err).NotTo(HaveOccurred())
				Expect(value.Fuse.Lifecycle.PreStop.Exec.Command).To(ContainElement("custom-command"))
			})
		})

		When("custom lifecycle is defined in both profile and runtime", func() {
			It("should use custom lifecycle from runtime which overrides profile", func() {
				profile.Spec.Fuse.Lifecycle = &corev1.Lifecycle{
					PreStop: &corev1.LifecycleHandler{
						Exec: &corev1.ExecAction{
							Command: []string{"profile-command"},
						},
					},
				}

				thinruntime.Spec.Fuse.Lifecycle = &corev1.Lifecycle{
					PreStop: &corev1.LifecycleHandler{
						Exec: &corev1.ExecAction{
							Command: []string{"runtime-command"},
						},
					},
				}

				err := engine.parseLifecycle(thinruntime, profile, value)
				Expect(err).NotTo(HaveOccurred())
				Expect(value.Fuse.Lifecycle.PreStop.Exec.Command).To(ContainElement("runtime-command"))
				Expect(value.Fuse.Lifecycle.PreStop.Exec.Command).NotTo(ContainElement("profile-command"))
			})
		})

		When("postStart is set in profile", func() {
			It("should return error when postStart is set in profile", func() {
				profile.Spec.Fuse.Lifecycle = &corev1.Lifecycle{
					PostStart: &corev1.LifecycleHandler{
						Exec: &corev1.ExecAction{
							Command: []string{"post-start"},
						},
					},
				}

				err := engine.parseLifecycle(thinruntime, profile, value)
				Expect(err).To(HaveOccurred())
			})
		})

		When("postStart is set in runtime", func() {
			It("should return error when postStart is set in runtime", func() {
				thinruntime.Spec.Fuse.Lifecycle = &corev1.Lifecycle{
					PostStart: &corev1.LifecycleHandler{
						Exec: &corev1.ExecAction{
							Command: []string{"post-start"},
						},
					},
				}

				err := engine.parseLifecycle(thinruntime, profile, value)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Test ThinEngine.parseHostVolumeFromDataset", func() {
		var (
			dataset *datav1alpha1.Dataset
			value   *ThinValue
		)

		BeforeEach(func() {
			dataset = &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataset",
					Namespace: "default",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							Options: map[string]string{
								common.DatasetOptionFluidFuseHostVolume: "/host/path:/container/path",
							},
						},
					},
				},
			}
			value = &ThinValue{}
		})

		When("a valid host volume is specified", func() {
			It("should correctly parse host volume configuration", func() {
				err := engine.parseHostVolumeFromDataset(dataset, value)
				Expect(err).NotTo(HaveOccurred())
				Expect(value.Fuse.Volumes).To(HaveLen(1))
				Expect(value.Fuse.VolumeMounts).To(HaveLen(1))
				Expect(value.Fuse.Volumes[0].HostPath.Path).To(Equal("/host/path"))
				Expect(value.Fuse.VolumeMounts[0].MountPath).To(Equal("/container/path"))
			})
		})

		When("multiple host volumes are specified", func() {
			It("should handle multiple host volumes", func() {
				dataset.Spec.Mounts = append(dataset.Spec.Mounts, datav1alpha1.Mount{
					Options: map[string]string{
						common.DatasetOptionFluidFuseHostVolume: "/host/path2:/container/path2",
					},
				})

				err := engine.parseHostVolumeFromDataset(dataset, value)
				Expect(err).NotTo(HaveOccurred())
				Expect(value.Fuse.Volumes).To(HaveLen(2))
				Expect(value.Fuse.VolumeMounts).To(HaveLen(2))
			})
		})

		When("host volume format is invalid", func() {
			It("should return error for invalid host volume format", func() {
				dataset.Spec.Mounts[0].Options[common.DatasetOptionFluidFuseHostVolume] = "/host/path:/container/path:/invalid"

				err := engine.parseHostVolumeFromDataset(dataset, value)
				Expect(err).To(HaveOccurred())
			})
		})

		When("host path is not absolute", func() {
			It("should return error for non-absolute host path", func() {
				dataset.Spec.Mounts[0].Options[common.DatasetOptionFluidFuseHostVolume] = "host/path:/container/path"

				err := engine.parseHostVolumeFromDataset(dataset, value)
				Expect(err).To(HaveOccurred())
			})
		})

		When("mount path is not absolute", func() {
			It("should return error for non-absolute mount path", func() {
				dataset.Spec.Mounts[0].Options[common.DatasetOptionFluidFuseHostVolume] = "/host/path:container/path"

				err := engine.parseHostVolumeFromDataset(dataset, value)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Test ThinEngine.transformFuseConfig", func() {
		When("dataset has mounts and runtime has fuse options", func() {
			BeforeEach(func() {
				thinruntime.Spec.Fuse.Options = map[string]string{
					"runtimeOption1": "runtimeValue1",
				}
			})

			It("should generate correct config value and storage type", func() {
				err := engine.transformFuseConfig(thinruntime, dataset, value)
				Expect(err).To(BeNil())
				Expect(value.Fuse.ConfigStorage).To(Equal("configmap"))
				Expect(value.Fuse.ConfigValue).NotTo(BeEmpty())

				// Verify the structure of ConfigValue
				config := &Config{}
				err = json.Unmarshal([]byte(value.Fuse.ConfigValue), config)
				Expect(err).To(BeNil())
				Expect(config.TargetPath).To(Equal("/thin/default/test-dataset/thin-fuse"))
				Expect(config.Mounts).To(HaveLen(1))
				Expect(config.Mounts[0].MountPoint).To(Equal("s3://mybucket/mypath"))
				Expect(config.Mounts[0].Options).To(HaveKeyWithValue("endpoint", dataset.Spec.Mounts[0].Options["endpoint"]))
				Expect(config.RuntimeOptions).To(HaveKeyWithValue("runtimeOption1", "runtimeValue1"))
				Expect(config.AccessModes).To(ContainElement(corev1.ReadOnlyMany))

				// Verify secret related fields
				secretName := "my-s3-secret"
				Expect(config.Mounts[0].Options).To(HaveKeyWithValue("access-key-id", fmt.Sprintf("/etc/fluid/secrets/%s/myak", secretName)))
				Expect(config.Mounts[0].Options).To(HaveKeyWithValue("access-key-secret", fmt.Sprintf("/etc/fluid/secrets/%s/mysk", secretName)))
				Expect(value.Fuse.VolumeMounts).To(ContainElement(corev1.VolumeMount{
					Name:      fmt.Sprintf("thin-fuseconfig-%s", secretName),
					MountPath: fmt.Sprintf("/etc/fluid/secrets/%s", secretName),
					ReadOnly:  true,
				}))
				Expect(value.Fuse.Volumes).To(ContainElement(corev1.Volume{
					Name: fmt.Sprintf("thin-fuseconfig-%s", secretName),
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: secretName,
						},
					},
				}))
			})
		})

		When("dataset has both shared options and mount options", func() {
			BeforeEach(func() {
				dataset.Spec.SharedOptions = map[string]string{
					"option1":    "value-to-be-overwritten",
					"shared-opt": "shared-value",
				}
				dataset.Spec.Mounts = []datav1alpha1.Mount{
					{
						MountPoint: "local:///mnt/test",
						Options: map[string]string{
							"option1": "value1",
							"option2": "value2",
						},
					},
				}
			})
			It("should merge shared options and mount options", func() {
				err := engine.transformFuseConfig(thinruntime, dataset, value)
				Expect(err).To(BeNil())
				Expect(value.Fuse.ConfigStorage).To(Equal("configmap"))
				Expect(value.Fuse.ConfigValue).NotTo(BeEmpty())

				// Verify the structure of ConfigValue
				config := &Config{}
				err = json.Unmarshal([]byte(value.Fuse.ConfigValue), config)
				Expect(err).To(BeNil())
				Expect(config.Mounts[0].Options).To(HaveKeyWithValue("shared-opt", "shared-value"))
				Expect(config.Mounts[0].Options).To(HaveKeyWithValue("option1", "value1"))
				Expect(config.Mounts[0].Options).To(HaveKeyWithValue("option2", "value2"))
			})
		})

		When("dataset has no mounts", func() {
			BeforeEach(func() {
				dataset.Spec.Mounts = []datav1alpha1.Mount{}
				dataset.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany}
			})

			It("should handle empty mounts and use dataset access modes", func() {
				err := engine.transformFuseConfig(thinruntime, dataset, value)
				Expect(err).To(BeNil())
				Expect(value.Fuse.ConfigValue).NotTo(BeEmpty())

				config := &Config{}
				err = json.Unmarshal([]byte(value.Fuse.ConfigValue), config)
				Expect(err).To(BeNil())
				Expect(config.Mounts).To(HaveLen(0))
				Expect(config.AccessModes).To(ContainElement(corev1.ReadWriteMany))
			})
		})

		When("dataset has no access modes", func() {
			BeforeEach(func() {
				dataset.Spec.Mounts = []datav1alpha1.Mount{
					{
						MountPoint: "local:///mnt/test",
					},
				}
				dataset.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{}
			})

			It("should set default ReadOnlyMany access mode", func() {
				err := engine.transformFuseConfig(thinruntime, dataset, value)
				Expect(err).To(BeNil())

				config := &Config{}
				err = json.Unmarshal([]byte(value.Fuse.ConfigValue), config)
				Expect(err).To(BeNil())
				Expect(config.AccessModes).To(ContainElement(corev1.ReadOnlyMany))
			})
		})
	})
})
