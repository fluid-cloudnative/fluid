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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ThinEngine_transformResources", func() {
	Describe("transformResourcesForFuse", func() {
		It("should correctly transform fuse resources", func() {
			runtime := &datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: datav1alpha1.ThinRuntimeSpec{
					Fuse: datav1alpha1.ThinFuseSpec{
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("2Gi"),
								corev1.ResourceCPU:    resource.MustParse("2"),
							},
							Requests: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("2Gi"),
								corev1.ResourceCPU:    resource.MustParse("2"),
							},
						},
					},
				},
			}
			value := &ThinValue{}

			engine := &ThinEngine{
				Log:  fake.NullLogger(),
				name: runtime.Name,
			}
			engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "thin", base.WithTieredStore(runtime.Spec.TieredStore))
			engine.UnitTest = true
			engine.transformResourcesForFuse(runtime.Spec.Fuse.Resources, value)

			wantMemReq := runtime.Spec.Fuse.Resources.Requests[corev1.ResourceMemory]
			wantCpuReq := runtime.Spec.Fuse.Resources.Requests[corev1.ResourceCPU]
			wantMemLim := runtime.Spec.Fuse.Resources.Limits[corev1.ResourceMemory]
			wantCpuLim := runtime.Spec.Fuse.Resources.Limits[corev1.ResourceCPU]

			Expect(value.Fuse.Resources.Requests[corev1.ResourceMemory]).To(Equal(wantMemReq.String()))
			Expect(value.Fuse.Resources.Requests[corev1.ResourceCPU]).To(Equal(wantCpuReq.String()))
			Expect(value.Fuse.Resources.Limits[corev1.ResourceMemory]).To(Equal(wantMemLim.String()))
			Expect(value.Fuse.Resources.Limits[corev1.ResourceCPU]).To(Equal(wantCpuLim.String()))
		})
	})

	Describe("transformResourcesForWorker", func() {
		It("should correctly transform worker resources", func() {
			runtime := &datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: datav1alpha1.ThinRuntimeSpec{
					Worker: datav1alpha1.ThinCompTemplateSpec{
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("2Gi"),
								corev1.ResourceCPU:    resource.MustParse("2"),
							},
							Requests: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("2Gi"),
								corev1.ResourceCPU:    resource.MustParse("2"),
							},
						},
					},
				},
			}
			value := &ThinValue{}

			engine := &ThinEngine{
				Log:  fake.NullLogger(),
				name: runtime.Name,
			}
			engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "thin", base.WithTieredStore(runtime.Spec.TieredStore))
			engine.UnitTest = true
			engine.transformResourcesForWorker(runtime.Spec.Worker.Resources, value)

			wantMemReq := runtime.Spec.Worker.Resources.Requests[corev1.ResourceMemory]
			wantCpuReq := runtime.Spec.Worker.Resources.Requests[corev1.ResourceCPU]
			wantMemLim := runtime.Spec.Worker.Resources.Limits[corev1.ResourceMemory]
			wantCpuLim := runtime.Spec.Worker.Resources.Limits[corev1.ResourceCPU]

			Expect(value.Worker.Resources.Requests[corev1.ResourceMemory]).To(Equal(wantMemReq.String()))
			Expect(value.Worker.Resources.Requests[corev1.ResourceCPU]).To(Equal(wantCpuReq.String()))
			Expect(value.Worker.Resources.Limits[corev1.ResourceMemory]).To(Equal(wantMemLim.String()))
			Expect(value.Worker.Resources.Limits[corev1.ResourceCPU]).To(Equal(wantCpuLim.String()))
		})
	})
})
