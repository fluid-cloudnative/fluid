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

package vineyard

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
		runtime *datav1alpha1.VineyardRuntime
		got     *Vineyard
		want    *Vineyard
	}{
		"test vineyard master pass through resources with limits and request case 1": {
			runtime: mockVineyardRuntimeForMaster(
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
			got: &Vineyard{},
			want: &Vineyard{
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
		"test vineyard master pass through resources with request case 1": {
			runtime: mockVineyardRuntimeForMaster(
				corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("100Mi"),
					},
				},
			),
			got: &Vineyard{},
			want: &Vineyard{
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
		"test vineyard master pass through resources without request and limit case 1": {
			runtime: mockVineyardRuntimeForMaster(
				corev1.ResourceRequirements{
					Requests: corev1.ResourceList{},
				},
			),
			got:  &Vineyard{},
			want: &Vineyard{},
		},
		"test vineyard master pass through resources without request and limit case 2": {
			runtime: mockVineyardRuntimeForMaster(corev1.ResourceRequirements{}),
			got:     &Vineyard{},
			want:    &Vineyard{},
		},
		"test vineyard master pass through resources without request and limit case 3": {
			runtime: mockVineyardRuntimeForMaster(
				corev1.ResourceRequirements{
					Limits: corev1.ResourceList{},
				},
			),
			got:  &Vineyard{},
			want: &Vineyard{},
		},
	}

	engine := &VineyardEngine{}
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

func mockVineyardRuntimeForMaster(res corev1.ResourceRequirements) *datav1alpha1.VineyardRuntime {
	runtime := &datav1alpha1.VineyardRuntime{
		Spec: datav1alpha1.VineyardRuntimeSpec{
			Master: datav1alpha1.MasterSpec{
				VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
					Resources: res,
				},
			},
		},
	}
	return runtime

}

func TestTransformResourcesForWorkerNoValue(t *testing.T) {
	var tests = []struct {
		runtime       *datav1alpha1.VineyardRuntime
		vineyardValue *Vineyard
	}{
		{&datav1alpha1.VineyardRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: datav1alpha1.VineyardRuntimeSpec{
				TieredStore: datav1alpha1.TieredStore{},
			},
		}, &Vineyard{}},
	}
	for _, test := range tests {
		engine := &VineyardEngine{
			Log:       fake.NullLogger(),
			name:      "test",
			namespace: "test",
		}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "vineyard", test.runtime.Spec.TieredStore)
		runtimeObjs := []runtime.Object{}
		runtimeObjs = append(runtimeObjs, test.runtime.DeepCopy())
		s := runtime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, test.runtime)
		_ = corev1.AddToScheme(s)
		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)
		engine.Client = client
		err := engine.transformResourcesForWorker(test.runtime, test.vineyardValue)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	}
}

