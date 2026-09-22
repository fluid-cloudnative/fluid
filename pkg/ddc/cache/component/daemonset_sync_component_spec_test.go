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

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var _ = Describe("DaemonSetManager SyncComponentSpec", func() {
	var (
		manager    *DaemonSetManager
		ctx        context.Context
		identity   *common.ComponentIdentity
		existingDs *appsv1.DaemonSet
	)

	BeforeEach(func() {
		client := setupTestClient()
		manager = newDaemonSetManager(client)
		ctx = context.Background()
		ctx = log.IntoContext(ctx, GinkgoLogr)

		identity = &common.ComponentIdentity{
			Name:      "test-runtime-client",
			Namespace: "fluid",
		}

		existingDs = &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      identity.Name,
				Namespace: identity.Namespace,
			},
			Spec: appsv1.DaemonSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:            "client",
								Image:           "fluid-fuse:v1.0.0",
								ImagePullPolicy: corev1.PullIfNotPresent,
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("1"),
										corev1.ResourceMemory: resource.MustParse("2Gi"),
									},
									Limits: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("2"),
										corev1.ResourceMemory: resource.MustParse("4Gi"),
									},
								},
							},
						},
					},
				},
			},
		}

		err := manager.client.Create(ctx, existingDs)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("SyncComponentSpec", func() {
		Context("when updating image", func() {
			It("should update full image (name + tag)", func() {
				spec := ComponentSpec{
					Version: datav1alpha1.VersionSpec{
						Image:    "new-fluid-fuse",
						ImageTag: "v2.0.0",
					},
				}

				err := manager.SyncComponentSpec(ctx, identity, spec)
				Expect(err).NotTo(HaveOccurred())

				updatedDs := &appsv1.DaemonSet{}
				err = manager.client.Get(ctx, types.NamespacedName{
					Name:      identity.Name,
					Namespace: identity.Namespace,
				}, updatedDs)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedDs.Spec.Template.Spec.Containers[0].Image).To(Equal("new-fluid-fuse:v2.0.0"))
			})

			It("should not update when only imageTag specified", func() {
				spec := ComponentSpec{
					Version: datav1alpha1.VersionSpec{
						ImageTag: "v1.1.0",
					},
				}

				err := manager.SyncComponentSpec(ctx, identity, spec)
				Expect(err).NotTo(HaveOccurred())

				updatedDs := &appsv1.DaemonSet{}
				err = manager.client.Get(ctx, types.NamespacedName{
					Name:      identity.Name,
					Namespace: identity.Namespace,
				}, updatedDs)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedDs.Spec.Template.Spec.Containers[0].Image).To(Equal("fluid-fuse:v1.0.0"))
			})

			It("should not update when image unchanged", func() {
				spec := ComponentSpec{
					Version: datav1alpha1.VersionSpec{
						Image:    "fluid-fuse",
						ImageTag: "v1.0.0",
					},
				}

				err := manager.SyncComponentSpec(ctx, identity, spec)
				Expect(err).NotTo(HaveOccurred())

				updatedDs := &appsv1.DaemonSet{}
				err = manager.client.Get(ctx, types.NamespacedName{
					Name:      identity.Name,
					Namespace: identity.Namespace,
				}, updatedDs)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedDs.Spec.Template.Spec.Containers[0].Image).To(Equal("fluid-fuse:v1.0.0"))
			})
		})

		Context("when updating resources", func() {
			It("should update both requests and limits", func() {
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

				updatedDs := &appsv1.DaemonSet{}
				err = manager.client.Get(ctx, types.NamespacedName{
					Name:      identity.Name,
					Namespace: identity.Namespace,
				}, updatedDs)
				Expect(err).NotTo(HaveOccurred())

				container := updatedDs.Spec.Template.Spec.Containers[0]
				Expect(container.Resources.Requests[corev1.ResourceCPU]).To(Equal(resource.MustParse("2")))
				Expect(container.Resources.Requests[corev1.ResourceMemory]).To(Equal(resource.MustParse("4Gi")))
				Expect(container.Resources.Limits[corev1.ResourceCPU]).To(Equal(resource.MustParse("4")))
				Expect(container.Resources.Limits[corev1.ResourceMemory]).To(Equal(resource.MustParse("8Gi")))
			})

			It("should not update when resources unchanged", func() {
				spec := ComponentSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("2Gi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("2"),
							corev1.ResourceMemory: resource.MustParse("4Gi"),
						},
					},
				}

				err := manager.SyncComponentSpec(ctx, identity, spec)
				Expect(err).NotTo(HaveOccurred())

				updatedDs := &appsv1.DaemonSet{}
				err = manager.client.Get(ctx, types.NamespacedName{
					Name:      identity.Name,
					Namespace: identity.Namespace,
				}, updatedDs)
				Expect(err).NotTo(HaveOccurred())

				container := updatedDs.Spec.Template.Spec.Containers[0]
				Expect(container.Resources.Requests[corev1.ResourceCPU]).To(Equal(resource.MustParse("1")))
			})
		})

		Context("when replicas is set (should be ignored)", func() {
			It("should not error and should not affect the DaemonSet", func() {
				replicas := int32(5)
				spec := ComponentSpec{
					Replicas: &replicas,
					Version: datav1alpha1.VersionSpec{
						Image:    "fluid-fuse",
						ImageTag: "v1.0.0",
					},
				}

				err := manager.SyncComponentSpec(ctx, identity, spec)
				Expect(err).NotTo(HaveOccurred())

				// DaemonSetSpec has no Replicas field to check; just confirm the call
				// succeeded and didn't touch anything else unexpectedly.
				updatedDs := &appsv1.DaemonSet{}
				err = manager.client.Get(ctx, types.NamespacedName{
					Name:      identity.Name,
					Namespace: identity.Namespace,
				}, updatedDs)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedDs.Spec.Template.Spec.Containers[0].Image).To(Equal("fluid-fuse:v1.0.0"))
			})
		})

		Context("when updating multiple fields", func() {
			It("should update image and resources together", func() {
				spec := ComponentSpec{
					Version: datav1alpha1.VersionSpec{
						Image:    "fluid-fuse",
						ImageTag: "v1.1.0",
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("2"),
						},
					},
				}

				err := manager.SyncComponentSpec(ctx, identity, spec)
				Expect(err).NotTo(HaveOccurred())

				updatedDs := &appsv1.DaemonSet{}
				err = manager.client.Get(ctx, types.NamespacedName{
					Name:      identity.Name,
					Namespace: identity.Namespace,
				}, updatedDs)
				Expect(err).NotTo(HaveOccurred())

				Expect(updatedDs.Spec.Template.Spec.Containers[0].Image).To(Equal("fluid-fuse:v1.1.0"))
				Expect(updatedDs.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU]).To(Equal(resource.MustParse("2")))
			})
		})

		Context("when nothing changed", func() {
			It("should skip patch entirely and return nil", func() {
				spec := ComponentSpec{}

				err := manager.SyncComponentSpec(ctx, identity, spec)
				Expect(err).NotTo(HaveOccurred())

				updatedDs := &appsv1.DaemonSet{}
				err = manager.client.Get(ctx, types.NamespacedName{
					Name:      identity.Name,
					Namespace: identity.Namespace,
				}, updatedDs)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedDs.Spec.Template.Spec.Containers[0].Image).To(Equal("fluid-fuse:v1.0.0"))
			})
		})

		Context("error handling", func() {
			It("should return error when DaemonSet not found", func() {
				nonExistentIdentity := &common.ComponentIdentity{
					Name:      "non-existent",
					Namespace: "fluid",
				}
				spec := ComponentSpec{
					Version: datav1alpha1.VersionSpec{
						Image:    "fluid-fuse",
						ImageTag: "v2.0.0",
					},
				}

				err := manager.SyncComponentSpec(ctx, nonExistentIdentity, spec)
				Expect(err).To(HaveOccurred())
			})

			It("should return error when no containers found", func() {
				emptyDs := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "empty-containers",
						Namespace: "fluid",
					},
					Spec: appsv1.DaemonSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{},
							},
						},
					},
				}
				err := manager.client.Create(ctx, emptyDs)
				Expect(err).NotTo(HaveOccurred())

				emptyIdentity := &common.ComponentIdentity{
					Name:      "empty-containers",
					Namespace: "fluid",
				}
				spec := ComponentSpec{
					Version: datav1alpha1.VersionSpec{
						Image:    "fluid-fuse",
						ImageTag: "v2.0.0",
					},
				}

				err = manager.SyncComponentSpec(ctx, emptyIdentity, spec)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no containers found"))
			})
		})
	})

	Describe("updateImage", func() {
		It("should update full image", func() {
			version := datav1alpha1.VersionSpec{
				Image:    "new-image",
				ImageTag: "v2.0.0",
			}
			result := manager.updateImage(existingDs, version, GinkgoLogr)
			Expect(result).To(BeTrue())
			Expect(existingDs.Spec.Template.Spec.Containers[0].Image).To(Equal("new-image:v2.0.0"))
		})

		It("should return false when only imageTag specified", func() {
			version := datav1alpha1.VersionSpec{
				ImageTag: "v1.1.0",
			}
			result := manager.updateImage(existingDs, version, GinkgoLogr)
			Expect(result).To(BeFalse())
		})

		It("should return false when image unchanged", func() {
			version := datav1alpha1.VersionSpec{
				Image:    "fluid-fuse",
				ImageTag: "v1.0.0",
			}
			result := manager.updateImage(existingDs, version, GinkgoLogr)
			Expect(result).To(BeFalse())
		})

		It("should return false when no containers", func() {
			emptyDs := &appsv1.DaemonSet{
				Spec: appsv1.DaemonSetSpec{
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
			result := manager.updateImage(emptyDs, version, GinkgoLogr)
			Expect(result).To(BeFalse())
		})
	})

	Describe("updateResources", func() {
		It("should update resources with DeepCopy", func() {
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
			result := manager.updateResources(existingDs, resources, GinkgoLogr)
			Expect(result).To(BeTrue())

			container := existingDs.Spec.Template.Spec.Containers[0]
			Expect(container.Resources.Requests[corev1.ResourceCPU]).To(Equal(resource.MustParse("2")))
			Expect(container.Resources.Limits[corev1.ResourceCPU]).To(Equal(resource.MustParse("4")))
		})

		It("should return false when resources unchanged", func() {
			resources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("2Gi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("2"),
					corev1.ResourceMemory: resource.MustParse("4Gi"),
				},
			}
			result := manager.updateResources(existingDs, resources, GinkgoLogr)
			Expect(result).To(BeFalse())
		})

		It("should return false when no containers", func() {
			emptyDs := &appsv1.DaemonSet{
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{},
						},
					},
				},
			}
			resources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("2"),
				},
			}
			result := manager.updateResources(emptyDs, resources, GinkgoLogr)
			Expect(result).To(BeFalse())
		})
	})
})
