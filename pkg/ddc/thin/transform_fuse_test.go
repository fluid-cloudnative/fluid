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
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/testutil"
	"gopkg.in/yaml.v2"
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
})

func TestThinEngine_parseFromProfileFuse(t1 *testing.T) {
	profile := datav1alpha1.ThinRuntimeProfile{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.ThinRuntimeProfileSpec{
			Fuse: datav1alpha1.ThinFuseSpec{
				Image:           "test",
				ImageTag:        "v1",
				ImagePullPolicy: "Always",
				Env: []corev1.EnvVar{{
					Name:  "a",
					Value: "b",
				}, {
					Name: "b",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "test-cm",
							},
						},
					},
				}},
				NodeSelector: map[string]string{"a": "b"},
				Ports: []corev1.ContainerPort{{
					Name:          "port",
					ContainerPort: 8080,
				}},
				LivenessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: "/healthz",
						},
					},
					InitialDelaySeconds: 1,
					TimeoutSeconds:      1,
					PeriodSeconds:       1,
					SuccessThreshold:    1,
					FailureThreshold:    1,
				},
				ReadinessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: "/healthz",
						},
					},
					InitialDelaySeconds: 1,
					TimeoutSeconds:      1,
					PeriodSeconds:       1,
					SuccessThreshold:    1,
					FailureThreshold:    1,
				},
				NetworkMode: datav1alpha1.HostNetworkMode,
			},
		},
	}
	wantValue := &ThinValue{
		Fuse: Fuse{
			Image:           "test",
			ImageTag:        "v1",
			ImagePullPolicy: "Always",
			HostNetwork:     true,
			Envs: []corev1.EnvVar{{
				Name:  "a",
				Value: "b",
			}, {
				Name: "b",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "test-cm",
						},
					},
				},
			}},
			Resources: common.Resources{
				Requests: map[corev1.ResourceName]string{},
				Limits:   map[corev1.ResourceName]string{},
			},
			NodeSelector: map[string]string{"a": "b"},
			Ports: []corev1.ContainerPort{{
				Name:          "port",
				ContainerPort: 8080,
			}},
			LivenessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/healthz",
					},
				},
				InitialDelaySeconds: 1,
				TimeoutSeconds:      1,
				PeriodSeconds:       1,
				SuccessThreshold:    1,
				FailureThreshold:    1,
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/healthz",
					},
				},
				InitialDelaySeconds: 1,
				TimeoutSeconds:      1,
				PeriodSeconds:       1,
				SuccessThreshold:    1,
				FailureThreshold:    1,
			},
		},
	}
	value := &ThinValue{}
	t1.Run("test", func(t1 *testing.T) {
		t := &ThinEngine{
			Log: fake.NullLogger(),
		}
		t.parseFromProfileFuse(&profile, value)
		if !reflect.DeepEqual(value.Fuse, wantValue.Fuse) {
			t1.Errorf("parseFromProfileFuse() got = %v, want = %v", value, wantValue)
		}
	})
}