func TestTransformResourcesForWorkerWithTieredStore(t *testing.T) {
	result := resource.MustParse("20Gi")
	var tests = []struct {
		runtime       *datav1alpha1.VineyardRuntime
		vineyardValue *Vineyard
	}{
		{&datav1alpha1.VineyardRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
			Spec: datav1alpha1.VineyardRuntimeSpec{
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
					}},
				},
			},
		}, &Vineyard{}},
	}
	for _, test := range tests {
		engine := &VineyardEngine{
			Log:       fake.NullLogger(),
			name:      "test",
			namespace: "test",
		}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "vineyard", test.runtime.Spec.TieredStore)
		runtimeObjs := []runtime.Object{}
		runtimeObjs = append(runtimeObjs, test.runtime.DeepCopy())
		s := runtime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, test.runtime)
		_ = corev1.AddToScheme(s)
		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)
		engine.Client = client
		err := engine.transformResourcesForWorker(test.runtime, test.vineyardValue)
		t.Log(err)
		if err != nil {
			t.Errorf("expected no err, got err %v", err)
		}
		// 20Gi + 500Mi = 20980Mi
		if test.vineyardValue.Worker.Resources.Requests[corev1.ResourceMemory] != "20980Mi" {
			t.Errorf("expected 20980Mi, got %v", test.vineyardValue.Worker.Resources.Requests[corev1.ResourceMemory])
		}
		if test.vineyardValue.Worker.Resources.Limits[corev1.ResourceMemory] != "20980Mi" {
			t.Errorf("expected 20980Mi, got %v", test.vineyardValue.Worker.Resources.Limits[corev1.ResourceMemory])
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
		runtime       *datav1alpha1.VineyardRuntime
		vineyardValue *Vineyard
		wantRes       []string
	}{
		{&datav1alpha1.VineyardRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
			Spec: datav1alpha1.VineyardRuntimeSpec{
				Worker: datav1alpha1.VineyardCompTemplateSpec{
					Resources: resources,
				},
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
					}},
				},
			},
		}, &Vineyard{
			Master: Master{},
		}, []string{
			"20980Mi", "20980Mi",
		}},
		{&datav1alpha1.VineyardRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
			Spec: datav1alpha1.VineyardRuntimeSpec{
				Worker: datav1alpha1.VineyardCompTemplateSpec{
					Resources: resources1,
				},
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
					}},
				},
			},
		}, &Vineyard{
			Master: Master{},
		}, []string{
			"20980Mi", "20980Mi",
		}},
	}
	for _, test := range tests {
		engine := &VineyardEngine{
			Log:       fake.NullLogger(),
			name:      "test",
			namespace: "test",
		}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "vineyard", test.runtime.Spec.TieredStore)
		runtimeObjs := []runtime.Object{}
		runtimeObjs = append(runtimeObjs, test.runtime.DeepCopy())
		s := runtime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, test.runtime)
		_ = corev1.AddToScheme(s)
		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)
		engine.Client = client
		err := engine.transformResourcesForWorker(test.runtime, test.vineyardValue)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
		if test.vineyardValue.Worker.Resources.Limits[corev1.ResourceMemory] != test.wantRes[0] {
			t.Errorf("expected %v, got %v", test.wantRes[0], test.vineyardValue.Worker.Resources.Limits[corev1.ResourceMemory])
		}
		if test.vineyardValue.Worker.Resources.Requests[corev1.ResourceMemory] != test.wantRes[1] {
			t.Errorf("expected %v, got %v", test.wantRes[1], test.vineyardValue.Worker.Resources.Requests[corev1.ResourceMemory])
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
		runtime       *datav1alpha1.VineyardRuntime
		vineyardValue *Vineyard
		wantRes       string
	}{
		{&datav1alpha1.VineyardRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
			Spec: datav1alpha1.VineyardRuntimeSpec{
				Worker: datav1alpha1.VineyardCompTemplateSpec{
					Resources: resources,
				},
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
					}},
				},
			},
		}, &Vineyard{
			Master: Master{},
		}, "20980Mi"},
		{&datav1alpha1.VineyardRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
			Spec: datav1alpha1.VineyardRuntimeSpec{
				Worker: datav1alpha1.VineyardCompTemplateSpec{
					Resources: resources,
				},
				TieredStore: datav1alpha1.TieredStore{},
			},
		}, &Vineyard{
			Master: Master{},
		}, "1Gi"},
	}
	for _, test := range tests {
		engine := &VineyardEngine{
			Log:       fake.NullLogger(),
			name:      "test",
			namespace: "test",
		}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "vineyard", test.runtime.Spec.TieredStore)
		runtimeObjs := []runtime.Object{}
		runtimeObjs = append(runtimeObjs, test.runtime.DeepCopy())
		s := runtime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, test.runtime)
		_ = corev1.AddToScheme(s)
		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)
		engine.Client = client
		err := engine.transformResourcesForWorker(test.runtime, test.vineyardValue)
		if len(test.runtime.Spec.TieredStore.Levels) == 0 && err == nil {
			t.Errorf("expected error, got nil")
		}
		if len(test.runtime.Spec.TieredStore.Levels) != 0 && err != nil {
			t.Errorf("expected nil, got %v", err)
		}

		if len(test.runtime.Spec.TieredStore.Levels) == 0 {
			if result, found := test.vineyardValue.Worker.Resources.Requests[corev1.ResourceMemory]; found {
				t.Errorf("expected nil, got %v", result)
			}
			if result, found := test.vineyardValue.Worker.Resources.Limits[corev1.ResourceMemory]; found {
				t.Errorf("expected nil, got %v", result)
			}
		} else {
			if test.vineyardValue.Worker.Resources.Limits[corev1.ResourceMemory] != test.wantRes {
				t.Errorf("expected %s, got %v", test.wantRes, test.vineyardValue.Worker.Resources.Limits[corev1.ResourceMemory])
			}
			if test.vineyardValue.Worker.Resources.Requests[corev1.ResourceMemory] != test.wantRes {
				t.Errorf("expected %s, got %v", test.wantRes, test.vineyardValue.Worker.Resources.Requests[corev1.ResourceMemory])
			}
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
		runtime       *datav1alpha1.VineyardRuntime
		vineyardValue *Vineyard
		wantRes       []string
	}{
		{&datav1alpha1.VineyardRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
			Spec: datav1alpha1.VineyardRuntimeSpec{
				Worker: datav1alpha1.VineyardCompTemplateSpec{
					Resources: resources,
				},
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
					}},
				},
			},
		}, &Vineyard{
			Master: Master{},
		}, []string{"20980Mi", "20980Mi"}},
		{&datav1alpha1.VineyardRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
			Spec: datav1alpha1.VineyardRuntimeSpec{
				Worker: datav1alpha1.VineyardCompTemplateSpec{
					Resources: resources,
				},
				TieredStore: datav1alpha1.TieredStore{},
			},
		}, &Vineyard{
			Master: Master{},
		}, []string{"20980Mi", "20980Mi"}},
	}
	for _, test := range tests {
		engine := &VineyardEngine{
			Log:       fake.NullLogger(),
			name:      "test",
			namespace: "test",
		}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "vineyard", test.runtime.Spec.TieredStore)
		runtimeObjs := []runtime.Object{}
		runtimeObjs = append(runtimeObjs, test.runtime.DeepCopy())
		s := runtime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, test.runtime)
		_ = corev1.AddToScheme(s)
		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)
		engine.Client = client
		err := engine.transformResourcesForWorker(test.runtime, test.vineyardValue)
		if len(test.runtime.Spec.TieredStore.Levels) == 0 && err == nil {
			t.Errorf("expected error, got nil")
		}
		if len(test.runtime.Spec.TieredStore.Levels) != 0 && err != nil {
			t.Errorf("expected nil, got %v", err)
		}
		if len(test.runtime.Spec.TieredStore.Levels) == 0 {
			if result, found := test.vineyardValue.Worker.Resources.Requests[corev1.ResourceMemory]; found {
				t.Errorf("expected nil, got %v", result)
			}
			if result, found := test.vineyardValue.Worker.Resources.Limits[corev1.ResourceMemory]; found {
				t.Errorf("expected nil, got %v", result)
			}
		} else {
			if test.vineyardValue.Worker.Resources.Limits[corev1.ResourceMemory] != test.wantRes[0] {
				t.Errorf("expected %s, got %v", test.wantRes[0], test.vineyardValue.Worker.Resources.Limits[corev1.ResourceMemory])
			}
			if test.vineyardValue.Worker.Resources.Requests[corev1.ResourceMemory] != test.wantRes[0] {
				t.Errorf("expected %s, got %v", test.wantRes[0], test.vineyardValue.Worker.Resources.Limits[corev1.ResourceMemory])
			}
		}
	}
}

