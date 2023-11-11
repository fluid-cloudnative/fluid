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

package jindofsx

import (
	"reflect"
	"strings"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func TestTransformWorker(t *testing.T) {
	resources := corev1.ResourceRequirements{}
	resources.Limits = make(corev1.ResourceList)
	resources.Limits[corev1.ResourceMemory] = resource.MustParse("2Gi")

	result := resource.MustParse("20Gi")
	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     string
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Secret: "secret",
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
						High:       "0.8",
						Low:        "0.1",
					}},
				},
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Jindo{}, "1G"},
	}
	for _, test := range tests {
		engine := &JindoFSxEngine{Log: fake.NullLogger()}
		test.jindoValue.Worker.Port.Rpc = 8001
		test.jindoValue.Worker.Port.Raft = 8002
		dataPath := "/var/lib/docker/data"
		userQuotas := "1G"
		engine.transformWorker(test.runtime, dataPath, userQuotas, test.jindoValue)
		if test.jindoValue.Worker.WorkerProperties["storage.data-dirs.capacities"] != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Worker.WorkerProperties["storage.data-dirs.capacities"])
		}
	}
}

func TestTransformWorkerMountPath(t *testing.T) {
	var tests = []struct {
		storagePath                string
		quotaList                  string
		tieredStoreLevelMediumType string
		tieredStoreLevelVolumeType common.VolumeType
		tieredStoreLevels          []datav1alpha1.Level
		expect                     map[string]*Level
	}{
		{
			storagePath:                "/mnt/disk1,/mnt/disk2",
			quotaList:                  "10Gi,5Gi",
			tieredStoreLevelMediumType: string(common.SSD),
			tieredStoreLevelVolumeType: common.VolumeTypeHostPath,
			tieredStoreLevels:          []datav1alpha1.Level{},
			expect: map[string]*Level{
				"1": {
					Path:       "/mnt/disk1",
					Type:       string(common.VolumeTypeHostPath),
					MediumType: string(common.SSD),
					Quota:      "10Gi",
				},
				"2": {
					Path:       "/mnt/disk2",
					Type:       string(common.VolumeTypeHostPath),
					MediumType: string(common.SSD),
					Quota:      "5Gi",
				},
			},
		},
		{
			storagePath:                "/dev/shm",
			quotaList:                  "20Gi",
			tieredStoreLevelMediumType: string(common.Memory),
			tieredStoreLevelVolumeType: common.VolumeTypeEmptyDir,
			tieredStoreLevels: []datav1alpha1.Level{
				{
					VolumeType:   common.VolumeTypeEmptyDir,
					VolumeSource: datav1alpha1.VolumeSource{VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{Medium: corev1.StorageMedium("LocalSSD")}}},
				},
			},
			expect: map[string]*Level{
				"1": {
					Path:       "/dev/shm",
					Type:       string(common.VolumeTypeEmptyDir),
					MediumType: "LocalSSD",
					Quota:      "20Gi",
				},
			},
		},
	}
	for _, test := range tests {
		engine := &JindoFSxEngine{Log: fake.NullLogger()}
		originPath := strings.Split(test.storagePath, ",")
		quotas := strings.Split(test.quotaList, ",")

		properties := engine.transformWorkerMountPath(originPath, quotas, engine.getMediumTypeFromVolumeSource(test.tieredStoreLevelMediumType, test.tieredStoreLevels), test.tieredStoreLevelVolumeType)
		if !reflect.DeepEqual(properties, test.expect) {
			t.Errorf("expected value %v, but got %v", test.expect, properties)
		}
	}
}