func TestThinEngine_transformFuseWithDuplicateOptionKey(t1 *testing.T) {
	profile := &datav1alpha1.ThinRuntimeProfile{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: datav1alpha1.ThinRuntimeProfileSpec{
			FileSystemType: "test",
			Fuse: datav1alpha1.ThinFuseSpec{
				Image:           "test",
				ImageTag:        "v1",
				ImagePullPolicy: "Always",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						// Should be inherited
						corev1.ResourceCPU: resource.MustParse("100m"),
						// Should be overridden
						corev1.ResourceMemory: resource.MustParse("2Gi"),
					},
				},
				Env: []corev1.EnvVar{{
					Name:  "a",
					Value: "b",
				}},
				NodeSelector: map[string]string{"a": "b"},
				Ports: []corev1.ContainerPort{{
					Name:          "port",
					ContainerPort: 8080,
				}},
				NetworkMode: datav1alpha1.HostNetworkMode,
				VolumeMounts: []corev1.VolumeMount{{
					Name:      "a",
					MountPath: "/test",
				}},
			},
			Volumes: []corev1.Volume{{
				Name: "a",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{Path: "/test"},
				},
			}},
		},
	}
	runtime := &datav1alpha1.ThinRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.ThinRuntimeSpec{
			ThinRuntimeProfileName: "test",
			Fuse: datav1alpha1.ThinFuseSpec{
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("1Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("200m"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
				},
				Env: []corev1.EnvVar{{
					Name: "b",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "test-cm"},
						},
					},
				}},
				NodeSelector: map[string]string{"b": "c"},
				VolumeMounts: []corev1.VolumeMount{{
					Name:      "b",
					MountPath: "/b",
				}},
				LivenessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: "/healthz",
						},
					},
					InitialDelaySeconds: 1,
					TimeoutSeconds:      1,
					PeriodSeconds:       1,
					SuccessThreshold:    1,
					FailureThreshold:    1,
				},
				ReadinessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: "/healthz",
						},
					},
					InitialDelaySeconds: 1,
					TimeoutSeconds:      1,
					PeriodSeconds:       1,
					SuccessThreshold:    1,
					FailureThreshold:    1,
				},
			},
			Volumes: []corev1.Volume{{
				Name: "b",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{Path: "/b"},
				},
			}},
		},
	}
	dataset := &datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{
			SharedOptions: map[string]string{
				"a": "c",
			},
			Mounts: []datav1alpha1.Mount{{
				MountPoint: "abc",
				Options:    map[string]string{"a": "b"},
			}},
		},
	}
	wantValue := &ThinValue{
		Fuse: Fuse{
			Enabled:         true,
			Image:           "test",
			ImageTag:        "v1",
			ImagePullPolicy: "Always",
			TargetPath:      "/thin/fluid/test/thin-fuse",
			Resources: common.Resources{
				Requests: map[corev1.ResourceName]string{
					corev1.ResourceCPU:    "100m",
					corev1.ResourceMemory: "1Gi",
				},
				Limits: map[corev1.ResourceName]string{
					corev1.ResourceCPU:    "200m",
					corev1.ResourceMemory: "4Gi",
				},
			},
			HostNetwork: true,
			Envs: []corev1.EnvVar{{
				Name:  "a",
				Value: "b",
			}, {
				Name: "b",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "test-cm",
						},
					},
				},
			}, {
				Name:  common.ThinFusePointEnvKey,
				Value: "/thin/fluid/test/thin-fuse",
			}},
			NodeSelector: map[string]string{"b": "c", "fluid.io/f-fluid-test": "true"},
			Ports: []corev1.ContainerPort{{
				Name:          "port",
				ContainerPort: 8080,
			}},
			LivenessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/healthz",
					},
				},
				InitialDelaySeconds: 1,
				TimeoutSeconds:      1,
				PeriodSeconds:       1,
				SuccessThreshold:    1,
				FailureThreshold:    1,
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/healthz",
					},
				},
				InitialDelaySeconds: 1,
				TimeoutSeconds:      1,
				PeriodSeconds:       1,
				SuccessThreshold:    1,
				FailureThreshold:    1,
			},
			Volumes: []corev1.Volume{{
				Name: "a",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{Path: "/test"},
				},
			}, {
				Name: "b",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{Path: "/b"},
				},
			}},
			VolumeMounts: []corev1.VolumeMount{{
				Name:      "a",
				MountPath: "/test",
			}, {
				Name:      "b",
				MountPath: "/b",
			}},
			// ConfigValue: "{\"/thin/fluid/test/thin-fuse\":\"a=b\"}",
			// MountPath:   "/thin/fluid/test/thin-fuse",
			ConfigValue:   "{\"mounts\":[{\"mountPoint\":\"abc\",\"options\":{\"a\":\"b\"}}],\"targetPath\":\"/thin/fluid/test/thin-fuse\",\"accessModes\":[\"ReadOnlyMany\"]}",
			ConfigStorage: "configmap",
			Lifecycle: &corev1.Lifecycle{
				PreStop: &corev1.LifecycleHandler{
					Exec: &corev1.ExecAction{
						Command: []string{
							"umount",
							"/thin/fluid/test/thin-fuse",
						},
					},
				},
			},
		},
	}
	value := &ThinValue{}
	t1.Run("test", func(t1 *testing.T) {
		runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "thin")
		if err != nil {
			t1.Errorf("fail to create the runtimeInfo with error %v", err)
		}

		t := &ThinEngine{
			Log:         fake.NullLogger(),
			namespace:   "fluid",
			name:        "test",
			runtime:     runtime,
			runtimeInfo: runtimeInfo,
			Client:      fake.NewFakeClientWithScheme(testScheme),
		}
		if err := t.transformFuse(runtime, profile, dataset, value); err != nil {
			t1.Errorf("transformFuse() error = %v", err)
		}

		value.Fuse.Envs = testutil.SortEnvVarByName(value.Fuse.Envs, common.ThinFuseOptionEnvKey)
		if !testutil.DeepEqualIgnoringSliceOrder(t1, value.Fuse, wantValue.Fuse) {
			valueYaml, _ := yaml.Marshal(value.Fuse)
			wantYaml, _ := yaml.Marshal(wantValue.Fuse)
			t1.Errorf("transformFuse() \ngot = %v, \nwant = %v", string(valueYaml), string(wantYaml))
		}
	})
}
