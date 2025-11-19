package jindocache

import (
	"context"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JindoCacheEngine resources transformation tests", Label("pkg.ddc.jindocache.transform_resources_test.go"), func() {
	var (
		dataset      *datav1alpha1.Dataset
		jindoruntime *datav1alpha1.JindoRuntime
		engine       *JindoCacheEngine
		client       client.Client
		resources    []runtime.Object
	)

	BeforeEach(func() {
		dataset, jindoruntime = mockFluidObjectsForTests(types.NamespacedName{Namespace: "fluid", Name: "jindo-demo"})
		engine = mockJindoCacheEngineForTests(dataset, jindoruntime)
		resources = []runtime.Object{dataset, jindoruntime}
	})

	// JustBeforeEach is guaranteed to run after every BeforeEach()
	// So it's easy to modify resources' specs with an extra BeforeEach()
	JustBeforeEach(func() {
		client = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
		engine.Client = client
	})

	Describe("Test JindoCacheEngine.transformMasterResources()", func() {
		It("Should not set resources in master when no resources set in runtime.spec.master.resources", func() {
			jindoruntime.Spec.Master.Resources = corev1.ResourceRequirements{}
			value := Jindo{}
			Expect(engine.transformMasterResources(jindoruntime, &value, "")).NotTo(HaveOccurred())

			Expect(value.Master.Resources).To(Equal(common.Resources{}))
		})

		It("Should set resources in master when resources set in runtime.spec.master.resources", func() {
			jindoruntime.Spec.Master.Resources = corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("200m"),
					corev1.ResourceMemory: resource.MustParse("2Gi"),
				},
			}
			value := Jindo{}
			Expect(engine.transformMasterResources(jindoruntime, &value, "")).NotTo(HaveOccurred())

			Expect(value.Master.Resources.Requests).NotTo(BeNil())
			Expect(value.Master.Resources.Limits).NotTo(BeNil())
			Expect(value.Master.Resources.Requests[corev1.ResourceCPU]).To(Equal("100m"))
			Expect(value.Master.Resources.Requests[corev1.ResourceMemory]).To(Equal("1Gi"))
			Expect(value.Master.Resources.Limits[corev1.ResourceCPU]).To(Equal("200m"))
			Expect(value.Master.Resources.Limits[corev1.ResourceMemory]).To(Equal("2Gi"))
		})

		It("Should only set requests when only requests set in runtime.spec.master.resources", func() {
			jindoruntime.Spec.Master.Resources = corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
			}
			value := Jindo{}
			Expect(engine.transformMasterResources(jindoruntime, &value, "")).NotTo(HaveOccurred())

			Expect(value.Master.Resources.Requests).NotTo(BeNil())
			Expect(value.Master.Resources.Limits).To(BeNil())
			Expect(value.Master.Resources.Requests[corev1.ResourceCPU]).To(Equal("100m"))
			Expect(value.Master.Resources.Requests[corev1.ResourceMemory]).To(Equal("1Gi"))
		})

		It("Should only set limits when only limits set in runtime.spec.master.resources", func() {
			jindoruntime.Spec.Master.Resources = corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("200m"),
					corev1.ResourceMemory: resource.MustParse("2Gi"),
				},
			}
			value := Jindo{}
			Expect(engine.transformMasterResources(jindoruntime, &value, "")).NotTo(HaveOccurred())

			Expect(value.Master.Resources.Requests).To(BeNil())
			Expect(value.Master.Resources.Limits).NotTo(BeNil())
			Expect(value.Master.Resources.Limits[corev1.ResourceCPU]).To(Equal("200m"))
			Expect(value.Master.Resources.Limits[corev1.ResourceMemory]).To(Equal("2Gi"))
		})

		It("Should set extra custom resources when they are set in runtime.spec.master.resources", func() {
			resourceName := "my-custom-resource"
			jindoruntime.Spec.Master.Resources = corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceName(resourceName): resource.MustParse("2"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceName(resourceName): resource.MustParse("2"),
				},
			}

			value := Jindo{}
			Expect(engine.transformMasterResources(jindoruntime, &value, "")).NotTo(HaveOccurred())

			Expect(value.Master.Resources.Requests).NotTo(BeNil())
			Expect(value.Master.Resources.Limits).NotTo(BeNil())
			Expect(value.Master.Resources.Requests[corev1.ResourceName(resourceName)]).To(Equal("2"))
			Expect(value.Master.Resources.Limits[corev1.ResourceName(resourceName)]).To(Equal("2"))
		})
	})

	Describe("Test JindoCacheEngine.transformWorkerResources()", func() {
		When("JindoRuntime have a non-mem type tieredstore", func() {
			BeforeEach(func() {
				ssdQuota := resource.MustParse("50Gi")
				jindoruntime.Spec.TieredStore = datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{
						{
							MediumType: common.SSD,
							Quota:      &ssdQuota,
							VolumeType: common.VolumeTypeEmptyDir,
						},
					},
				}
			})

			It("Should not set resources in worker when no resources set in runtime.spec.worker.resources", func() {
				jindoruntime.Spec.Worker.Resources = corev1.ResourceRequirements{}
				value := Jindo{}
				Expect(engine.transformWorkerResources(jindoruntime, &value, "")).NotTo(HaveOccurred())
				Expect(value.Worker.Resources).To(Equal(common.Resources{}))
			})

			It("Should set resources in worker when resources set in runtime.spec.worker.resources", func() {
				jindoruntime.Spec.Worker.Resources = corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("1Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("200m"),
						corev1.ResourceMemory: resource.MustParse("2Gi"),
					},
				}
				value := Jindo{}
				Expect(engine.transformWorkerResources(jindoruntime, &value, "")).NotTo(HaveOccurred())

				Expect(value.Worker.Resources.Requests).NotTo(BeNil())
				Expect(value.Worker.Resources.Limits).NotTo(BeNil())
				Expect(value.Worker.Resources.Requests[corev1.ResourceCPU]).To(Equal("100m"))
				Expect(value.Worker.Resources.Requests[corev1.ResourceMemory]).To(Equal("1Gi"))
				Expect(value.Worker.Resources.Limits[corev1.ResourceCPU]).To(Equal("200m"))
				Expect(value.Worker.Resources.Limits[corev1.ResourceMemory]).To(Equal("2Gi"))
			})

			It("Should only set requests when only requests set in runtime.spec.worker.resources", func() {
				jindoruntime.Spec.Worker.Resources = corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("1Gi"),
					},
				}
				value := Jindo{}
				Expect(engine.transformWorkerResources(jindoruntime, &value, "")).NotTo(HaveOccurred())

				Expect(value.Worker.Resources.Requests).NotTo(BeNil())
				Expect(value.Worker.Resources.Limits).To(BeNil())
				Expect(value.Worker.Resources.Requests[corev1.ResourceCPU]).To(Equal("100m"))
				Expect(value.Worker.Resources.Requests[corev1.ResourceMemory]).To(Equal("1Gi"))
			})

			It("Should only set limits when only limits set in runtime.spec.worker.resources", func() {
				jindoruntime.Spec.Worker.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("200m"),
						corev1.ResourceMemory: resource.MustParse("2Gi"),
					},
				}
				value := Jindo{}
				Expect(engine.transformWorkerResources(jindoruntime, &value, "")).NotTo(HaveOccurred())

				Expect(value.Worker.Resources.Requests).To(BeNil())
				Expect(value.Worker.Resources.Limits).NotTo(BeNil())
				Expect(value.Worker.Resources.Limits[corev1.ResourceCPU]).To(Equal("200m"))
				Expect(value.Worker.Resources.Limits[corev1.ResourceMemory]).To(Equal("2Gi"))
			})

			It("Should set extra custom resources when they are set in runtime.spec.worker.resources", func() {
				resourceName := "my-custom-resource"
				jindoruntime.Spec.Worker.Resources = corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceName(resourceName): resource.MustParse("2"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceName(resourceName): resource.MustParse("2"),
					},
				}

				value := Jindo{}
				Expect(engine.transformWorkerResources(jindoruntime, &value, "")).NotTo(HaveOccurred())

				Expect(value.Worker.Resources.Requests).NotTo(BeNil())
				Expect(value.Worker.Resources.Limits).NotTo(BeNil())
				Expect(value.Worker.Resources.Requests[corev1.ResourceName(resourceName)]).To(Equal("2"))
				Expect(value.Worker.Resources.Limits[corev1.ResourceName(resourceName)]).To(Equal("2"))
			})
		})

		When("JindoRuntime have a MEM type tieredstore level set", func() {
			memQuota := resource.MustParse("4Gi")
			BeforeEach(func() {
				jindoruntime.Spec.TieredStore = datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{
						{
							MediumType: common.Memory,
							Quota:      &memQuota,
							VolumeType: common.VolumeTypeEmptyDir,
						},
					},
				}
			})

			It("should automatically add memory request if no memory request is set in runtime.spec.worker.resources", func() {
				jindoruntime.Spec.Worker.Resources = corev1.ResourceRequirements{}
				value := Jindo{}
				Expect(engine.transformWorkerResources(jindoruntime, &value, memQuota.String())).NotTo(HaveOccurred())
				Expect(value.Worker.Resources.Requests[corev1.ResourceMemory]).To(Equal("4Gi"))
			})

			It("should automatically add memory request if memory request is lower than mem quota", func() {
				jindoruntime.Spec.Worker.Resources = corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("2Gi"),
					},
				}
				value := Jindo{}
				Expect(engine.transformWorkerResources(jindoruntime, &value, memQuota.String())).NotTo(HaveOccurred())
				Expect(value.Worker.Resources.Requests[corev1.ResourceMemory]).To(Equal("4Gi"))

				updatedRuntime := &datav1alpha1.JindoRuntime{}
				Expect(client.Get(context.TODO(), types.NamespacedName{Name: jindoruntime.Name, Namespace: jindoruntime.Namespace}, updatedRuntime)).NotTo(HaveOccurred())
				Expect(updatedRuntime.Spec.Worker.Resources.Requests[corev1.ResourceMemory]).To(Equal(resource.MustParse("4Gi")))
			})

			It("should use user-defined memory request if it is larger than mem quota", func() {
				jindoruntime.Spec.Worker.Resources = corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("5Gi"),
					},
				}
				value := Jindo{}
				Expect(engine.transformWorkerResources(jindoruntime, &value, memQuota.String())).NotTo(HaveOccurred())
				Expect(value.Worker.Resources.Requests[corev1.ResourceMemory]).To(Equal("5Gi"))
			})

			It("should return error when user-defined memory limit is less than mem quota", func() {
				jindoruntime.Spec.Worker.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("3Gi"),
					},
				}
				value := Jindo{}
				err := engine.transformWorkerResources(jindoruntime, &value, memQuota.String())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("greater than worker limits memory"))
			})
		})

	})

	Describe("Test JindoCacheEngine.transformFuseResources()", func() {
		It("Should not set resources in fuse when no resources set in runtime.spec.fuse.resources", func() {
			jindoruntime.Spec.Fuse.Resources = corev1.ResourceRequirements{}
			value := Jindo{}
			engine.transformFuseResources(jindoruntime, &value)

			Expect(value.Fuse.Resources).To(Equal(common.Resources{}))
		})

		It("Should set resources in fuse when resources set in runtime.spec.fuse.resources", func() {
			jindoruntime.Spec.Fuse.Resources = corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("200m"),
					corev1.ResourceMemory: resource.MustParse("2Gi"),
				},
			}
			value := Jindo{}
			engine.transformFuseResources(jindoruntime, &value)

			Expect(value.Fuse.Resources.Requests).NotTo(BeNil())
			Expect(value.Fuse.Resources.Limits).NotTo(BeNil())
			Expect(value.Fuse.Resources.Requests[corev1.ResourceCPU]).To(Equal("100m"))
			Expect(value.Fuse.Resources.Requests[corev1.ResourceMemory]).To(Equal("1Gi"))
			Expect(value.Fuse.Resources.Limits[corev1.ResourceCPU]).To(Equal("200m"))
			Expect(value.Fuse.Resources.Limits[corev1.ResourceMemory]).To(Equal("2Gi"))
		})

		It("Should only set requests when only requests set in runtime.spec.fuse.resources", func() {
			jindoruntime.Spec.Fuse.Resources = corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
			}
			value := Jindo{}
			engine.transformFuseResources(jindoruntime, &value)

			Expect(value.Fuse.Resources.Requests).NotTo(BeNil())
			Expect(value.Fuse.Resources.Limits).To(BeNil())
			Expect(value.Fuse.Resources.Requests[corev1.ResourceCPU]).To(Equal("100m"))
			Expect(value.Fuse.Resources.Requests[corev1.ResourceMemory]).To(Equal("1Gi"))
		})

		It("Should only set limits when only limits set in runtime.spec.fuse.resources", func() {
			jindoruntime.Spec.Fuse.Resources = corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("200m"),
					corev1.ResourceMemory: resource.MustParse("2Gi"),
				},
			}
			value := Jindo{}
			engine.transformFuseResources(jindoruntime, &value)

			Expect(value.Fuse.Resources.Requests).To(BeNil())
			Expect(value.Fuse.Resources.Limits).NotTo(BeNil())
			Expect(value.Fuse.Resources.Limits[corev1.ResourceCPU]).To(Equal("200m"))
			Expect(value.Fuse.Resources.Limits[corev1.ResourceMemory]).To(Equal("2Gi"))
		})

		It("Should set extra custom resources when they are set in runtime.spec.fuse.resources", func() {
			resourceName := "my-custom-resource"
			jindoruntime.Spec.Fuse.Resources = corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceName(resourceName): resource.MustParse("2"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceName(resourceName): resource.MustParse("2"),
				},
			}

			value := Jindo{}
			engine.transformFuseResources(jindoruntime, &value)

			Expect(value.Fuse.Resources.Requests).NotTo(BeNil())
			Expect(value.Fuse.Resources.Limits).NotTo(BeNil())
			Expect(value.Fuse.Resources.Requests[corev1.ResourceName(resourceName)]).To(Equal("2"))
			Expect(value.Fuse.Resources.Limits[corev1.ResourceName(resourceName)]).To(Equal("2"))
		})
	})

})
