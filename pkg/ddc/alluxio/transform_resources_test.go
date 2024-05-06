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
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
				JobMaster: JobMaster{
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
				JobMaster: JobMaster{
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
			JobMaster: datav1alpha1.AlluxioCompTemplateSpec{
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
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				TieredStore: datav1alpha1.TieredStore{},
			},
		}, &Alluxio{
			Properties: map[string]string{},
		}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{
			Log:       fake.NullLogger(),
			name:      "test",
			namespace: "test",
		}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "alluxio", test.runtime.Spec.TieredStore)
		runtimeObjs := []runtime.Object{}
		runtimeObjs = append(runtimeObjs, test.runtime.DeepCopy())
		s := runtime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, test.runtime)
		_ = corev1.AddToScheme(s)
		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)
		engine.Client = client
		err := engine.transformResourcesForWorker(test.runtime, test.alluxioValue)
		if err != nil {
			t.Errorf("got err %v", err)
		}
		if result, found := test.alluxioValue.Worker.Resources.Limits[corev1.ResourceMemory]; found {
			t.Errorf("expected nil, got %v", result)
		}
	}
}

func TestTransformResourcesForWorkerWithTieredStore(t *testing.T) {
	result := resource.MustParse("20Gi")

	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
	}{
		{&datav1alpha1.AlluxioRuntime{
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
		}, &Alluxio{
			Properties: map[string]string{},
		}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{
			Log:       fake.NullLogger(),
			name:      "test",
			namespace: "test",
		}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "alluxio", test.runtime.Spec.TieredStore)
		runtimeObjs := []runtime.Object{}
		runtimeObjs = append(runtimeObjs, test.runtime.DeepCopy())
		s := runtime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, test.runtime)
		_ = corev1.AddToScheme(s)
		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)
		engine.Client = client
		err := engine.transformResourcesForWorker(test.runtime, test.alluxioValue)
		t.Log(err)
		if err != nil {
			t.Errorf("expected no err, got err %v", err)
		}
		if test.alluxioValue.Worker.Resources.Requests[corev1.ResourceMemory] != "20Gi" {
			t.Errorf("expected 20Gi, got %v", test.alluxioValue.Worker.Resources.Requests[corev1.ResourceMemory])
		}
		if result, found := test.alluxioValue.Worker.Resources.Limits[corev1.ResourceMemory]; found {
			t.Errorf("expected nil, got %v", result)
		}
	}
}

func TestTransformResourcesForWorkerWithValue(t *testing.T) {

	resources := corev1.ResourceRequirements{}
	resources.Limits = make(corev1.ResourceList)
	resources.Limits[corev1.ResourceMemory] = resource.MustParse("2Gi")
	resources.Limits[corev1.ResourceCPU] = resource.MustParse("500m")
	resources.Requests = make(corev1.ResourceList)
	resources.Requests[corev1.ResourceMemory] = resource.MustParse("1Gi")
	resources.Requests[corev1.ResourceCPU] = resource.MustParse("500m")

	resources1 := corev1.ResourceRequirements{}
	resources1.Limits = make(corev1.ResourceList)
	resources1.Limits[corev1.ResourceMemory] = resource.MustParse("20Gi")
	resources1.Limits[corev1.ResourceCPU] = resource.MustParse("500m")
	resources1.Requests = make(corev1.ResourceList)
	resources1.Requests[corev1.ResourceMemory] = resource.MustParse("1Gi")
	resources1.Requests[corev1.ResourceCPU] = resource.MustParse("500m")

	result := resource.MustParse("20Gi")

	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
		wantRes      []string
	}{
		{&datav1alpha1.AlluxioRuntime{
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
		}, &Alluxio{
			Properties: map[string]string{},
			Master:     Master{},
			JobMaster:  JobMaster{},
		}, []string{
			"err", "2Gi", "1Gi",
		}},
		{&datav1alpha1.AlluxioRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Worker: datav1alpha1.AlluxioCompTemplateSpec{
					Resources: resources1,
				},
				JobWorker: datav1alpha1.AlluxioCompTemplateSpec{
					Resources: resources1,
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
			JobMaster:  JobMaster{},
		}, []string{
			"nil", "20Gi", "20Gi",
		}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{
			Log:       fake.NullLogger(),
			name:      "test",
			namespace: "test",
		}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "alluxio", test.runtime.Spec.TieredStore)
		runtimeObjs := []runtime.Object{}
		runtimeObjs = append(runtimeObjs, test.runtime.DeepCopy())
		s := runtime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, test.runtime)
		_ = corev1.AddToScheme(s)
		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)
		engine.Client = client
		engine.UnitTest = true
		err := engine.transformResourcesForWorker(test.runtime, test.alluxioValue)
		if (err != nil && test.wantRes[0] != "err") || (err == nil && test.wantRes[0] != "nil") {
			t.Errorf("expected %v, got %v", test.wantRes[0], err)
		}
		if test.alluxioValue.Worker.Resources.Limits[corev1.ResourceMemory] != test.wantRes[1] {
			t.Errorf("expected %v, got %v", test.wantRes[1], test.alluxioValue.Worker.Resources.Limits[corev1.ResourceMemory])
		}
		if test.alluxioValue.Worker.Resources.Requests[corev1.ResourceMemory] != test.wantRes[2] {
			t.Errorf("expected %v, got %v", test.wantRes[2], test.alluxioValue.Worker.Resources.Requests[corev1.ResourceMemory])
		}
	}
}

