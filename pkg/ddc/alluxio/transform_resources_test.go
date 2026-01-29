/*
Copyright 2020 The Fluid Authors.

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

package alluxio

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func mockAlluxioRuntimeForMaster(res corev1.ResourceRequirements) *datav1alpha1.AlluxioRuntime {
	return &datav1alpha1.AlluxioRuntime{
		Spec: datav1alpha1.AlluxioRuntimeSpec{
			Master: datav1alpha1.AlluxioCompTemplateSpec{
				Resources: res,
			},
			JobMaster: datav1alpha1.AlluxioCompTemplateSpec{
				Resources: res,
			},
		},
	}
}

var _ = Describe("TransformResourcesForMaster", func() {
	var engine *AlluxioEngine

	BeforeEach(func() {
		engine = &AlluxioEngine{}
	})

	DescribeTable("should transform master resources correctly",
		func(runtime *datav1alpha1.AlluxioRuntime, wantMasterResources, wantJobMasterResources common.Resources) {
			got := &Alluxio{}
			engine.transformResourcesForMaster(runtime, got)
			Expect(got.Master.Resources).To(Equal(wantMasterResources))
			Expect(got.JobMaster.Resources).To(Equal(wantJobMasterResources))
		},
		Entry("with limits and requests",
			mockAlluxioRuntimeForMaster(corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("100Mi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("400m"),
					corev1.ResourceMemory: resource.MustParse("400Mi"),
				},
			}),
			common.Resources{
				Requests: common.ResourceList{
					corev1.ResourceCPU:    "100m",
					corev1.ResourceMemory: "100Mi",
				},
				Limits: common.ResourceList{
					corev1.ResourceCPU:    "400m",
					corev1.ResourceMemory: "400Mi",
				},
			},
			common.Resources{
				Requests: common.ResourceList{
					corev1.ResourceCPU:    "100m",
					corev1.ResourceMemory: "100Mi",
				},
				Limits: common.ResourceList{
					corev1.ResourceCPU:    "400m",
					corev1.ResourceMemory: "400Mi",
				},
			},
		),
		Entry("with only requests",
			mockAlluxioRuntimeForMaster(corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("100Mi"),
				},
			}),
			common.Resources{
				Requests: common.ResourceList{
					corev1.ResourceCPU:    "100m",
					corev1.ResourceMemory: "100Mi",
				},
				Limits: common.ResourceList{},
			},
			common.Resources{
				Requests: common.ResourceList{
					corev1.ResourceCPU:    "100m",
					corev1.ResourceMemory: "100Mi",
				},
				Limits: common.ResourceList{},
			},
		),
		Entry("with empty requests",
			mockAlluxioRuntimeForMaster(corev1.ResourceRequirements{
				Requests: corev1.ResourceList{},
			}),
			common.Resources{
				Requests: common.ResourceList{},
				Limits:   common.ResourceList{},
			},
			common.Resources{
				Requests: common.ResourceList{},
				Limits:   common.ResourceList{},
			},
		),
		Entry("with empty requirements",
			mockAlluxioRuntimeForMaster(corev1.ResourceRequirements{}),
			common.Resources{
				Requests: common.ResourceList{},
				Limits:   common.ResourceList{},
			},
			common.Resources{
				Requests: common.ResourceList{},
				Limits:   common.ResourceList{},
			},
		),
		Entry("with empty limits",
			mockAlluxioRuntimeForMaster(corev1.ResourceRequirements{
				Limits: corev1.ResourceList{},
			}),
			common.Resources{
				Requests: common.ResourceList{},
				Limits:   common.ResourceList{},
			},
			common.Resources{
				Requests: common.ResourceList{},
				Limits:   common.ResourceList{},
			},
		),
	)
})

var _ = Describe("TransformResourcesForWorker", func() {
	var engine *AlluxioEngine

	BeforeEach(func() {
		engine = &AlluxioEngine{
			Log:       fake.NullLogger(),
			name:      "test",
			namespace: "test",
		}
	})

	Context("when no resource values are specified", func() {
		It("should not set memory limit", func() {
			testRuntime := &datav1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					TieredStore: datav1alpha1.TieredStore{},
				},
			}
			alluxioValue := &Alluxio{
				Properties: map[string]string{},
			}

			engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "alluxio", base.WithTieredStore(testRuntime.Spec.TieredStore))
			client := fake.NewFakeClientWithScheme(testScheme, testRuntime.DeepCopy())
			engine.Client = client

			err := engine.transformResourcesForWorker(testRuntime, alluxioValue)
			Expect(err).NotTo(HaveOccurred())
			_, found := alluxioValue.Worker.Resources.Limits[corev1.ResourceMemory]
			Expect(found).To(BeFalse())
		})
	})

	Context("when tiered store is configured", func() {
		It("should set memory request from tiered store quota", func() {
			result := resource.MustParse("20Gi")
			testRuntime := &datav1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					TieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{{
							MediumType: common.Memory,
							Quota:      &result,
						}},
					},
				},
			}
			alluxioValue := &Alluxio{
				Properties: map[string]string{},
			}

			engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "alluxio", base.WithTieredStore(testRuntime.Spec.TieredStore))
			client := fake.NewFakeClientWithScheme(testScheme, testRuntime.DeepCopy())
			engine.Client = client

			err := engine.transformResourcesForWorker(testRuntime, alluxioValue)
			Expect(err).NotTo(HaveOccurred())
			Expect(alluxioValue.Worker.Resources.Requests[corev1.ResourceMemory]).To(Equal("20Gi"))
			_, found := alluxioValue.Worker.Resources.Limits[corev1.ResourceMemory]
			Expect(found).To(BeFalse())
		})
	})

	Context("when resource limits and requests are specified with tiered store", func() {
		It("should return error when memory limit is less than tiered store quota", func() {
			resources := corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("2Gi"),
					corev1.ResourceCPU:    resource.MustParse("500m"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("1Gi"),
					corev1.ResourceCPU:    resource.MustParse("500m"),
				},
			}
			result := resource.MustParse("20Gi")
			testRuntime := &datav1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Worker: datav1alpha1.AlluxioCompTemplateSpec{
						Resources: resources,
					},
					JobWorker: datav1alpha1.AlluxioCompTemplateSpec{
						Resources: resources,
					},
					TieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{{
							MediumType: common.Memory,
							Quota:      &result,
						}},
					},
				},
			}
			alluxioValue := &Alluxio{
				Properties: map[string]string{},
				Master:     Master{},
				JobMaster:  JobMaster{},
			}

			engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "alluxio", base.WithTieredStore(testRuntime.Spec.TieredStore))
			client := fake.NewFakeClientWithScheme(testScheme, testRuntime.DeepCopy())
			engine.Client = client
			engine.UnitTest = true

			err := engine.transformResourcesForWorker(testRuntime, alluxioValue)
			Expect(err).To(HaveOccurred())
			Expect(alluxioValue.Worker.Resources.Limits[corev1.ResourceMemory]).To(Equal("2Gi"))
			Expect(alluxioValue.Worker.Resources.Requests[corev1.ResourceMemory]).To(Equal("1Gi"))
		})

		It("should adjust memory request when memory limit is sufficient", func() {
			resources := corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("20Gi"),
					corev1.ResourceCPU:    resource.MustParse("500m"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("1Gi"),
					corev1.ResourceCPU:    resource.MustParse("500m"),
				},
			}
			result := resource.MustParse("20Gi")
			testRuntime := &datav1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Worker: datav1alpha1.AlluxioCompTemplateSpec{
						Resources: resources,
					},
					JobWorker: datav1alpha1.AlluxioCompTemplateSpec{
						Resources: resources,
					},
					TieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{{
							MediumType: common.Memory,
							Quota:      &result,
						}},
					},
				},
			}
			alluxioValue := &Alluxio{
				Properties: map[string]string{},
				Master:     Master{},
				JobMaster:  JobMaster{},
			}

			engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "alluxio", base.WithTieredStore(testRuntime.Spec.TieredStore))
			client := fake.NewFakeClientWithScheme(testScheme, testRuntime.DeepCopy())
			engine.Client = client
			engine.UnitTest = true

			err := engine.transformResourcesForWorker(testRuntime, alluxioValue)
			Expect(err).NotTo(HaveOccurred())
			Expect(alluxioValue.Worker.Resources.Limits[corev1.ResourceMemory]).To(Equal("20Gi"))
			Expect(alluxioValue.Worker.Resources.Requests[corev1.ResourceMemory]).To(Equal("20Gi"))
		})
	})

	Context("when only requests are specified", func() {
		It("should adjust memory request to tiered store quota when configured", func() {
			resources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("1Gi"),
					corev1.ResourceCPU:    resource.MustParse("500m"),
				},
			}
			result := resource.MustParse("20Gi")
			testRuntime := &datav1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Worker: datav1alpha1.AlluxioCompTemplateSpec{
						Resources: resources,
					},
					JobWorker: datav1alpha1.AlluxioCompTemplateSpec{
						Resources: resources,
					},
					TieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{{
							MediumType: common.Memory,
							Quota:      &result,
						}},
					},
				},
			}
			alluxioValue := &Alluxio{
				Properties: map[string]string{},
				Master:     Master{},
				JobMaster:  JobMaster{},
			}

			engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "alluxio", base.WithTieredStore(testRuntime.Spec.TieredStore))
			engine.UnitTest = true
			client := fake.NewFakeClientWithScheme(testScheme, testRuntime.DeepCopy())
			engine.Client = client

			err := engine.transformResourcesForWorker(testRuntime, alluxioValue)
			Expect(err).NotTo(HaveOccurred())
			_, found := alluxioValue.Worker.Resources.Limits[corev1.ResourceMemory]
			Expect(found).To(BeFalse())
			Expect(alluxioValue.Worker.Resources.Requests[corev1.ResourceMemory]).To(Equal("20Gi"))
		})

		It("should keep original memory request when no tiered store", func() {
			resources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("1Gi"),
					corev1.ResourceCPU:    resource.MustParse("500m"),
				},
			}
			testRuntime := &datav1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Worker: datav1alpha1.AlluxioCompTemplateSpec{
						Resources: resources,
					},
					JobWorker: datav1alpha1.AlluxioCompTemplateSpec{
						Resources: resources,
					},
					TieredStore: datav1alpha1.TieredStore{},
				},
			}
			alluxioValue := &Alluxio{
				Properties: map[string]string{},
				Master:     Master{},
				JobMaster:  JobMaster{},
			}

			engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "alluxio", base.WithTieredStore(testRuntime.Spec.TieredStore))
			engine.UnitTest = true
			client := fake.NewFakeClientWithScheme(testScheme, testRuntime.DeepCopy())
			engine.Client = client

			err := engine.transformResourcesForWorker(testRuntime, alluxioValue)
			Expect(err).NotTo(HaveOccurred())
			_, found := alluxioValue.Worker.Resources.Limits[corev1.ResourceMemory]
			Expect(found).To(BeFalse())
			Expect(alluxioValue.Worker.Resources.Requests[corev1.ResourceMemory]).To(Equal("1Gi"))
		})
	})

	Context("when only limits are specified", func() {
		It("should set memory request from tiered store quota when configured", func() {
			resources := corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("20Gi"),
					corev1.ResourceCPU:    resource.MustParse("500m"),
				},
			}
			result := resource.MustParse("20Gi")
			testRuntime := &datav1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Worker: datav1alpha1.AlluxioCompTemplateSpec{
						Resources: resources,
					},
					JobWorker: datav1alpha1.AlluxioCompTemplateSpec{
						Resources: resources,
					},
					TieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{{
							MediumType: common.Memory,
							Quota:      &result,
						}},
					},
				},
			}
			alluxioValue := &Alluxio{
				Properties: map[string]string{},
				Master:     Master{},
				JobMaster:  JobMaster{},
			}

			engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "alluxio", base.WithTieredStore(testRuntime.Spec.TieredStore))
			engine.UnitTest = true
			client := fake.NewFakeClientWithScheme(testScheme, testRuntime.DeepCopy())
			engine.Client = client

			err := engine.transformResourcesForWorker(testRuntime, alluxioValue)
			Expect(err).NotTo(HaveOccurred())
			Expect(alluxioValue.Worker.Resources.Limits[corev1.ResourceMemory]).To(Equal("20Gi"))
			Expect(alluxioValue.Worker.Resources.Requests[corev1.ResourceMemory]).To(Equal("20Gi"))
		})

		It("should not set memory request when no tiered store", func() {
			resources := corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("20Gi"),
					corev1.ResourceCPU:    resource.MustParse("500m"),
				},
			}
			testRuntime := &datav1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Worker: datav1alpha1.AlluxioCompTemplateSpec{
						Resources: resources,
					},
					JobWorker: datav1alpha1.AlluxioCompTemplateSpec{
						Resources: resources,
					},
					TieredStore: datav1alpha1.TieredStore{},
				},
			}
			alluxioValue := &Alluxio{
				Properties: map[string]string{},
				Master:     Master{},
				JobMaster:  JobMaster{},
			}

			engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "alluxio", base.WithTieredStore(testRuntime.Spec.TieredStore))
			engine.UnitTest = true
			client := fake.NewFakeClientWithScheme(testScheme, testRuntime.DeepCopy())
			engine.Client = client

			err := engine.transformResourcesForWorker(testRuntime, alluxioValue)
			Expect(err).NotTo(HaveOccurred())
			Expect(alluxioValue.Worker.Resources.Limits[corev1.ResourceMemory]).To(Equal("20Gi"))
			_, found := alluxioValue.Worker.Resources.Requests[corev1.ResourceMemory]
			Expect(found).To(BeFalse())
		})
	})
})

var _ = Describe("TransformResourcesForFuse", func() {
	var engine *AlluxioEngine

	BeforeEach(func() {
		engine = &AlluxioEngine{
			Log:       fake.NullLogger(),
			name:      "test",
			namespace: "test",
		}
	})

	Context("when no resource values are specified", func() {
		It("should not set memory limit", func() {
			testRuntime := &datav1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: datav1alpha1.AlluxioRuntimeSpec{},
			}
			alluxioValue := &Alluxio{
				Properties: map[string]string{},
			}

			engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "alluxio")
			engine.Client = fake.NewFakeClientWithScheme(testScheme, testRuntime.DeepCopy())
			engine.transformResourcesForFuse(testRuntime, alluxioValue)
			_, found := alluxioValue.Fuse.Resources.Limits[corev1.ResourceMemory]
			Expect(found).To(BeFalse())
		})
	})

	Context("when resource values are specified with tiered store", func() {
		It("should add tiered store quota to memory limit", func() {
			resources := corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("2Gi"),
				},
			}
			result := resource.MustParse("20Gi")
			testRuntime := &datav1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Fuse: datav1alpha1.AlluxioFuseSpec{
						Resources: resources,
					},
					TieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{{
							MediumType: common.Memory,
							Quota:      &result,
						}},
					},
				},
			}
			alluxioValue := &Alluxio{
				Properties: map[string]string{},
				Master:     Master{},
			}

			engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "alluxio", base.WithTieredStore(testRuntime.Spec.TieredStore))
			engine.UnitTest = true
			engine.Client = fake.NewFakeClientWithScheme(testScheme, testRuntime.DeepCopy())
			engine.transformResourcesForFuse(testRuntime, alluxioValue)
			Expect(alluxioValue.Fuse.Resources.Limits[corev1.ResourceMemory]).To(Equal("22Gi"))
		})
	})
})
