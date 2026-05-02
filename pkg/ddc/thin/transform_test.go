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
	data1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ThinEngine transform tests", Label("pkg.ddc.thin.transform_test.go"), func() {
	var (
		dataset     *data1alpha1.Dataset
		thinruntime *data1alpha1.ThinRuntime
		profile     *data1alpha1.ThinRuntimeProfile
		engine      *ThinEngine
		resources   []runtime.Object
	)

	BeforeEach(func() {
		dataset, thinruntime, profile = mockFluidObjectsForTests(types.NamespacedName{Name: "test-dataset", Namespace: "default"})
		engine = mockThinEngineForTests(dataset, thinruntime, profile)
		resources = []runtime.Object{dataset, thinruntime, profile}
	})

	JustBeforeEach(func() {
		engine.Client = fake.NewFakeClientWithScheme(data1alpha1.UnitTestScheme, resources...)
	})

	Describe("transform", func() {
		It("returns an error when runtime is nil", func() {
			value, err := engine.transform(nil, profile)

			Expect(err).To(MatchError("the thinRuntime is null"))
			Expect(value).To(BeNil())
		})

		It("returns a dataset lookup error when dataset is missing", func() {
			engine.Client = fake.NewFakeClientWithScheme(data1alpha1.UnitTestScheme, thinruntime, profile)

			value, err := engine.transform(thinruntime, profile)

			Expect(apierrors.IsNotFound(err)).To(BeTrue())
			Expect(value).To(BeNil())
		})

		When("worker is disabled", func() {
			BeforeEach(func() {
				thinruntime.Spec.Replicas = 1
				dataset.Spec.PlacementMode = ""
				dataset.Spec.Tolerations = []corev1.Toleration{{
					Key:      "dedicated",
					Operator: corev1.TolerationOpEqual,
					Value:    "thin",
				}}
				profile.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{Name: "profile-secret"}}
			})

			It("uses profile defaults on the happy path", func() {
				value, err := engine.transform(thinruntime, profile)

				Expect(err).NotTo(HaveOccurred())
				Expect(value.ImagePullSecrets).To(Equal(profile.Spec.ImagePullSecrets))
				Expect(value.Worker.Enabled).To(BeFalse())
				Expect(value.Tolerations).To(Equal(dataset.Spec.Tolerations))
				Expect(value.PlacementMode).To(Equal(string(data1alpha1.ExclusiveMode)))
				Expect(value.OwnerDatasetId).To(Equal(dataset.Labels[common.LabelAnnotationDatasetId]))
				Expect(value.RuntimeIdentity).To(Equal(common.RuntimeIdentity{
					Namespace: thinruntime.Namespace,
					Name:      thinruntime.Name,
				}))
				Expect(value.Owner).NotTo(BeNil())
				Expect(value.Owner.Name).To(Equal(thinruntime.Name))
			})
		})

		When("worker is enabled", func() {
			BeforeEach(func() {
				thinruntime.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{Name: "runtime-secret"}}
				thinruntime.Spec.TieredStore.Levels = []data1alpha1.Level{{Path: "/runtime/cache"}}
				thinruntime.Spec.Worker = data1alpha1.ThinCompTemplateSpec{
					Enabled:          true,
					Image:            "runtime-worker",
					ImageTag:         "runtime-tag",
					ImagePullPolicy:  string(corev1.PullIfNotPresent),
					ImagePullSecrets: []corev1.LocalObjectReference{{Name: "runtime-worker-secret"}},
					Env:              []corev1.EnvVar{{Name: "RUNTIME_ENV", Value: "runtime"}},
					Ports:            []corev1.ContainerPort{{Name: "http", ContainerPort: 8080}},
					NodeSelector:     map[string]string{"node": "runtime"},
					NetworkMode:      data1alpha1.HostNetworkMode,
					VolumeMounts: []corev1.VolumeMount{{
						Name:      "runtime-volume",
						MountPath: "/runtime/mount",
					}},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")},
						Limits:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("2Gi")},
					},
				}
				thinruntime.Spec.Volumes = []corev1.Volume{{
					Name: "runtime-volume",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				}}
				profile.Spec.Worker = data1alpha1.ThinCompTemplateSpec{
					Image:            "profile-worker",
					ImageTag:         "profile-tag",
					ImagePullPolicy:  string(corev1.PullAlways),
					ImagePullSecrets: []corev1.LocalObjectReference{{Name: "profile-worker-secret"}},
					Env:              []corev1.EnvVar{{Name: "PROFILE_ENV", Value: "profile"}},
					Ports:            []corev1.ContainerPort{{Name: "metrics", ContainerPort: 9090}},
					NodeSelector:     map[string]string{"node": "profile"},
					VolumeMounts: []corev1.VolumeMount{{
						Name:      "profile-volume",
						MountPath: "/profile/mount",
					}},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
						Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")},
					},
				}
				profile.Spec.Volumes = []corev1.Volume{{
					Name: "profile-volume",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				}}
			})

			It("uses runtime image pull secrets and transforms worker fields", func() {
				value, err := engine.transform(thinruntime, profile)

				Expect(err).NotTo(HaveOccurred())
				Expect(value.ImagePullSecrets).To(Equal(thinruntime.Spec.ImagePullSecrets))
				Expect(value.Worker.Image).To(Equal(thinruntime.Spec.Worker.Image))
				Expect(value.Worker.ImageTag).To(Equal(thinruntime.Spec.Worker.ImageTag))
				Expect(value.Worker.ImagePullPolicy).To(Equal(thinruntime.Spec.Worker.ImagePullPolicy))
				Expect(value.Worker.ImagePullSecrets).To(Equal(thinruntime.Spec.Worker.ImagePullSecrets))
				Expect(value.Worker.Envs).To(ContainElements(profile.Spec.Worker.Env[0], thinruntime.Spec.Worker.Env[0]))
				Expect(value.Worker.Ports).To(ContainElements(profile.Spec.Worker.Ports[0], thinruntime.Spec.Worker.Ports[0]))
				Expect(value.Worker.NodeSelector).To(Equal(thinruntime.Spec.Worker.NodeSelector))
				Expect(value.Worker.CacheDir).To(Equal("/runtime/cache"))
				Expect(value.Worker.HostNetwork).To(BeTrue())
				Expect(value.Worker.VolumeMounts).To(ContainElements(profile.Spec.Worker.VolumeMounts[0], thinruntime.Spec.Worker.VolumeMounts[0]))
				Expect(value.Worker.Volumes).To(ContainElement(profile.Spec.Volumes[0]))
				Expect(value.Worker.Volumes).To(ContainElement(thinruntime.Spec.Volumes[0]))
				Expect(value.Worker.Resources.Requests).To(HaveKeyWithValue(corev1.ResourceCPU, "1"))
				Expect(value.Worker.Resources.Requests).To(HaveKeyWithValue(corev1.ResourceMemory, "1Gi"))
				Expect(value.Worker.Resources.Limits).To(HaveKeyWithValue(corev1.ResourceCPU, "500m"))
				Expect(value.Worker.Resources.Limits).To(HaveKeyWithValue(corev1.ResourceMemory, "2Gi"))
			})
		})

		It("handles a missing profile when callers continue after profile lookup not found", func() {
			thinruntime.Spec.Fuse.Image = "runtime-fuse"
			thinruntime.Spec.Fuse.ImageTag = "runtime-tag"

			value, err := engine.transform(thinruntime, nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(value).NotTo(BeNil())
			Expect(value.ImagePullSecrets).To(BeNil())
		})
	})

	Describe("transformWorkers", func() {
		var value *ThinValue

		BeforeEach(func() {
			value = &ThinValue{}
			profile.Spec.Worker = data1alpha1.ThinCompTemplateSpec{
				Image:            "profile-worker",
				ImageTag:         "profile-tag",
				ImagePullPolicy:  string(corev1.PullAlways),
				ImagePullSecrets: []corev1.LocalObjectReference{{Name: "profile-worker-secret"}},
				Env:              []corev1.EnvVar{{Name: "PROFILE_ENV", Value: "profile"}},
				Ports:            []corev1.ContainerPort{{Name: "profile-port", ContainerPort: 9090}},
				NodeSelector:     map[string]string{"profile": "true"},
				NetworkMode:      data1alpha1.ContainerNetworkMode,
				VolumeMounts: []corev1.VolumeMount{{
					Name:      "profile-volume",
					MountPath: "/profile/mount",
				}},
			}
			profile.Spec.Volumes = []corev1.Volume{{
				Name: "profile-volume",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			}}
			thinruntime.Spec.Worker = data1alpha1.ThinCompTemplateSpec{
				ImageTag:         "runtime-tag",
				ImagePullSecrets: []corev1.LocalObjectReference{{Name: "runtime-worker-secret"}},
				Env:              []corev1.EnvVar{{Name: "RUNTIME_ENV", Value: "runtime"}},
				Ports:            []corev1.ContainerPort{{Name: "runtime-port", ContainerPort: 8080}},
				NodeSelector:     map[string]string{"runtime": "true"},
				NetworkMode:      data1alpha1.HostNetworkMode,
				VolumeMounts: []corev1.VolumeMount{{
					Name:      "runtime-volume",
					MountPath: "/runtime/mount",
				}},
			}
			thinruntime.Spec.Volumes = []corev1.Volume{{
				Name: "runtime-volume",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			}}
			thinruntime.Spec.TieredStore.Levels = []data1alpha1.Level{{Path: "/runtime/cache"}}
		})

		It("applies profile defaults and runtime overrides", func() {
			err := engine.transformWorkers(thinruntime, profile, value)

			Expect(err).NotTo(HaveOccurred())
			Expect(value.Worker.Image).To(Equal(profile.Spec.Worker.Image))
			Expect(value.Worker.ImageTag).To(Equal(thinruntime.Spec.Worker.ImageTag))
			Expect(value.Worker.ImagePullPolicy).To(Equal(profile.Spec.Worker.ImagePullPolicy))
			Expect(value.Worker.ImagePullSecrets).To(Equal(thinruntime.Spec.Worker.ImagePullSecrets))
			Expect(value.Worker.Envs).To(ContainElements(profile.Spec.Worker.Env[0], thinruntime.Spec.Worker.Env[0]))
			Expect(value.Worker.Ports).To(ContainElements(profile.Spec.Worker.Ports[0], thinruntime.Spec.Worker.Ports[0]))
			Expect(value.Worker.NodeSelector).To(Equal(thinruntime.Spec.Worker.NodeSelector))
			Expect(value.Worker.VolumeMounts).To(ContainElements(profile.Spec.Worker.VolumeMounts[0], thinruntime.Spec.Worker.VolumeMounts[0]))
			Expect(value.Worker.Volumes).To(ContainElements(profile.Spec.Volumes[0], thinruntime.Spec.Volumes[0]))
			Expect(value.Worker.CacheDir).To(Equal("/runtime/cache"))
			Expect(value.Worker.HostNetwork).To(BeTrue())
		})

		It("swallows worker volume transformation errors", func() {
			profile.Spec.Worker.VolumeMounts = []corev1.VolumeMount{{Name: "missing-volume", MountPath: "/profile/mount"}}
			profile.Spec.Volumes = nil

			err := engine.transformWorkers(thinruntime, profile, value)

			Expect(err).NotTo(HaveOccurred())
			Expect(value.Worker.VolumeMounts).To(ContainElement(thinruntime.Spec.Worker.VolumeMounts[0]))
			Expect(value.Worker.Volumes).To(ContainElement(thinruntime.Spec.Volumes[0]))
		})
	})

	Describe("parseFromProfile", func() {
		var value *ThinValue

		BeforeEach(func() {
			value = &ThinValue{}
		})

		It("is a no-op for a nil profile", func() {
			engine.parseFromProfile(nil, value)

			Expect(value).To(Equal(&ThinValue{}))
		})

		It("copies worker settings from the profile", func() {
			profile.Spec.Worker = data1alpha1.ThinCompTemplateSpec{
				Image:            "test",
				ImageTag:         "v1",
				ImagePullPolicy:  string(corev1.PullAlways),
				ImagePullSecrets: []corev1.LocalObjectReference{{Name: "pull-secret"}},
				Env: []corev1.EnvVar{{
					Name:  "a",
					Value: "b",
				}, {
					Name: "b",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "test-cm"},
						},
					},
				}},
				NodeSelector: map[string]string{"a": "b"},
				Ports: []corev1.ContainerPort{{
					Name:          "port",
					ContainerPort: 8080,
				}},
				LivenessProbe: &corev1.Probe{
					ProbeHandler:        corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{Path: "/healthz"}},
					InitialDelaySeconds: 1,
					TimeoutSeconds:      1,
					PeriodSeconds:       1,
					SuccessThreshold:    1,
					FailureThreshold:    1,
				},
				ReadinessProbe: &corev1.Probe{
					ProbeHandler:        corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{Path: "/healthz"}},
					InitialDelaySeconds: 1,
					TimeoutSeconds:      1,
					PeriodSeconds:       1,
					SuccessThreshold:    1,
					FailureThreshold:    1,
				},
				NetworkMode: data1alpha1.HostNetworkMode,
			}

			engine.parseFromProfile(profile, value)

			Expect(value.Worker.Image).To(Equal("test"))
			Expect(value.Worker.ImageTag).To(Equal("v1"))
			Expect(value.Worker.ImagePullPolicy).To(Equal(string(corev1.PullAlways)))
			Expect(value.Worker.ImagePullSecrets).To(Equal([]corev1.LocalObjectReference{{Name: "pull-secret"}}))
			Expect(value.Worker.Resources).To(Equal(common.Resources{
				Requests: map[corev1.ResourceName]string{},
				Limits:   map[corev1.ResourceName]string{},
			}))
			Expect(value.Worker.HostNetwork).To(BeTrue())
			Expect(value.Worker.Envs).To(Equal(profile.Spec.Worker.Env))
			Expect(value.Worker.NodeSelector).To(Equal(profile.Spec.Worker.NodeSelector))
			Expect(value.Worker.Ports).To(Equal(profile.Spec.Worker.Ports))
			Expect(value.Worker.LivenessProbe).To(Equal(profile.Spec.Worker.LivenessProbe))
			Expect(value.Worker.ReadinessProbe).To(Equal(profile.Spec.Worker.ReadinessProbe))
		})
	})

	Describe("parseWorkerImage", func() {
		It("selectively overrides only populated runtime image fields", func() {
			value := &ThinValue{Worker: Worker{
				Image:            "profile-image",
				ImageTag:         "profile-tag",
				ImagePullPolicy:  string(corev1.PullAlways),
				ImagePullSecrets: []corev1.LocalObjectReference{{Name: "profile-secret"}},
			}}
			thinruntime.Spec.Worker = data1alpha1.ThinCompTemplateSpec{
				ImageTag:         "runtime-tag",
				ImagePullSecrets: []corev1.LocalObjectReference{{Name: "runtime-secret"}},
			}

			engine.parseWorkerImage(thinruntime, value)

			Expect(value.Worker.Image).To(Equal("profile-image"))
			Expect(value.Worker.ImageTag).To(Equal("runtime-tag"))
			Expect(value.Worker.ImagePullPolicy).To(Equal(string(corev1.PullAlways)))
			Expect(value.Worker.ImagePullSecrets).To(Equal([]corev1.LocalObjectReference{{Name: "runtime-secret"}}))
		})
	})

	Describe("transformPlacementMode", func() {
		It("defaults to exclusive mode when dataset placement mode is empty", func() {
			value := &ThinValue{}

			engine.transformPlacementMode(&data1alpha1.Dataset{}, value)

			Expect(value.PlacementMode).To(Equal(string(data1alpha1.ExclusiveMode)))
		})

		It("keeps an existing dataset placement mode", func() {
			value := &ThinValue{}

			engine.transformPlacementMode(&data1alpha1.Dataset{Spec: data1alpha1.DatasetSpec{PlacementMode: data1alpha1.ShareMode}}, value)

			Expect(value.PlacementMode).To(Equal(string(data1alpha1.ShareMode)))
		})
	})

	Describe("transformTolerations", func() {
		It("copies dataset tolerations into values", func() {
			value := &ThinValue{}
			tolerations := []corev1.Toleration{{
				Key:      "a",
				Operator: corev1.TolerationOpEqual,
				Value:    "b",
			}}

			engine.transformTolerations(&data1alpha1.Dataset{Spec: data1alpha1.DatasetSpec{Tolerations: tolerations}}, value)

			Expect(value.Tolerations).To(Equal(tolerations))
		})
	})
})
