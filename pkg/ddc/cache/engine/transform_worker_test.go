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

var _ = Describe("CacheEngine Transform Worker Tests", Label("pkg.ddc.cache.engine.transform_worker_test.go"), func() {
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

		// Create dataset with node affinity
		dataset = &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-dataset",
				Namespace: "default",
				UID:       "test-dataset-uid",
			},
			Spec: datav1alpha1.DatasetSpec{
				NodeAffinity: &datav1alpha1.CacheableNodeAffinity{
					Required: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{
										Key:      "dataset-node-label",
										Operator: corev1.NodeSelectorOpIn,
										Values:   []string{"true"},
									},
								},
							},
						},
					},
				},
			},
		}

		// Create runtime
		runtimeObj = &datav1alpha1.CacheRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-runtime",
				Namespace: "default",
			},
			Spec: datav1alpha1.CacheRuntimeSpec{
				Worker: datav1alpha1.CacheRuntimeWorkerSpec{
					RuntimeComponentCommonSpec: datav1alpha1.RuntimeComponentCommonSpec{
						NodeSelector: map[string]string{
							"runtime-worker-label": "true",
						},
					},
				},
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

		// Create runtime class with worker template - CacheRuntimeClass has no Spec field
		runtimeClass = &datav1alpha1.CacheRuntimeClass{
			FileSystemType: "test-fs",
			Topology: &datav1alpha1.RuntimeTopology{
				Worker: &datav1alpha1.RuntimeComponentDefinition{
					WorkloadType: metav1.TypeMeta{
						Kind:       "StatefulSet",
						APIVersion: "apps/v1",
					},
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Affinity: &corev1.Affinity{
								NodeAffinity: &corev1.NodeAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
										NodeSelectorTerms: []corev1.NodeSelectorTerm{
											{
												MatchExpressions: []corev1.NodeSelectorRequirement{
													{
														Key:      "runtime-class-label",
														Operator: corev1.NodeSelectorOpIn,
														Values:   []string{"true"},
													},
												},
											},
										},
									},
								},
							},
							NodeSelector: map[string]string{
								"original-selector": "value",
							},
							Containers: []corev1.Container{
								{
									Name:  "worker",
									Image: "test-image:latest",
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

		// Initialize value
		value = &common.CacheRuntimeValue{}
	})

	Describe("transformWorker", func() {
		Context("when transforming worker configuration", func() {
			It("should not modify the original runtimeClass PodTemplate", func() {
				// Store original PodTemplate for comparison
				originalPodTemplate := runtimeClass.Topology.Worker.Template.DeepCopy()
				originRuntime := runtimeObj.DeepCopy()
				originDataset := dataset.DeepCopy()

				// Call transformWorker
				err := engine.transformWorker(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Verify that worker was created
				Expect(value.Worker).NotTo(BeNil())
				Expect(value.Worker.Enabled).To(BeTrue())

				// Verify that the original runtimeClass template was NOT modified by direct comparison
				Expect(runtimeClass.Topology.Worker.Template).To(Equal(*originalPodTemplate))
				Expect(*dataset).To(Equal(*originDataset))
				Expect(*runtimeObj).To(Equal(*originRuntime))
			})

			It("should merge affinities correctly in the worker value without affecting runtimeClass", func() {
				// Call transformWorker
				err := engine.transformWorker(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Verify that the worker value has merged affinities
				Expect(value.Worker.PodTemplateSpec.Spec.Affinity).NotTo(BeNil())
				Expect(value.Worker.PodTemplateSpec.Spec.Affinity.NodeAffinity).NotTo(BeNil())
				Expect(value.Worker.PodTemplateSpec.Spec.Affinity.NodeAffinity.
					RequiredDuringSchedulingIgnoredDuringExecution).NotTo(BeNil())

				// Should have 2 node selector terms (one from runtimeClass, one from dataset)
				terms := value.Worker.PodTemplateSpec.Spec.Affinity.NodeAffinity.
					RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
				Expect(terms).To(HaveLen(2))

				// Verify the original runtimeClass still has only 1 term
				originalTerms := runtimeClass.Topology.Worker.Template.Spec.Affinity.
					NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
				Expect(originalTerms).To(HaveLen(1))
			})

			It("should merge node selectors correctly in the worker value without affecting runtimeClass", func() {
				// Call transformWorker
				err := engine.transformWorker(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Verify that the worker value has merged node selectors
				Expect(value.Worker.PodTemplateSpec.Spec.NodeSelector).To(HaveKey("original-selector"))
				Expect(value.Worker.PodTemplateSpec.Spec.NodeSelector).To(HaveKey("runtime-worker-label"))

				// Verify the original runtimeClass still has only the original selector
				Expect(runtimeClass.Topology.Worker.Template.Spec.NodeSelector).To(HaveKey("original-selector"))
				Expect(runtimeClass.Topology.Worker.Template.Spec.NodeSelector).NotTo(HaveKey("runtime-worker-label"))
			})

			It("should handle nil affinity in runtimeClass template", func() {
				// Set runtimeClass template affinity to nil
				runtimeClass.Topology.Worker.Template.Spec.Affinity = nil

				err := engine.transformWorker(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Verify that the worker value has affinity from dataset
				Expect(value.Worker.PodTemplateSpec.Spec.Affinity).NotTo(BeNil())
				Expect(value.Worker.PodTemplateSpec.Spec.Affinity.NodeAffinity).NotTo(BeNil())

				// Verify the original runtimeClass template is still nil
				Expect(runtimeClass.Topology.Worker.Template.Spec.Affinity).To(BeNil())
			})

			It("should handle nil node affinity in dataset", func() {
				// Set dataset node affinity to nil
				dataset.Spec.NodeAffinity = nil

				err := engine.transformWorker(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Verify that the worker value still has the original affinity from runtimeClass
				Expect(value.Worker.PodTemplateSpec.Spec.Affinity).NotTo(BeNil())
				Expect(value.Worker.PodTemplateSpec.Spec.Affinity.NodeAffinity).NotTo(BeNil())

				// Should have only 1 term (from runtimeClass)
				terms := value.Worker.PodTemplateSpec.Spec.Affinity.NodeAffinity.
					RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
				Expect(terms).To(HaveLen(1))
				Expect(terms[0].MatchExpressions[0].Key).To(Equal("runtime-class-label"))

				// Verify the original runtimeClass is unchanged
				originalTerms := runtimeClass.Topology.Worker.Template.Spec.Affinity.
					NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
				Expect(originalTerms).To(HaveLen(1))
			})

			Context("when transforming worker volumes", func() {
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
						{
							Name: "config-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "test-configmap",
									},
								},
							},
						},
					}
					runtimeObj.Spec.Worker.VolumeMounts = []corev1.VolumeMount{
						{
							Name:      "test-volume",
							MountPath: "/etc/test",
							ReadOnly:  true,
						},
						{
							Name:      "config-volume",
							MountPath: "/etc/config",
						},
					}
				})

				It("should add volumes and volumeMounts to worker PodTemplateSpec", func() {
					err := engine.transformWorker(dataset, runtimeObj, runtimeClass, config, value)
					Expect(err).NotTo(HaveOccurred())

					// Verify that volumes were added (including runtime-config from common config)
					Expect(value.Worker.PodTemplateSpec.Spec.Volumes).To(HaveLen(3))
					volumeNames := make([]string, len(value.Worker.PodTemplateSpec.Spec.Volumes))
					for i, vol := range value.Worker.PodTemplateSpec.Spec.Volumes {
						volumeNames[i] = vol.Name
					}
					Expect(volumeNames).To(ContainElements("runtime-config", "test-volume", "config-volume"))

					// Verify that volumeMounts were added to the first container (including runtime-config)
					Expect(value.Worker.PodTemplateSpec.Spec.Containers[0].VolumeMounts).To(HaveLen(3))
					volumeMountNames := make([]string, len(value.Worker.PodTemplateSpec.Spec.Containers[0].VolumeMounts))
					for i, vm := range value.Worker.PodTemplateSpec.Spec.Containers[0].VolumeMounts {
						volumeMountNames[i] = vm.Name
					}
					Expect(volumeMountNames).To(ContainElements("runtime-config", "test-volume", "config-volume"))
				})

				It("should not add duplicate volumes", func() {
					// Add a volume with the same name to runtimeClass template
					runtimeObj.Spec.Volumes = append(runtimeObj.Spec.Volumes, corev1.Volume{
						Name: "test-volume",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					})

					err := engine.transformWorker(dataset, runtimeObj, runtimeClass, config, value)
					Expect(err).NotTo(HaveOccurred())

					// Should have 3 volumes (runtime-config from common config, test-volume from runtimeClass, config-volume from runtime)
					Expect(value.Worker.PodTemplateSpec.Spec.Volumes).To(HaveLen(3))
				})

				It("should return error if volumeMount has no corresponding volume", func() {
					// Add a volumeMount without a corresponding volume
					runtimeObj.Spec.Worker.VolumeMounts = append(runtimeObj.Spec.Worker.VolumeMounts, corev1.VolumeMount{
						Name:      "non-existent-volume",
						MountPath: "/etc/nonexistent",
					})

					err := engine.transformWorker(dataset, runtimeObj, runtimeClass, config, value)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("volume not found for volumeMount non-existent-volume"))
				})

				It("should not add volumes when worker is disabled", func() {
					runtimeObj.Spec.Worker.Disabled = true

					err := engine.transformWorker(dataset, runtimeObj, runtimeClass, config, value)
					Expect(err).NotTo(HaveOccurred())

					// Worker should be disabled
					Expect(value.Worker.Enabled).To(BeFalse())
					// No volumes should be added
					Expect(value.Worker.PodTemplateSpec.Spec.Volumes).To(BeEmpty())
				})
			})

			It("should set correct pod labels for worker pods to enable PodAntiAffinity", func() {
				err := engine.transformWorker(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Verify pod labels are set in PodTemplateSpec.Labels
				// These labels are used by PodAntiAffinity rules for scheduling isolation
				Expect(value.Worker.PodTemplateSpec.Labels).NotTo(BeNil())
				Expect(value.Worker.PodTemplateSpec.Labels).To(HaveKey(common.LabelAnnotationDataset))
				Expect(value.Worker.PodTemplateSpec.Labels).To(HaveKey(common.LabelAnnotationDatasetPlacement))

				// Verify the dataset label uses human-readable format (namespace-name) instead of UID
				// This is consistent with other runtimes like alluxio, juicefs, etc.
				// Note: In test environment, runtimeInfo is built from runtime name, not dataset name
				Expect(value.Worker.PodTemplateSpec.Labels[common.LabelAnnotationDataset]).To(Equal("default-test-runtime"))
				Expect(value.Worker.PodTemplateSpec.Labels[common.LabelAnnotationDatasetPlacement]).To(Equal(string(datav1alpha1.ExclusiveMode)))
			})

			It("should preserve all original fields in runtimeClass after multiple transformations", func() {
				// Store complete original PodTemplate
				originalPodTemplate := runtimeClass.Topology.Worker.Template.DeepCopy()

				// Call transformWorker multiple times
				for i := 0; i < 3; i++ {
					testValue := &common.CacheRuntimeValue{}
					err := engine.transformWorker(dataset, runtimeObj, runtimeClass, config, testValue)
					Expect(err).NotTo(HaveOccurred())
				}

				// Verify that the original runtimeClass template is completely unchanged
				Expect(runtimeClass.Topology.Worker.Template).To(Equal(*originalPodTemplate))
			})
		})

		Context("when worker is disabled", func() {
			BeforeEach(func() {
				runtimeObj.Spec.Worker.Disabled = true
			})

			It("should not modify runtimeClass when worker is disabled", func() {
				originalPodTemplate := runtimeClass.Topology.Worker.Template.DeepCopy()

				err := engine.transformWorker(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Worker should be disabled
				Expect(value.Worker).NotTo(BeNil())
				Expect(value.Worker.Enabled).To(BeFalse())

				// RuntimeClass should be unchanged
				Expect(runtimeClass.Topology.Worker.Template).To(Equal(*originalPodTemplate))
			})
		})

		Context("when runtimeClass topology or worker is nil", func() {
			It("should not panic when topology is nil", func() {
				runtimeClass.Topology = nil

				err := engine.transformWorker(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Worker should be disabled
				Expect(value.Worker).NotTo(BeNil())
				Expect(value.Worker.Enabled).To(BeFalse())
			})

			It("should not panic when worker is nil", func() {
				runtimeClass.Topology.Worker = nil

				err := engine.transformWorker(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Worker should be disabled
				Expect(value.Worker).NotTo(BeNil())
				Expect(value.Worker.Enabled).To(BeFalse())
			})
		})
	})

	Describe("Worker Affinity Configuration", func() {
		Context("when dataset is in exclusive mode", func() {
			BeforeEach(func() {
				dataset.Spec.PlacementMode = datav1alpha1.ExclusiveMode
			})

			It("should set RequiredDuringScheduling PodAntiAffinity for exclusive mode", func() {
				err := engine.transformWorker(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Verify PodAntiAffinity exists
				Expect(value.Worker.PodTemplateSpec.Spec.Affinity).NotTo(BeNil())
				Expect(value.Worker.PodTemplateSpec.Spec.Affinity.PodAntiAffinity).NotTo(BeNil())

				// In exclusive mode, should have 1 rule in RequiredDuringSchedulingIgnoredDuringExecution
				// (only the dataset exists rule, not the placement rule)
				requiredRules := value.Worker.PodTemplateSpec.Spec.Affinity.PodAntiAffinity.
					RequiredDuringSchedulingIgnoredDuringExecution
				Expect(requiredRules).To(HaveLen(1))

				// Rule: fluid.io/dataset exists
				Expect(requiredRules[0].LabelSelector.MatchExpressions).To(ContainElement(
					metav1.LabelSelectorRequirement{
						Key:      common.LabelAnnotationDataset,
						Operator: metav1.LabelSelectorOpExists,
					},
				))
				Expect(requiredRules[0].TopologyKey).To(Equal(common.K8sNodeNameLabelKey))
			})

			It("should not set PreferredDuringScheduling in exclusive mode", func() {
				err := engine.transformWorker(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// PreferredDuringScheduling should be empty or nil
				preferredRules := value.Worker.PodTemplateSpec.Spec.Affinity.PodAntiAffinity.
					PreferredDuringSchedulingIgnoredDuringExecution
				Expect(preferredRules).To(BeEmpty())
			})
		})

		Context("when dataset is in share mode", func() {
			BeforeEach(func() {
				dataset.Spec.PlacementMode = datav1alpha1.ShareMode
			})

			It("should set PreferredDuringScheduling PodAntiAffinity for share mode", func() {
				err := engine.transformWorker(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Verify PodAntiAffinity exists
				Expect(value.Worker.PodTemplateSpec.Spec.Affinity).NotTo(BeNil())
				Expect(value.Worker.PodTemplateSpec.Spec.Affinity.PodAntiAffinity).NotTo(BeNil())

				// Should have 1 rule in PreferredDuringSchedulingIgnoredDuringExecution with weight 50
				preferredRules := value.Worker.PodTemplateSpec.Spec.Affinity.PodAntiAffinity.
					PreferredDuringSchedulingIgnoredDuringExecution
				Expect(preferredRules).To(HaveLen(1))
				Expect(preferredRules[0].Weight).To(Equal(int32(50)))

				// Verify the label selector
				Expect(preferredRules[0].PodAffinityTerm.LabelSelector.MatchExpressions).To(ContainElement(
					metav1.LabelSelectorRequirement{
						Key:      common.LabelAnnotationDataset,
						Operator: metav1.LabelSelectorOpExists,
					},
				))
				Expect(preferredRules[0].PodAffinityTerm.TopologyKey).To(Equal(common.K8sNodeNameLabelKey))

				// Should also have 1 rule in RequiredDuringScheduling for placement
				requiredRules := value.Worker.PodTemplateSpec.Spec.Affinity.PodAntiAffinity.
					RequiredDuringSchedulingIgnoredDuringExecution
				Expect(requiredRules).To(HaveLen(1))
			})
		})
	})

	Describe("Worker Disabled Scenarios", func() {
		Context("when worker is disabled", func() {
			BeforeEach(func() {
				runtimeObj.Spec.Worker.Disabled = true
			})

			It("should not build affinity when worker is disabled", func() {
				err := engine.transformWorker(dataset, runtimeObj, runtimeClass, config, value)
				Expect(err).NotTo(HaveOccurred())

				// Worker should be disabled
				Expect(value.Worker.Enabled).To(BeFalse())

				// Pod labels should not be set for disabled worker
				Expect(value.Worker.PodTemplateSpec.Labels).To(BeNil())
			})
		})
	})
})
