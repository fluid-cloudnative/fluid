/*
Copyright 2021 The Fluid Authors.

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

package juicefs

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

func TestTransformResourcesForWorkerNoValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.JuiceFSRuntime
		juicefsValue *JuiceFS
	}{
		{&datav1alpha1.JuiceFSRuntime{
			Spec: datav1alpha1.JuiceFSRuntimeSpec{},
		}, &JuiceFS{}},
	}
	for _, test := range tests {
		engine := &JuiceFSEngine{Log: log.NullLogger{}}
		engine.transformResourcesForWorker(test.runtime, test.juicefsValue)
		if result, found := test.juicefsValue.Worker.Resources.Limits[corev1.ResourceMemory]; found {
			t.Errorf("expected nil, got %v", result)
		}
	}
}

func TestTransformResourcesForWorkerWithValue(t *testing.T) {
	resources := corev1.ResourceRequirements{}
	resources.Limits = make(corev1.ResourceList)
	resources.Limits[corev1.ResourceMemory] = resource.MustParse("2Gi")

	var tests = []struct {
		runtime      *datav1alpha1.JuiceFSRuntime
		juicefsValue *JuiceFS
	}{
		{&datav1alpha1.JuiceFSRuntime{
			Spec: datav1alpha1.JuiceFSRuntimeSpec{
				Worker: datav1alpha1.JuiceFSCompTemplateSpec{
					Resources: resources,
				},
			},
		}, &JuiceFS{}},
	}
	for _, test := range tests {
		engine := &JuiceFSEngine{Log: log.NullLogger{}}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "juicefs", test.runtime.Spec.TieredStore)
		engine.UnitTest = true
		engine.transformResourcesForWorker(test.runtime, test.juicefsValue)
		if test.juicefsValue.Worker.Resources.Limits[corev1.ResourceMemory] != "2Gi" {
			t.Errorf("expected 22Gi, got %v", test.juicefsValue.Worker.Resources.Limits[corev1.ResourceMemory])
		}
	}
}

func TestTransformResourcesForFuseNoValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.JuiceFSRuntime
		juicefsValue *JuiceFS
	}{
		{&datav1alpha1.JuiceFSRuntime{
			Spec: datav1alpha1.JuiceFSRuntimeSpec{},
		}, &JuiceFS{}},
	}
	for _, test := range tests {
		engine := &JuiceFSEngine{Log: log.NullLogger{}}
		engine.transformResourcesForFuse(test.runtime, test.juicefsValue)
		if result, found := test.juicefsValue.Fuse.Resources.Limits[corev1.ResourceMemory]; found {
			t.Errorf("expected nil, got %v", result)
		}
	}
}

func TestTransformResourcesForFuseWithValue(t *testing.T) {

	resources := corev1.ResourceRequirements{}
	resources.Limits = make(corev1.ResourceList)
	resources.Limits[corev1.ResourceMemory] = resource.MustParse("2Gi")

	result := resource.MustParse("20Gi")

	var tests = []struct {
		runtime    *datav1alpha1.JuiceFSRuntime
		juiceValue *JuiceFS
	}{
		{&datav1alpha1.JuiceFSRuntime{
			Spec: datav1alpha1.JuiceFSRuntimeSpec{
				Fuse: datav1alpha1.JuiceFSFuseSpec{
					Resources: resources,
				},
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
					}},
				},
			},
		}, &JuiceFS{}},
	}
	for _, test := range tests {
		engine := &JuiceFSEngine{Log: log.NullLogger{}}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "juicefs", test.runtime.Spec.TieredStore)
		engine.UnitTest = true
		engine.transformResourcesForFuse(test.runtime, test.juiceValue)
		if test.juiceValue.Fuse.Resources.Limits[corev1.ResourceMemory] != "22Gi" {
			t.Errorf("expected 22Gi, got %v", test.juiceValue.Fuse.Resources.Limits[corev1.ResourceMemory])
		}
	}
}
