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

package cachefs

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestTransformResourcesForWorkerNoValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.CacheFSRuntime
		cachefsValue *CacheFS
	}{
		{&datav1alpha1.CacheFSRuntime{
			Spec: datav1alpha1.CacheFSRuntimeSpec{},
		}, &CacheFS{}},
	}
	for _, test := range tests {
		engine := &CacheFSEngine{Log: fake.NullLogger()}
		err := engine.transformResourcesForWorker(test.runtime, test.cachefsValue)
		if err != nil {
			t.Errorf("unexpected err %v", err)
		}
		if result, found := test.cachefsValue.Worker.Resources.Limits[corev1.ResourceMemory]; found {
			t.Errorf("expected nil, got %v", result)
		}
	}
}

func TestTransformResourcesForWorkerWithValue(t *testing.T) {
	resources := corev1.ResourceRequirements{}
	resources.Requests = make(corev1.ResourceList)
	resources.Requests[corev1.ResourceMemory] = resource.MustParse("2Gi")

	result := resource.MustParse("20Gi")

	var tests = []struct {
		runtime       *datav1alpha1.CacheFSRuntime
		cachefsValue  *CacheFS
		wantedRequest string
	}{
		{
			runtime: &datav1alpha1.CacheFSRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: datav1alpha1.CacheFSRuntimeSpec{
					Worker: datav1alpha1.CacheFSCompTemplateSpec{
						Resources: resources,
					},
				},
			},
			cachefsValue:  &CacheFS{},
			wantedRequest: "2Gi",
		},
		{
			runtime: &datav1alpha1.CacheFSRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test2",
				},
				Spec: datav1alpha1.CacheFSRuntimeSpec{
					Worker: datav1alpha1.CacheFSCompTemplateSpec{
						Resources: resources,
					},
					TieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{{
							MediumType: common.Memory,
							Quota:      &result,
						}},
					},
				},
			},
			cachefsValue: &CacheFS{
			},
			wantedRequest: "20Gi",
		},
	}
	for _, test := range tests {
		client := fake.NewFakeClientWithScheme(testScheme, test.runtime)
		engine := &CacheFSEngine{
			Log:    fake.NullLogger(),
			Client: client,
			name:   test.runtime.Name,
		}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "cachefs", test.runtime.Spec.TieredStore)
		engine.UnitTest = true
		err := engine.transformResourcesForWorker(test.runtime, test.cachefsValue)
		if err != nil {
			t.Error(err)
		}
		quantity := test.cachefsValue.Worker.Resources.Requests[corev1.ResourceMemory]
		if quantity != test.wantedRequest {
			t.Errorf("expected 22Gi, got %v", test.cachefsValue.Worker.Resources.Requests[corev1.ResourceMemory])
		}
	}
}

func TestTransformResourcesForFuseNoValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.CacheFSRuntime
		cachefsValue *CacheFS
	}{
		{&datav1alpha1.CacheFSRuntime{
			Spec: datav1alpha1.CacheFSRuntimeSpec{},
		}, &CacheFS{}},
	}
	for _, test := range tests {
		engine := &CacheFSEngine{Log: fake.NullLogger()}
		err := engine.transformResourcesForFuse(test.runtime, test.cachefsValue)
		if err != nil {
			t.Error(err)
		}
		if result, found := test.cachefsValue.Fuse.Resources.Limits[corev1.ResourceMemory]; found {
			t.Errorf("expected nil, got %v", result)
		}
	}
}

func TestTransformResourcesForFuseWithValue(t *testing.T) {

	resources := corev1.ResourceRequirements{}
	resources.Requests = make(corev1.ResourceList)
	resources.Requests[corev1.ResourceMemory] = resource.MustParse("2Gi")

	result := resource.MustParse("20Gi")

	var tests = []struct {
		runtime    *datav1alpha1.CacheFSRuntime
		cachefsValue *CacheFS
	}{
		{&datav1alpha1.CacheFSRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
			Spec: datav1alpha1.CacheFSRuntimeSpec{
				Fuse: datav1alpha1.CacheFSFuseSpec{
					Resources: resources,
				},
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
					}},
				},
			},
			Status: datav1alpha1.RuntimeStatus{},
		}, &CacheFS{}},
	}
	for _, test := range tests {
		client := fake.NewFakeClientWithScheme(testScheme, test.runtime)
		engine := &CacheFSEngine{
			Log:    fake.NullLogger(),
			Client: client,
			name:   test.runtime.Name,
		}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "cachefs", test.runtime.Spec.TieredStore)
		engine.UnitTest = true
		err := engine.transformResourcesForFuse(test.runtime, test.cachefsValue)
		if err != nil {
			t.Error(err)
		}
		quantity := test.cachefsValue.Fuse.Resources.Requests[corev1.ResourceMemory]
		if quantity != "20Gi" {
			t.Errorf("expected 22Gi, got %v", test.cachefsValue.Fuse.Resources.Limits[corev1.ResourceMemory])
		}
	}
}
