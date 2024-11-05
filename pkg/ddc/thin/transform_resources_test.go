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
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestThinEngine_transformResourcesForFuse(t1 *testing.T) {
	var tests = []struct {
		runtime *datav1alpha1.ThinRuntime
		value   *ThinValue
	}{
		{&datav1alpha1.ThinRuntime{
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
		}, &ThinValue{}},
	}
	for _, test := range tests {
		engine := &ThinEngine{
			Log:  fake.NullLogger(),
			name: test.runtime.Name,
		}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "thin", base.WithTieredStore(test.runtime.Spec.TieredStore))
		engine.UnitTest = true
		engine.transformResourcesForFuse(test.runtime.Spec.Fuse.Resources, test.value)
		wantMemReq := test.runtime.Spec.Fuse.Resources.Requests[corev1.ResourceMemory]
		wantCpuReq := test.runtime.Spec.Fuse.Resources.Requests[corev1.ResourceCPU]
		wantMemLim := test.runtime.Spec.Fuse.Resources.Limits[corev1.ResourceMemory]
		wantCpuLim := test.runtime.Spec.Fuse.Resources.Limits[corev1.ResourceCPU]
		if wantMemReq.String() != test.value.Fuse.Resources.Requests[corev1.ResourceMemory] ||
			wantCpuReq.String() != test.value.Fuse.Resources.Requests[corev1.ResourceCPU] ||
			wantMemLim.String() != test.value.Fuse.Resources.Limits[corev1.ResourceMemory] ||
			wantCpuLim.String() != test.value.Fuse.Resources.Limits[corev1.ResourceCPU] {
			t1.Errorf("expected %v, got %v", test.runtime.Spec.Fuse.Resources, test.value.Fuse.Resources)
		}
	}
}

func TestThinEngine_transformResourcesForWorker(t1 *testing.T) {
	var tests = []struct {
		runtime *datav1alpha1.ThinRuntime
		value   *ThinValue
	}{
		{&datav1alpha1.ThinRuntime{
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
		}, &ThinValue{}},
	}
	for _, test := range tests {
		engine := &ThinEngine{
			Log:  fake.NullLogger(),
			name: test.runtime.Name,
		}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "thin", base.WithTieredStore(test.runtime.Spec.TieredStore))
		engine.UnitTest = true
		engine.transformResourcesForWorker(test.runtime.Spec.Worker.Resources, test.value)
		wantMemReq := test.runtime.Spec.Worker.Resources.Requests[corev1.ResourceMemory]
		wantCpuReq := test.runtime.Spec.Worker.Resources.Requests[corev1.ResourceCPU]
		wantMemLim := test.runtime.Spec.Worker.Resources.Limits[corev1.ResourceMemory]
		wantCpuLim := test.runtime.Spec.Worker.Resources.Limits[corev1.ResourceCPU]
		if wantMemReq.String() != test.value.Worker.Resources.Requests[corev1.ResourceMemory] ||
			wantCpuReq.String() != test.value.Worker.Resources.Requests[corev1.ResourceCPU] ||
			wantMemLim.String() != test.value.Worker.Resources.Limits[corev1.ResourceMemory] ||
			wantCpuLim.String() != test.value.Worker.Resources.Limits[corev1.ResourceCPU] {
			t1.Errorf("expected %v, got %v", test.runtime.Spec.Worker.Resources, test.value.Worker.Resources)
		}
	}
}
