/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package goosefs

import (
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestTransformResourcesForMaster(t *testing.T) {
	testCases := map[string]struct {
		runtime *datav1alpha1.GooseFSRuntime
		got     *GooseFS
		want    *GooseFS
	}{
		"test goosefs master pass through resources with limits and request case 1": {
			runtime: mockGooseFSRuntimeForMaster(
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
			got: &GooseFS{},
			want: &GooseFS{
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
		"test GooseFS master pass through resources with request case 1": {
			runtime: mockGooseFSRuntimeForMaster(
				corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("100Mi"),
					},
				},
			),
			got: &GooseFS{},
			want: &GooseFS{
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
		"test goosefs master pass through resources without request and limit case 1": {
			runtime: mockGooseFSRuntimeForMaster(
				corev1.ResourceRequirements{
					Requests: corev1.ResourceList{},
				},
			),
			got:  &GooseFS{},
			want: &GooseFS{},
		},
		"test goosefs master pass through resources without request and limit case 2": {
			runtime: mockGooseFSRuntimeForMaster(corev1.ResourceRequirements{}),
			got:     &GooseFS{},
			want:    &GooseFS{},
		},
		"test goosefs master pass through resources without request and limit case 3": {
			runtime: mockGooseFSRuntimeForMaster(
				corev1.ResourceRequirements{
					Limits: corev1.ResourceList{},
				},
			),
			got:  &GooseFS{},
			want: &GooseFS{},
		},
	}

	engine := &GooseFSEngine{}
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

func mockGooseFSRuntimeForMaster(res corev1.ResourceRequirements) *datav1alpha1.GooseFSRuntime {
	runtime := &datav1alpha1.GooseFSRuntime{
		Spec: datav1alpha1.GooseFSRuntimeSpec{
			Master: datav1alpha1.GooseFSCompTemplateSpec{
				Resources: res,
			},
			JobMaster: datav1alpha1.GooseFSCompTemplateSpec{
				Resources: res,
			},
		},
	}
	return runtime

}

func TestTransformResourcesForWorkerNoValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{},
		}, &GooseFS{
			Properties: map[string]string{},
		}},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{Log: fake.NullLogger()}
		engine.transformResourcesForWorker(test.runtime, test.goosefsValue)
		if result, found := test.goosefsValue.Worker.Resources.Limits[corev1.ResourceMemory]; found {
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

	result := resource.MustParse("20Gi")

	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Worker: datav1alpha1.GooseFSCompTemplateSpec{
					Resources: resources,
				},
				JobWorker: datav1alpha1.GooseFSCompTemplateSpec{
					Resources: resources,
				},
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
					}},
				},
			},
		}, &GooseFS{
			Properties: map[string]string{},
			Master:     Master{},
		}},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{Log: fake.NullLogger()}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "goosefs", test.runtime.Spec.TieredStore)
		engine.UnitTest = true
		engine.transformResourcesForWorker(test.runtime, test.goosefsValue)
		if test.goosefsValue.Worker.Resources.Limits[corev1.ResourceMemory] != "22Gi" {
			t.Errorf("expected 22Gi, got %v", test.goosefsValue.Worker.Resources.Limits[corev1.ResourceMemory])
		}
	}
}

func TestTransformResourcesForFuseNoValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{},
		}, &GooseFS{
			Properties: map[string]string{},
		}},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{Log: fake.NullLogger()}
		engine.transformResourcesForFuse(test.runtime, test.goosefsValue)
		if result, found := test.goosefsValue.Fuse.Resources.Limits[corev1.ResourceMemory]; found {
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
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Fuse: datav1alpha1.GooseFSFuseSpec{
					Resources: resources,
				},
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
					}},
				},
			},
		}, &GooseFS{
			Properties: map[string]string{},
			Master:     Master{},
			JobMaster:  JobMaster{},
		}},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{Log: fake.NullLogger()}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "goosefs", test.runtime.Spec.TieredStore)
		engine.UnitTest = true
		engine.transformResourcesForFuse(test.runtime, test.goosefsValue)
		if test.goosefsValue.Fuse.Resources.Limits[corev1.ResourceMemory] != "22Gi" {
			t.Errorf("expected 22Gi, got %v", test.goosefsValue.Fuse.Resources.Limits[corev1.ResourceMemory])
		}
	}
}