func TestTransformResourcesForWorkerNoValue(t *testing.T) {
	quotas := []resource.Quantity{resource.MustParse("10Gi")}
	var tests = []struct {
		name          string
		namespace     string
		size          string
		runtime       *datav1alpha1.JindoRuntime
		jindoValue    *Jindo
		wantResources Resources
	}{
		{
			name:      "test",
			namespace: "default",
			size:      "10g",
			runtime: &datav1alpha1.JindoRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: datav1alpha1.JindoRuntimeSpec{
					TieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{{
							MediumType: common.Memory,
							Quota:      &quotas[0],
							High:       "0.8",
							Low:        "0.1",
						}},
					},
				},
			},
			jindoValue: &Jindo{
				Properties: map[string]string{},
			}, wantResources: Resources{
				Requests: Resource{
					CPU:    "",
					Memory: "10Gi",
				},
			}}, {
			name:      "noTieredStore",
			namespace: "default",
			size:      "0g",
			runtime: &datav1alpha1.JindoRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: datav1alpha1.JindoRuntimeSpec{},
			},
			jindoValue: &Jindo{
				Properties: map[string]string{},
			}, wantResources: Resources{
				Requests: Resource{},
			}},
	}
	for _, test := range tests {
		// engine := &JindoFSxEngine{Log: fake.NullLogger()}

		runtimeObjs := []runtime.Object{}
		runtimeObjs = append(runtimeObjs, test.runtime.DeepCopy())
		s := runtime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, test.runtime)
		_ = corev1.AddToScheme(s)
		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)
		engine := &JindoFSxEngine{
			name:      test.name,
			namespace: test.namespace,
			Client:    client,
			Log:       fake.NullLogger(),
		}
		err := engine.transformResources(test.runtime, test.jindoValue, test.size)
		if err != nil {
			t.Errorf("got error %v", err)
		}
		if test.jindoValue.Worker.Resources.Requests.Memory != test.wantResources.Requests.Memory {
			t.Errorf("expected %v, got %v",
				test.wantResources.Requests.Memory,
				test.jindoValue.Worker.Resources.Requests.Memory)
		}
		if test.jindoValue.Worker.Resources.Requests.CPU != test.wantResources.Requests.CPU {
			t.Errorf("expected %v, got %v",
				test.wantResources.Requests.CPU,
				test.jindoValue.Worker.Resources.Requests.CPU)
		}
		if test.jindoValue.Worker.Resources.Limits.Memory != test.wantResources.Limits.Memory {
			t.Errorf("expected %v, got %v",
				test.wantResources.Limits.Memory,
				test.jindoValue.Worker.Resources.Limits.Memory)
		}
		if test.jindoValue.Worker.Resources.Limits.CPU != test.wantResources.Limits.CPU {
			t.Errorf("expected %v, got %v", test.wantResources.Limits.CPU, test.jindoValue.Worker.Resources.Limits.CPU)
		}
	}
}

func TestTransformResourcesForWorkerWithValue(t *testing.T) {

	resources := corev1.ResourceRequirements{}
	resources.Limits = make(corev1.ResourceList)
	resources.Limits[corev1.ResourceMemory] = resource.MustParse("2Gi")
	resources.Limits[corev1.ResourceCPU] = resource.MustParse("2")
	resources.Requests = make(corev1.ResourceList)
	resources.Requests[corev1.ResourceMemory] = resource.MustParse("1Gi")
	resources.Requests[corev1.ResourceCPU] = resource.MustParse("1")
	result := resource.MustParse("20Gi")

	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		jindoValue *Jindo
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Worker: datav1alpha1.JindoCompTemplateSpec{
					Resources: resources,
				},
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.SSD,
						Quota:      &result,
					}},
				},
			},
		}, &Jindo{
			Properties: map[string]string{},
			Master:     Master{},
		}},
	}
	for _, test := range tests {
		engine := &JindoFSxEngine{Log: fake.NullLogger()}
		err := engine.transformResources(test.runtime, test.jindoValue, "10g")
		if err != nil {
			t.Errorf("got error %v", err)
		}
		if test.jindoValue.Worker.Resources.Requests.Memory != "1Gi" {
			t.Errorf("expected nil, got %v", test.jindoValue.Worker.Resources.Requests.Memory)
		}
		if test.jindoValue.Worker.Resources.Requests.CPU != "1" {
			t.Errorf("expected nil, got %v", test.jindoValue.Worker.Resources.Requests.CPU)
		}
		if test.jindoValue.Worker.Resources.Limits.Memory != "2Gi" {
			t.Errorf("expected nil, got %v", test.jindoValue.Worker.Resources.Limits.Memory)
		}
		if test.jindoValue.Worker.Resources.Limits.CPU != "2" {
			t.Errorf("expected nil, got %v", test.jindoValue.Worker.Resources.Limits.CPU)
		}
	}
}
