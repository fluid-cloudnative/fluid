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

var _ = Describe("CacheEngine Transform Client Tests", Label("pkg.ddc.cache.engine.transform_client_test.go"), func() {
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

		// Create dataset with encrypt options
		dataset = &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-dataset",
				Namespace: "default",
				UID:       "test-dataset-uid",
			},
			Spec: datav1alpha1.DatasetSpec{
				SharedEncryptOptions: []datav1alpha1.EncryptOption{
					{
						Name: "test-secret",
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{
								Name: "test-secret-name",
								Key:  "key",
							},
						},
					},
				},
			},
		}

		// Create runtime with client configuration
		runtimeObj = &datav1alpha1.CacheRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-runtime",
				Namespace: "default",
			},
			Spec: datav1alpha1.CacheRuntimeSpec{
				Client: datav1alpha1.CacheRuntimeClientSpec{
					RuntimeComponentCommonSpec: datav1alpha1.RuntimeComponentCommonSpec{
						NodeSelector: map[string]string{
							"runtime-client-label": "true",
						},
					},
				},
			},
		}

		// Create runtime class with client template (DaemonSet)
		runtimeClass = &datav1alpha1.CacheRuntimeClass{
			FileSystemType: "test-fs",
			Topology: &datav1alpha1.RuntimeTopology{
				Client: &datav1alpha1.RuntimeComponentDefinition{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							NodeSelector: map[string]string{
								"original-selector": "value",
							},
							Containers: []corev1.Container{
								{
									Name:  "client",
									Image: "test-client-image:latest",
								},
							},
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

	Describe("transformClient", func() {
		Context("when transforming client configuration", func() {
			It("should not modify the original runtimeClass PodTemplate", func() {
				// Store original PodTemplate for comparison
				originalPodTemplate := runtimeClass.Topology.Client.Template.DeepCopy()
				originRuntime := runtimeObj.DeepCopy()
				originDataset := dataset.DeepCopy()

				// Call transformClient
				err := engine.transformClient(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Verify that client was created
				Expect(value.Client).NotTo(BeNil())
				Expect(value.Client.Enabled).To(BeTrue())

				// Verify that the original runtimeClass template was NOT modified
				Expect(runtimeClass.Topology.Client.Template).To(Equal(*originalPodTemplate))
				Expect(*dataset).To(Equal(*originDataset))
				Expect(*runtimeObj).To(Equal(*originRuntime))
			})

			It("should set correct client component properties", func() {
				err := engine.transformClient(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Verify basic properties
				Expect(value.Client.Name).To(Equal(GetComponentName("test-runtime", common.ComponentTypeClient)))
				Expect(value.Client.Namespace).To(Equal("default"))
				Expect(value.Client.Enabled).To(BeTrue())
				Expect(value.Client.ComponentType).To(Equal(common.ComponentTypeClient))
				Expect(value.Client.Replicas).To(Equal(int32(1))) // Client always has 1 replica
			})

			It("should merge node selectors correctly", func() {
				err := engine.transformClient(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Verify node selector merging (runtime takes higher priority)
				Expect(value.Client.PodTemplateSpec.Spec.NodeSelector).To(HaveKey("original-selector"))
				Expect(value.Client.PodTemplateSpec.Spec.NodeSelector).To(HaveKey("runtime-client-label"))
			})
		})

		Context("when client is disabled", func() {
			BeforeEach(func() {
				runtimeObj.Spec.Client.Disabled = true
			})

			It("should set client as disabled", func() {
				err := engine.transformClient(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				Expect(value.Client).NotTo(BeNil())
				Expect(value.Client.Enabled).To(BeFalse())
			})

			It("should not modify runtimeClass when client is disabled", func() {
				originalPodTemplate := runtimeClass.Topology.Client.Template.DeepCopy()

				err := engine.transformClient(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				Expect(runtimeClass.Topology.Client.Template).To(Equal(*originalPodTemplate))
			})
		})

		Context("when runtimeClass topology or client is nil", func() {
			It("should not panic when topology is nil", func() {
				runtimeClass.Topology = nil

				err := engine.transformClient(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				Expect(value.Client).NotTo(BeNil())
				Expect(value.Client.Enabled).To(BeFalse())
			})

			It("should not panic when client is nil", func() {
				runtimeClass.Topology.Client = nil

				err := engine.transformClient(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				Expect(value.Client).NotTo(BeNil())
				Expect(value.Client.Enabled).To(BeFalse())
			})
		})

		Context("secret mount behavior (Client default disabled)", func() {
			It("should NOT mount secrets by default for Client", func() {
				err := engine.transformClient(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Client should NOT have secret volumes by default (defaultMountSecrets=false)
				volumeNames := make([]string, len(value.Client.PodTemplateSpec.Spec.Volumes))
				for i, vol := range value.Client.PodTemplateSpec.Spec.Volumes {
					volumeNames[i] = vol.Name
				}

				// Should only have runtime-config, not secret volumes
				Expect(volumeNames).To(ContainElement("runtime-config"))
				Expect(volumeNames).NotTo(ContainElement(ContainSubstring("secret")))
			})

			It("should mount secrets when SecretMount is explicitly enabled", func() {
				// Enable SecretMount for client
				runtimeClass.Topology.Client.Dependencies.SecretMount = &datav1alpha1.SecretMountComponentDependency{
					Enabled: true,
				}

				err := engine.transformClient(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Now client SHOULD have secret volumes
				volumeNames := make([]string, len(value.Client.PodTemplateSpec.Spec.Volumes))
				for i, vol := range value.Client.PodTemplateSpec.Spec.Volumes {
					volumeNames[i] = vol.Name
				}

				// Should have both runtime-config and secret volume
				Expect(volumeNames).To(ContainElement("runtime-config"))
				Expect(volumeNames).To(ContainElement(ContainSubstring("secret")))
			})

			It("should NOT mount secrets when SecretMount is explicitly disabled", func() {
				// Explicitly disable SecretMount for client
				runtimeClass.Topology.Client.Dependencies.SecretMount = &datav1alpha1.SecretMountComponentDependency{
					Enabled: false,
				}

				err := engine.transformClient(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Client should NOT have secret volumes
				volumeNames := make([]string, len(value.Client.PodTemplateSpec.Spec.Volumes))
				for i, vol := range value.Client.PodTemplateSpec.Spec.Volumes {
					volumeNames[i] = vol.Name
				}

				Expect(volumeNames).To(ContainElement("runtime-config"))
				Expect(volumeNames).NotTo(ContainElement(ContainSubstring("secret")))
			})
		})

		Context("environment variable injection", func() {
			It("should inject FLUID environment variables", func() {
				err := engine.transformClient(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				envVars := value.Client.PodTemplateSpec.Spec.Containers[0].Env
				envNames := make([]string, len(envVars))
				for i, env := range envVars {
					envNames[i] = env.Name
				}

				// Verify essential FLUID environment variables
				Expect(envNames).To(ContainElement("FLUID_DATASET_NAME"))
				Expect(envNames).To(ContainElement("FLUID_DATASET_NAMESPACE"))
				Expect(envNames).To(ContainElement("FLUID_RUNTIME_CONFIG_PATH"))
				Expect(envNames).To(ContainElement("FLUID_RUNTIME_COMPONENT_TYPE"))
			})

			It("should set FLUID_RUNTIME_COMPONENT_TYPE to Client", func() {
				err := engine.transformClient(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				envVars := value.Client.PodTemplateSpec.Spec.Containers[0].Env
				for _, env := range envVars {
					if env.Name == "FLUID_RUNTIME_COMPONENT_TYPE" {
						Expect(env.Value).To(Equal(string(common.ComponentTypeClient)))
					}
				}
			})
		})

		Context("volume transformation", func() {
			BeforeEach(func() {
				// Add volumes and volumeMounts to runtime
				runtimeObj.Spec.Volumes = []corev1.Volume{
					{
						Name: "test-volume",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "test-configmap",
								},
							},
						},
					},
				}
				runtimeObj.Spec.Client.VolumeMounts = []corev1.VolumeMount{
					{
						Name:      "test-volume",
						MountPath: "/etc/test",
						ReadOnly:  true,
					},
				}
			})

			It("should add runtime spec volumes and volumeMounts", func() {
				err := engine.transformClient(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Verify test-volume exists
				volumeNames := make([]string, len(value.Client.PodTemplateSpec.Spec.Volumes))
				for i, vol := range value.Client.PodTemplateSpec.Spec.Volumes {
					volumeNames[i] = vol.Name
				}
				Expect(volumeNames).To(ContainElement("test-volume"))

				// Verify test-volume mount exists
				mountNames := make([]string, len(value.Client.PodTemplateSpec.Spec.Containers[0].VolumeMounts))
				for i, vm := range value.Client.PodTemplateSpec.Spec.Containers[0].VolumeMounts {
					mountNames[i] = vm.Name
				}
				Expect(mountNames).To(ContainElement("test-volume"))
			})
		})

		Context("deep copy verification", func() {
			It("should preserve all original fields after multiple transformations", func() {
				originalPodTemplate := runtimeClass.Topology.Client.Template.DeepCopy()

				// Call transformClient multiple times
				for i := 0; i < 3; i++ {
					testValue := &common.CacheRuntimeValue{}
					err := engine.transformClient(dataset, runtimeObj, runtimeClass, config, testValue)
					Expect(err).NotTo(HaveOccurred())
				}

				// Verify that the original runtimeClass template is completely unchanged
				Expect(runtimeClass.Topology.Client.Template).To(Equal(*originalPodTemplate))
			})
		})
	})
})
