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
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "thin", test.runtime.Spec.TieredStore)
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
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "thin", test.runtime.Spec.TieredStore)
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