func TestTransformResourcesForWorkerWithOnlyRequest(t *testing.T) {

	resources := corev1.ResourceRequirements{}
	resources.Requests = make(corev1.ResourceList)
	resources.Requests[corev1.ResourceMemory] = resource.MustParse("1Gi")
	resources.Requests[corev1.ResourceCPU] = resource.MustParse("500m")

	result := resource.MustParse("20Gi")

	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
		wantRes      string
	}{
		{&datav1alpha1.AlluxioRuntime{
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
		}, &Alluxio{
			Properties: map[string]string{},
			Master:     Master{},
			JobMaster:  JobMaster{},
		}, "20Gi"},
		{&datav1alpha1.AlluxioRuntime{
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
		}, &Alluxio{
			Properties: map[string]string{},
			Master:     Master{},
			JobMaster:  JobMaster{},
		}, "1Gi"},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{
			Log:       fake.NullLogger(),
			name:      "test",
			namespace: "test",
		}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "alluxio", test.runtime.Spec.TieredStore)
		engine.UnitTest = true
		runtimeObjs := []runtime.Object{}
		runtimeObjs = append(runtimeObjs, test.runtime.DeepCopy())
		s := runtime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, test.runtime)
		_ = corev1.AddToScheme(s)
		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)
		engine.Client = client
		err := engine.transformResourcesForWorker(test.runtime, test.alluxioValue)
		if err != nil {
			t.Errorf("expected nil, got err %v", err)
		}
		if result, found := test.alluxioValue.Worker.Resources.Limits[corev1.ResourceMemory]; found {
			t.Errorf("expected nil, got %v", result)
		}
		if test.alluxioValue.Worker.Resources.Requests[corev1.ResourceMemory] != test.wantRes {
			t.Errorf("expected %s, got %v", test.wantRes, test.alluxioValue.Worker.Resources.Requests[corev1.ResourceMemory])
		}
	}
}

func TestTransformResourcesForWorkerWithOnlyLimit(t *testing.T) {

	resources := corev1.ResourceRequirements{}
	resources.Limits = make(corev1.ResourceList)
	resources.Limits[corev1.ResourceMemory] = resource.MustParse("20Gi")
	resources.Limits[corev1.ResourceCPU] = resource.MustParse("500m")

	result := resource.MustParse("20Gi")

	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
		wantRes      []string
	}{
		{&datav1alpha1.AlluxioRuntime{
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
		}, &Alluxio{
			Properties: map[string]string{},
			Master:     Master{},
			JobMaster:  JobMaster{},
		}, []string{"20Gi", "20Gi"}},
		{&datav1alpha1.AlluxioRuntime{
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
		}, &Alluxio{
			Properties: map[string]string{},
			Master:     Master{},
			JobMaster:  JobMaster{},
		}, []string{"20Gi", "nil"}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{
			Log:       fake.NullLogger(),
			name:      "test",
			namespace: "test",
		}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "alluxio", test.runtime.Spec.TieredStore)
		engine.UnitTest = true
		runtimeObjs := []runtime.Object{}
		runtimeObjs = append(runtimeObjs, test.runtime.DeepCopy())
		s := runtime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, test.runtime)
		_ = corev1.AddToScheme(s)
		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)
		engine.Client = client
		err := engine.transformResourcesForWorker(test.runtime, test.alluxioValue)
		if err != nil {
			t.Errorf("expected nil, got err %v", err)
		}
		if test.alluxioValue.Worker.Resources.Limits[corev1.ResourceMemory] != test.wantRes[0] {
			t.Errorf("expected %s, got %v", test.wantRes[0], test.alluxioValue.Worker.Resources.Limits[corev1.ResourceMemory])
		}
		if result, found := test.alluxioValue.Worker.Resources.Requests[corev1.ResourceMemory]; (found && result != test.wantRes[1]) || (!found && test.wantRes[1] != "nil") {
			t.Errorf("expected %s, got %v", test.wantRes[1], test.alluxioValue.Worker.Resources.Requests[corev1.ResourceMemory])
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
		engine := &AlluxioEngine{Log: fake.NullLogger()}
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
		engine := &AlluxioEngine{Log: fake.NullLogger()}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "alluxio", test.runtime.Spec.TieredStore)
		engine.UnitTest = true
		engine.transformResourcesForFuse(test.runtime, test.alluxioValue)
		if test.alluxioValue.Fuse.Resources.Limits[corev1.ResourceMemory] != "22Gi" {
			t.Errorf("expected 22Gi, got %v", test.alluxioValue.Fuse.Resources.Limits[corev1.ResourceMemory])
		}
	}
}