func TestTransformResourcesForFuseNoValue(t *testing.T) {
	var tests = []struct {
		runtime       *datav1alpha1.VineyardRuntime
		vineyardValue *Vineyard
	}{
		{&datav1alpha1.VineyardRuntime{
			Spec: datav1alpha1.VineyardRuntimeSpec{},
		}, &Vineyard{}},
	}
	for _, test := range tests {
		engine := &VineyardEngine{Log: fake.NullLogger()}
		engine.transformResourcesForFuse(test.runtime, test.vineyardValue)
		if result, found := test.vineyardValue.Fuse.Resources.Limits[corev1.ResourceMemory]; found {
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
		runtime       *datav1alpha1.VineyardRuntime
		vineyardValue *Vineyard
	}{
		{&datav1alpha1.VineyardRuntime{
			Spec: datav1alpha1.VineyardRuntimeSpec{
				Fuse: datav1alpha1.VineyardSockSpec{
					Resources: resources,
				},
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
					}},
				},
			},
		}, &Vineyard{
			Master: Master{},
		}},
	}
	for _, test := range tests {
		engine := &VineyardEngine{Log: fake.NullLogger()}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "vineyard", test.runtime.Spec.TieredStore)
		engine.transformResourcesForFuse(test.runtime, test.vineyardValue)
		if test.vineyardValue.Fuse.Resources.Limits[corev1.ResourceMemory] != "2Gi" {
			t.Errorf("expected 2Gi, got %v", test.vineyardValue.Fuse.Resources.Limits[corev1.ResourceMemory])
		}
	}
}
