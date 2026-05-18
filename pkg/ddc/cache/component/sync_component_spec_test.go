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

package component

import (
	"context"

	workloadv1alpha1 "github.com/fluid-cloudnative/advanced-statefulset/api/workload/v1alpha1"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var _ = Describe("AdvancedStatefulSetManager SyncComponentSpec", func() {
	var (
		manager      *AdvancedStatefulSetManager
		ctx          context.Context
		identity     *common.ComponentIdentity
		existingAsts *workloadv1alpha1.AdvancedStatefulSet
	)

	BeforeEach(func() {
		client := setupTestClient()
		manager = newAdvancedStatefulSetManager(client)
		ctx = context.Background()
		ctx = log.IntoContext(ctx, GinkgoLogr)

		identity = &common.ComponentIdentity{
			Name:      "test-runtime-worker",
			Namespace: "fluid",
		}

		// Create existing AdvancedStatefulSet
		replicas := int32(3)
		existingAsts = &workloadv1alpha1.AdvancedStatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      identity.Name,
				Namespace: identity.Namespace,
			},
			Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
				Replicas: &replicas,
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:            "worker",
								Image:           "fluid-cache:v1.0.0",
								ImagePullPolicy: corev1.PullIfNotPresent,
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("2"),
										corev1.ResourceMemory: resource.MustParse("4Gi"),
									},
									Limits: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("4"),
										corev1.ResourceMemory: resource.MustParse("8Gi"),
									},
								},
							},
						},
					},
				},
			},
		}

		err := manager.client.Create(ctx, existingAsts)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("SyncComponentSpec", func() {
		Context("when updating replicas", func() {
			It("should update replicas successfully", func() {
				newReplicas := int32(5)
				spec := ComponentSpec{
					Replicas: &newReplicas,
				}

				err := manager.SyncComponentSpec(ctx, identity, spec)
				Expect(err).NotTo(HaveOccurred())

				// Verify replicas updated
				updatedAsts := &workloadv1alpha1.AdvancedStatefulSet{}
				err = manager.client.Get(ctx, types.NamespacedName{
					Name:      identity.Name,
					Namespace: identity.Namespace,
				}, updatedAsts)
				Expect(err).NotTo(HaveOccurred())
				Expect(*updatedAsts.Spec.Replicas).To(Equal(int32(5)))
			})

			It("should not update when replicas unchanged", func() {
				newReplicas := int32(3) // Same as current
				spec := ComponentSpec{
					Replicas: &newReplicas,
				}

				err := manager.SyncComponentSpec(ctx, identity, spec)
				Expect(err).NotTo(HaveOccurred())

				// Verify no change
				updatedAsts := &workloadv1alpha1.AdvancedStatefulSet{}
				err = manager.client.Get(ctx, types.NamespacedName{
					Name:      identity.Name,
					Namespace: identity.Namespace,
				}, updatedAsts)
				Expect(err).NotTo(HaveOccurred())
				Expect(*updatedAsts.Spec.Replicas).To(Equal(int32(3)))
			})

			It("should skip update when replicas is nil", func() {
				spec := ComponentSpec{
					Replicas: nil,
				}

				err := manager.SyncComponentSpec(ctx, identity, spec)
				Expect(err).NotTo(HaveOccurred())

				// Verify no change
				updatedAsts := &workloadv1alpha1.AdvancedStatefulSet{}
				err = manager.client.Get(ctx, types.NamespacedName{
					Name:      identity.Name,
					Namespace: identity.Namespace,
				}, updatedAsts)
				Expect(err).NotTo(HaveOccurred())
				Expect(*updatedAsts.Spec.Replicas).To(Equal(int32(3)))
			})
		})

		Context("when updating image", func() {
			It("should update full image (name + tag)", func() {
				spec := ComponentSpec{
					Version: datav1alpha1.VersionSpec{
						Image:    "new-fluid-cache",
						ImageTag: "v2.0.0",
					},
				}

				err := manager.SyncComponentSpec(ctx, identity, spec)
				Expect(err).NotTo(HaveOccurred())

				// Verify image updated
				updatedAsts := &workloadv1alpha1.AdvancedStatefulSet{}
				err = manager.client.Get(ctx, types.NamespacedName{
					Name:      identity.Name,
					Namespace: identity.Namespace,
				}, updatedAsts)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedAsts.Spec.Template.Spec.Containers[0].Image).To(Equal("new-fluid-cache:v2.0.0"))
			})

			It("should not update when only imageTag specified", func() {
				spec := ComponentSpec{
					Version: datav1alpha1.VersionSpec{
						ImageTag: "v1.1.0",
					},
				}

				err := manager.SyncComponentSpec(ctx, identity, spec)
				Expect(err).NotTo(HaveOccurred())

				// Verify no change (image unchanged because imageTag alone is invalid)
				updatedAsts := &workloadv1alpha1.AdvancedStatefulSet{}
				err = manager.client.Get(ctx, types.NamespacedName{
					Name:      identity.Name,
					Namespace: identity.Namespace,
				}, updatedAsts)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedAsts.Spec.Template.Spec.Containers[0].Image).To(Equal("fluid-cache:v1.0.0"))
			})

			It("should not update when only image specified", func() {
				spec := ComponentSpec{
					Version: datav1alpha1.VersionSpec{
						Image: "new-image",
					},
				}

				err := manager.SyncComponentSpec(ctx, identity, spec)
				Expect(err).NotTo(HaveOccurred())

				// Verify no change (image unchanged because image alone is invalid)
				updatedAsts := &workloadv1alpha1.AdvancedStatefulSet{}
				err = manager.client.Get(ctx, types.NamespacedName{
					Name:      identity.Name,
					Namespace: identity.Namespace,
				}, updatedAsts)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedAsts.Spec.Template.Spec.Containers[0].Image).To(Equal("fluid-cache:v1.0.0"))
			})

			It("should not update when image unchanged", func() {
				spec := ComponentSpec{
					Version: datav1alpha1.VersionSpec{
						Image:    "fluid-cache",
						ImageTag: "v1.0.0",
					},
				}

				err := manager.SyncComponentSpec(ctx, identity, spec)
				Expect(err).NotTo(HaveOccurred())

				// Verify no change
				updatedAsts := &workloadv1alpha1.AdvancedStatefulSet{}
				err = manager.client.Get(ctx, types.NamespacedName{
					Name:      identity.Name,
					Namespace: identity.Namespace,
				}, updatedAsts)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedAsts.Spec.Template.Spec.Containers[0].Image).To(Equal("fluid-cache:v1.0.0"))
			})
		})

		Context("when updating resources", func() {
			It("should update both requests and limits", func() {
				spec := ComponentSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("4"),
							corev1.ResourceMemory: resource.MustParse("8Gi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("8"),
							corev1.ResourceMemory: resource.MustParse("16Gi"),
						},
					},
				}

				err := manager.SyncComponentSpec(ctx, identity, spec)
				Expect(err).NotTo(HaveOccurred())

				// Verify all resources updated
				updatedAsts := &workloadv1alpha1.AdvancedStatefulSet{}
				err = manager.client.Get(ctx, types.NamespacedName{
					Name:      identity.Name,
					Namespace: identity.Namespace,
				}, updatedAsts)
				Expect(err).NotTo(HaveOccurred())

				container := updatedAsts.Spec.Template.Spec.Containers[0]
				Expect(container.Resources.Requests[corev1.ResourceCPU]).To(Equal(resource.MustParse("4")))
				Expect(container.Resources.Requests[corev1.ResourceMemory]).To(Equal(resource.MustParse("8Gi")))
				Expect(container.Resources.Limits[corev1.ResourceCPU]).To(Equal(resource.MustParse("8")))
				Expect(container.Resources.Limits[corev1.ResourceMemory]).To(Equal(resource.MustParse("16Gi")))
			})

			It("should not update when resources unchanged", func() {
				spec := ComponentSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("2"),
							corev1.ResourceMemory: resource.MustParse("4Gi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("4"),
							corev1.ResourceMemory: resource.MustParse("8Gi"),
						},
					},
				}

				err := manager.SyncComponentSpec(ctx, identity, spec)
				Expect(err).NotTo(HaveOccurred())

				// Verify no change
				updatedAsts := &workloadv1alpha1.AdvancedStatefulSet{}
				err = manager.client.Get(ctx, types.NamespacedName{
					Name:      identity.Name,
					Namespace: identity.Namespace,
				}, updatedAsts)
				Expect(err).NotTo(HaveOccurred())

				container := updatedAsts.Spec.Template.Spec.Containers[0]
				Expect(container.Resources.Requests[corev1.ResourceCPU]).To(Equal(resource.MustParse("2")))
			})
		})

		Context("when updating multiple fields", func() {
			It("should update replicas, image and resources together", func() {
				newReplicas := int32(5)
				spec := ComponentSpec{
					Replicas: &newReplicas,
					Version: datav1alpha1.VersionSpec{
						Image:    "fluid-cache",
						ImageTag: "v1.1.0",
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("4"),
						},
					},
				}

				err := manager.SyncComponentSpec(ctx, identity, spec)
				Expect(err).NotTo(HaveOccurred())

				// Verify all fields updated
				updatedAsts := &workloadv1alpha1.AdvancedStatefulSet{}
				err = manager.client.Get(ctx, types.NamespacedName{
					Name:      identity.Name,
					Namespace: identity.Namespace,
				}, updatedAsts)
				Expect(err).NotTo(HaveOccurred())

				Expect(*updatedAsts.Spec.Replicas).To(Equal(int32(5)))
				Expect(updatedAsts.Spec.Template.Spec.Containers[0].Image).To(Equal("fluid-cache:v1.1.0"))
				Expect(updatedAsts.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU]).To(Equal(resource.MustParse("4")))
			})
		})

		Context("error handling", func() {
			It("should return error when AdvancedStatefulSet not found", func() {
				nonExistentIdentity := &common.ComponentIdentity{
					Name:      "non-existent",
					Namespace: "fluid",
				}
				spec := ComponentSpec{
					Replicas: func() *int32 { r := int32(5); return &r }(),
				}

				err := manager.SyncComponentSpec(ctx, nonExistentIdentity, spec)
				Expect(err).To(HaveOccurred())
			})

			It("should return error when no containers found", func() {
				// Create ASTS without containers
				emptyAsts := &workloadv1alpha1.AdvancedStatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "empty-containers",
						Namespace: "fluid",
					},
					Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
						Replicas: func() *int32 { r := int32(1); return &r }(),
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{},
							},
						},
					},
				}
				err := manager.client.Create(ctx, emptyAsts)
				Expect(err).NotTo(HaveOccurred())

				emptyIdentity := &common.ComponentIdentity{
					Name:      "empty-containers",
					Namespace: "fluid",
				}
				spec := ComponentSpec{
					Replicas: func() *int32 { r := int32(5); return &r }(),
				}

				err = manager.SyncComponentSpec(ctx, emptyIdentity, spec)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no containers found"))
			})
		})
	})

	Describe("updateReplicas", func() {
		It("should return true when replicas changed", func() {
			newReplicas := int32(5)
			result := manager.updateReplicas(existingAsts, newReplicas, GinkgoLogr)
			Expect(result).To(BeTrue())
			Expect(*existingAsts.Spec.Replicas).To(Equal(int32(5)))
		})

		It("should return false when replicas unchanged", func() {
			newReplicas := int32(3)
			result := manager.updateReplicas(existingAsts, newReplicas, GinkgoLogr)
			Expect(result).To(BeFalse())
		})
	})

	Describe("updateImage", func() {
		It("should update full image", func() {
			version := datav1alpha1.VersionSpec{
				Image:    "new-image",
				ImageTag: "v2.0.0",
			}
			result := manager.updateImage(existingAsts, version, GinkgoLogr)
			Expect(result).To(BeTrue())
			Expect(existingAsts.Spec.Template.Spec.Containers[0].Image).To(Equal("new-image:v2.0.0"))
		})

		It("should return false when only imageTag specified", func() {
			version := datav1alpha1.VersionSpec{
				ImageTag: "v1.1.0",
			}
			result := manager.updateImage(existingAsts, version, GinkgoLogr)
			Expect(result).To(BeFalse())
		})

		It("should return false when only image specified", func() {
			version := datav1alpha1.VersionSpec{
				Image: "new-image",
			}
			result := manager.updateImage(existingAsts, version, GinkgoLogr)
			Expect(result).To(BeFalse())
		})

		It("should return false when image unchanged", func() {
			version := datav1alpha1.VersionSpec{
				Image:    "fluid-cache",
				ImageTag: "v1.0.0",
			}
			result := manager.updateImage(existingAsts, version, GinkgoLogr)
			Expect(result).To(BeFalse())
		})

		It("should return false when no containers", func() {
			emptyAsts := &workloadv1alpha1.AdvancedStatefulSet{
				Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{},
						},
					},
				},
			}
			version := datav1alpha1.VersionSpec{
				ImageTag: "v2.0.0",
			}
			result := manager.updateImage(emptyAsts, version, GinkgoLogr)
			Expect(result).To(BeFalse())
		})
	})

	Describe("updateResources", func() {
		It("should update resources with DeepCopy", func() {
			resources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("4"),
					corev1.ResourceMemory: resource.MustParse("8Gi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("8"),
					corev1.ResourceMemory: resource.MustParse("16Gi"),
				},
			}
			result := manager.updateResources(existingAsts, resources, GinkgoLogr)
			Expect(result).To(BeTrue())

			container := existingAsts.Spec.Template.Spec.Containers[0]
			Expect(container.Resources.Requests[corev1.ResourceCPU]).To(Equal(resource.MustParse("4")))
			Expect(container.Resources.Limits[corev1.ResourceCPU]).To(Equal(resource.MustParse("8")))
		})

		It("should return false when resources unchanged", func() {
			resources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("2"),
					corev1.ResourceMemory: resource.MustParse("4Gi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("4"),
					corev1.ResourceMemory: resource.MustParse("8Gi"),
				},
			}
			result := manager.updateResources(existingAsts, resources, GinkgoLogr)
			Expect(result).To(BeFalse())
		})

		It("should update to nil resources (remove limits)", func() {
			resources := corev1.ResourceRequirements{
				Requests: nil,
				Limits:   nil,
			}
			result := manager.updateResources(existingAsts, resources, GinkgoLogr)
			Expect(result).To(BeTrue())

			container := existingAsts.Spec.Template.Spec.Containers[0]
			Expect(container.Resources.Requests).To(BeNil())
			Expect(container.Resources.Limits).To(BeNil())
		})

		It("should return false when no containers", func() {
			emptyAsts := &workloadv1alpha1.AdvancedStatefulSet{
				Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{},
						},
					},
				},
			}
			resources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("4"),
				},
			}
			result := manager.updateResources(emptyAsts, resources, GinkgoLogr)
			Expect(result).To(BeFalse())
		})
	})
})
