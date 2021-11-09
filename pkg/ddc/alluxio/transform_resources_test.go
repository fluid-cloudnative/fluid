/*

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
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestTransformResourcesForMaster(t *testing.T) {
	testCases := map[string]struct {
		runtime *datav1alpha1.AlluxioRuntime
		got     *Alluxio
		want    *Alluxio
	}{
		"test alluxio master pass through resources with limits and request case 1": {
			runtime: mockAlluxioRuntimeForMaster(
				corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("100Mi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("400m"),
						corev1.ResourceMemory: resource.MustParse("400Mi"),
					},
				},
			),
			got: &Alluxio{},
			want: &Alluxio{
				Master: Master{
					Resources: common.Resources{
						Requests: common.ResourceList{
							corev1.ResourceCPU:    "100m",
							corev1.ResourceMemory: "100Mi",
						},
						Limits: common.ResourceList{
							corev1.ResourceCPU:    "400m",
							corev1.ResourceMemory: "400Mi",
						},
					},
				},
			},
		},
		"test alluxio master pass through resources with request case 1": {
			runtime: mockAlluxioRuntimeForMaster(
				corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("100Mi"),
					},
				},
			),
			got: &Alluxio{},
			want: &Alluxio{
				Master: Master{
					Resources: common.Resources{
						Requests: common.ResourceList{
							corev1.ResourceCPU:    "100m",
							corev1.ResourceMemory: "100Mi",
						},
					},
				},
			},
		},
		"test alluxio master pass through resources without request and limit case 1": {
			runtime: mockAlluxioRuntimeForMaster(
				corev1.ResourceRequirements{
					Requests: corev1.ResourceList{},
				},
			),
			got:  &Alluxio{},
			want: &Alluxio{},
		},
		"test alluxio master pass through resources without request and limit case 2": {
			runtime: mockAlluxioRuntimeForMaster(corev1.ResourceRequirements{}),
			got:     &Alluxio{},
			want:    &Alluxio{},
		},
		"test alluxio master pass through resources without request and limit case 3": {
			runtime: mockAlluxioRuntimeForMaster(
				corev1.ResourceRequirements{
					Limits: corev1.ResourceList{},
				},
			),
			got:  &Alluxio{},
			want: &Alluxio{},
		},
	}

	engine := &AlluxioEngine{}
	for k, item := range testCases {
		engine.transformResourcesForMaster(item.runtime, item.got)
		if !reflect.DeepEqual(item.want.Master.Resources, item.got.Master.Resources) {
			t.Errorf("%s failure, want resource: %+v,got resource: %+v",
				k,
				item.want.Master.Resources,
				item.got.Master.Resources,
			)
		}
	}
}

func mockAlluxioRuntimeForMaster(res corev1.ResourceRequirements) *datav1alpha1.AlluxioRuntime {
	runtime := &datav1alpha1.AlluxioRuntime{
		Spec: datav1alpha1.AlluxioRuntimeSpec{
			Master: datav1alpha1.AlluxioCompTemplateSpec{
				Resources: res,
			},
		},
	}
	return runtime

}

func TestTransformResourcesForWorkerNoValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{},
		}, &Alluxio{
			Properties: map[string]string{},
		}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{Log: log.NullLogger{}}
		engine.transformResourcesForWorker(test.runtime, test.alluxioValue)
		if result, found := test.alluxioValue.Worker.Resources.Limits[corev1.ResourceMemory]; found {
			t.Errorf("expected nil, got %v", result)
		}
	}
}

func TestTransformResourcesForWorkerWithValue(t *testing.T) {

	resources := corev1.ResourceRequirements{}
	resources.Limits = make(corev1.ResourceList)
	resources.Limits[corev1.ResourceMemory] = resource.MustParse("2Gi")

	result := resource.MustParse("20Gi")

	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Worker: datav1alpha1.AlluxioCompTemplateSpec{
					Resources: resources,
				},
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
					}},
				},
			},
		}, &Alluxio{
			Properties: map[string]string{},
			Master:     Master{},
		}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{Log: log.NullLogger{}}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "alluxio", test.runtime.Spec.TieredStore)
		engine.UnitTest = true
		engine.transformResourcesForWorker(test.runtime, test.alluxioValue)
		if test.alluxioValue.Worker.Resources.Limits[corev1.ResourceMemory] != "22Gi" {
			t.Errorf("expected 22Gi, got %v", test.alluxioValue.Worker.Resources.Limits[corev1.ResourceMemory])
		}
	}
}

func TestTransformResourcesForFuseNoValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{},
		}, &Alluxio{
			Properties: map[string]string{},
		}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{Log: log.NullLogger{}}
		engine.transformResourcesForFuse(test.runtime, test.alluxioValue)
		if result, found := test.alluxioValue.Fuse.Resources.Limits[corev1.ResourceMemory]; found {
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
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
	}{
		{&datav1alpha1.AlluxioRuntime{
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
		}, &Alluxio{
			Properties: map[string]string{},
			Master:     Master{},
		}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{Log: log.NullLogger{}}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "alluxio", test.runtime.Spec.TieredStore)
		engine.UnitTest = true
		engine.transformResourcesForFuse(test.runtime, test.alluxioValue)
		if test.alluxioValue.Fuse.Resources.Limits[corev1.ResourceMemory] != "22Gi" {
			t.Errorf("expected 22Gi, got %v", test.alluxioValue.Fuse.Resources.Limits[corev1.ResourceMemory])
		}
	}
}
