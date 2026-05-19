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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("CacheEngine Transform Master Tests", Label("pkg.ddc.cache.engine.transform_master_test.go"), func() {
	var (
		engine       *CacheEngine
		dataset      *datav1alpha1.Dataset
		runtimeObj   *datav1alpha1.CacheRuntime
		runtimeClass *datav1alpha1.CacheRuntimeClass
		config       *CacheRuntimeComponentCommonConfig
		value        *common.CacheRuntimeValue
	)

	BeforeEach(func() {
		// Create a fake client
		scheme := runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(scheme)

		// Create dataset
		dataset = &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-dataset",
				Namespace: "default",
				UID:       "test-dataset-uid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		}

		// Create runtime with master configuration
		runtimeObj = &datav1alpha1.CacheRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-runtime",
				Namespace: "default",
			},
			Spec: datav1alpha1.CacheRuntimeSpec{
				Master: datav1alpha1.CacheRuntimeMasterSpec{
					RuntimeComponentCommonSpec: datav1alpha1.RuntimeComponentCommonSpec{
						NodeSelector: map[string]string{
							"runtime-master-label": "true",
						},
					},
					Replicas: 3,
				},
			},
		}

		// Create runtime class with master template
		runtimeClass = &datav1alpha1.CacheRuntimeClass{
			FileSystemType: "test-fs",
			Topology: &datav1alpha1.RuntimeTopology{
				Master: &datav1alpha1.RuntimeComponentDefinition{
					WorkloadType: metav1.TypeMeta{
						Kind:       "StatefulSet",
						APIVersion: "apps/v1",
					},
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							NodeSelector: map[string]string{
								"original-selector": "value",
							},
							Containers: []corev1.Container{
								{
									Name:  "master",
									Image: "test-master-image:latest",
								},
							},
						},
					},
					Service: datav1alpha1.RuntimeComponentService{
						ComponentServiceConfig: datav1alpha1.ComponentServiceConfig{
							Headless: &datav1alpha1.HeadlessRuntimeComponentService{},
						},
					},
				},
			},
		}

		// Create config
		config = &CacheRuntimeComponentCommonConfig{
			Owner: &common.OwnerReference{
				APIVersion: "data.fluid.io/v1alpha1",
				Kind:       "CacheRuntime",
				Name:       "test-runtime",
				UID:        "test-uid",
			},
			RuntimeConfigs: &RuntimeConfigVolumeConfig{
				RuntimeConfigVolume: corev1.Volume{
					Name: "runtime-config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "test-runtime-config",
							},
						},
					},
				},
				RuntimeConfigVolumeMount: corev1.VolumeMount{
					Name:      "runtime-config",
					MountPath: "/etc/fluid/config",
				},
				ExtraConfigMapNames: make(map[string]bool),
			},
		}

		// Build fake client with objects
		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(dataset, runtimeObj).
			Build()

		engine = &CacheEngine{
			name:      "test-runtime",
			namespace: "default",
			Client:    fakeClient,
			Log:       ctrl.Log.WithName("test"),
		}

		// Initialize value
		value = &common.CacheRuntimeValue{}
	})

	Describe("transformMaster", func() {
		Context("when transforming master configuration", func() {
			It("should not modify the original runtimeClass PodTemplate", func() {
				// Store original PodTemplate for comparison
				originalPodTemplate := runtimeClass.Topology.Master.Template.DeepCopy()
				originRuntime := runtimeObj.DeepCopy()
				originDataset := dataset.DeepCopy()

				// Call transformMaster
				err := engine.transformMaster(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Verify that master was created
				Expect(value.Master).NotTo(BeNil())
				Expect(value.Master.Enabled).To(BeTrue())

				// Verify that the original runtimeClass template was NOT modified
				Expect(runtimeClass.Topology.Master.Template).To(Equal(*originalPodTemplate))
				Expect(*dataset).To(Equal(*originDataset))
				Expect(*runtimeObj).To(Equal(*originRuntime))
			})

			It("should set correct master component properties", func() {
				err := engine.transformMaster(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Verify basic properties
				Expect(value.Master.Name).To(Equal(GetComponentName("test-runtime", common.ComponentTypeMaster)))
				Expect(value.Master.Namespace).To(Equal("default"))
				Expect(value.Master.Enabled).To(BeTrue())
				Expect(value.Master.ComponentType).To(Equal(common.ComponentTypeMaster))
				Expect(value.Master.Replicas).To(Equal(int32(3)))
				Expect(value.Master.WorkloadType.Kind).To(Equal("StatefulSet"))
			})

			It("should configure headless service when defined", func() {
				err := engine.transformMaster(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Verify service configuration
				Expect(value.Master.Service).NotTo(BeNil())
				Expect(value.Master.Service.Name).To(Equal(GetComponentServiceName("test-runtime", common.ComponentTypeMaster)))
			})

			It("should merge node selectors correctly", func() {
				err := engine.transformMaster(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Verify node selector merging (runtime takes higher priority)
				Expect(value.Master.PodTemplateSpec.Spec.NodeSelector).To(HaveKey("original-selector"))
				Expect(value.Master.PodTemplateSpec.Spec.NodeSelector).To(HaveKey("runtime-master-label"))
			})
		})

		Context("when master is disabled", func() {
			BeforeEach(func() {
				runtimeObj.Spec.Master.Disabled = true
			})

			It("should set master as disabled", func() {
				err := engine.transformMaster(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				Expect(value.Master).NotTo(BeNil())
				Expect(value.Master.Enabled).To(BeFalse())
			})

			It("should not modify runtimeClass when master is disabled", func() {
				originalPodTemplate := runtimeClass.Topology.Master.Template.DeepCopy()

				err := engine.transformMaster(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				Expect(runtimeClass.Topology.Master.Template).To(Equal(*originalPodTemplate))
			})
		})

		Context("when runtimeClass topology or master is nil", func() {
			It("should not panic when topology is nil", func() {
				runtimeClass.Topology = nil

				err := engine.transformMaster(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				Expect(value.Master).NotTo(BeNil())
				Expect(value.Master.Enabled).To(BeFalse())
			})

			It("should not panic when master is nil", func() {
				runtimeClass.Topology.Master = nil

				err := engine.transformMaster(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				Expect(value.Master).NotTo(BeNil())
				Expect(value.Master.Enabled).To(BeFalse())
			})
		})

		Context("when transforming master volumes", func() {
			BeforeEach(func() {
				// Add volumes and volumeMounts to runtime
				runtimeObj.Spec.Volumes = []corev1.Volume{
					{
						Name: "test-volume",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "test-secret",
							},
						},
					},
				}
				runtimeObj.Spec.Master.VolumeMounts = []corev1.VolumeMount{
					{
						Name:      "test-volume",
						MountPath: "/etc/test",
						ReadOnly:  true,
					},
				}
			})

			It("should add runtime config volume and mount", func() {
				err := engine.transformMaster(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Verify runtime-config volume exists
				volumeNames := make([]string, len(value.Master.PodTemplateSpec.Spec.Volumes))
				for i, vol := range value.Master.PodTemplateSpec.Spec.Volumes {
					volumeNames[i] = vol.Name
				}
				Expect(volumeNames).To(ContainElement("runtime-config"))

				// Verify runtime-config volume mount exists
				mountNames := make([]string, len(value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts))
				for i, vm := range value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts {
					mountNames[i] = vm.Name
				}
				Expect(mountNames).To(ContainElement("runtime-config"))
			})

			It("should add runtime spec volumes and volumeMounts", func() {
				err := engine.transformMaster(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Verify test-volume exists
				volumeNames := make([]string, len(value.Master.PodTemplateSpec.Spec.Volumes))
				for i, vol := range value.Master.PodTemplateSpec.Spec.Volumes {
					volumeNames[i] = vol.Name
				}
				Expect(volumeNames).To(ContainElement("test-volume"))

				// Verify test-volume mount exists
				mountNames := make([]string, len(value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts))
				for i, vm := range value.Master.PodTemplateSpec.Spec.Containers[0].VolumeMounts {
					mountNames[i] = vm.Name
				}
				Expect(mountNames).To(ContainElement("test-volume"))
			})

			It("should return error if volumeMount has no corresponding volume", func() {
				// Add a volumeMount without a corresponding volume
				runtimeObj.Spec.Master.VolumeMounts = append(runtimeObj.Spec.Master.VolumeMounts, corev1.VolumeMount{
					Name:      "non-existent-volume",
					MountPath: "/etc/nonexistent",
				})

				err := engine.transformMaster(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("volume not found for volumeMount non-existent-volume"))
			})
		})

		Context("environment variable injection", func() {
			It("should inject FLUID environment variables", func() {
				err := engine.transformMaster(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				envVars := value.Master.PodTemplateSpec.Spec.Containers[0].Env
				envNames := make([]string, len(envVars))
				for i, env := range envVars {
					envNames[i] = env.Name
				}

				// Verify essential FLUID environment variables
				Expect(envNames).To(ContainElement("FLUID_DATASET_NAME"))
				Expect(envNames).To(ContainElement("FLUID_DATASET_NAMESPACE"))
				Expect(envNames).To(ContainElement("FLUID_RUNTIME_CONFIG_PATH"))
				Expect(envNames).To(ContainElement("FLUID_RUNTIME_COMPONENT_TYPE"))
				Expect(envNames).To(ContainElement("FLUID_RUNTIME_COMPONENT_SVC_NAME"))
			})

			It("should set FLUID_RUNTIME_COMPONENT_TYPE to Master", func() {
				err := engine.transformMaster(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				envVars := value.Master.PodTemplateSpec.Spec.Containers[0].Env
				for _, env := range envVars {
					if env.Name == "FLUID_RUNTIME_COMPONENT_TYPE" {
						Expect(env.Value).To(Equal(string(common.ComponentTypeMaster)))
					}
				}
			})
		})

		Context("deep copy verification", func() {
			It("should preserve all original fields after multiple transformations", func() {
				originalPodTemplate := runtimeClass.Topology.Master.Template.DeepCopy()

				// Call transformMaster multiple times
				for i := 0; i < 3; i++ {
					testValue := &common.CacheRuntimeValue{}
					err := engine.transformMaster(dataset, runtimeObj, runtimeClass, config, testValue)
					Expect(err).NotTo(HaveOccurred())
				}

				// Verify that the original runtimeClass template is completely unchanged
				Expect(runtimeClass.Topology.Master.Template).To(Equal(*originalPodTemplate))
			})
		})
	})
})
