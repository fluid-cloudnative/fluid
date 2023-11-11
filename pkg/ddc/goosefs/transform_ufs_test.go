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
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

func TestTransformDatasetToVolume(t *testing.T) {
	var ufsPath = UFSPath{}
	ufsPath.Name = "test"
	ufsPath.HostPath = "/mnt/test"
	ufsPath.ContainerPath = "/underFSStorage/test"

	var ufsPath1 = UFSPath{}
	ufsPath1.Name = "test"
	ufsPath1.HostPath = "/mnt/test"
	ufsPath1.ContainerPath = "/underFSStorage"

	var tests = []struct {
		runtime *datav1alpha1.GooseFSRuntime
		dataset *datav1alpha1.Dataset
		value   *GooseFS
		expect  UFSPath
	}{
		{&datav1alpha1.GooseFSRuntime{}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			},
		}, &GooseFS{}, ufsPath},
		{&datav1alpha1.GooseFSRuntime{}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
					Path:       "/",
				}},
			},
		}, &GooseFS{}, ufsPath1},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{}
		engine.transformDatasetToVolume(test.runtime, test.dataset, test.value)
		if test.value.UFSPaths[0].HostPath != test.expect.HostPath ||
			test.value.UFSPaths[0].ContainerPath != test.expect.ContainerPath {
			t.Errorf("expected %v, got %v", test.expect, test.value.UFSPaths[0])
		}
	}
}

func TestTransformDatasetToPVC(t *testing.T) {
	var ufsVolume = UFSVolume{}
	ufsVolume.Name = "test"
	ufsVolume.ContainerPath = "/underFSStorage/test"

	var ufsVolume1 = UFSVolume{}
	ufsVolume1.Name = "test1"
	ufsVolume1.ContainerPath = "/underFSStorage"

	var ufsVolume2 = UFSVolume{}
	ufsVolume2.Name = "test2"
	ufsVolume2.SubPath = "subpath"
	ufsVolume2.ContainerPath = "/underFSStorage/test2"

	var ufsVolume3 = UFSVolume{}
	ufsVolume3.Name = "test3"
	ufsVolume3.SubPath = "subpath"
	ufsVolume3.ContainerPath = "/underFSStorage"

	var tests = []struct {
		runtime *datav1alpha1.GooseFSRuntime
		dataset *datav1alpha1.Dataset
		value   *GooseFS
		expect  UFSVolume
	}{
		{&datav1alpha1.GooseFSRuntime{}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "pvc://test",
					Name:       "test",
				}},
			},
		}, &GooseFS{}, ufsVolume},
		{&datav1alpha1.GooseFSRuntime{}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "pvc://test1",
					Name:       "test1",
					Path:       "/",
				}},
			},
		}, &GooseFS{}, ufsVolume1},
		{&datav1alpha1.GooseFSRuntime{}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "pvc://test2/subpath",
					Name:       "test2",
				}},
			},
		}, &GooseFS{}, ufsVolume2},
		{&datav1alpha1.GooseFSRuntime{}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "pvc://test3/subpath",
					Name:       "test3",
					Path:       "/",
				}},
			},
		}, &GooseFS{}, ufsVolume3},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{}
		engine.transformDatasetToVolume(test.runtime, test.dataset, test.value)
		if test.value.UFSVolumes[0].ContainerPath != test.expect.ContainerPath ||
			test.value.UFSVolumes[0].Name != test.expect.Name ||
			test.value.UFSVolumes[0].SubPath != test.expect.SubPath {
			t.Errorf("expected %v, got %v", test.expect, test.value)
		}
	}
}

func TestTransformDatasetWithAffinity(t *testing.T) {
	var ufsPath = UFSPath{}
	ufsPath.Name = "test"
	ufsPath.HostPath = "/mnt/test"
	ufsPath.ContainerPath = "/opt/goosefs/underFSStorage/test"

	var tests = []struct {
		runtime *datav1alpha1.GooseFSRuntime
		dataset *datav1alpha1.Dataset
		value   *GooseFS
		expect  UFSPath
	}{
		{&datav1alpha1.GooseFSRuntime{}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
				NodeAffinity: &datav1alpha1.CacheableNodeAffinity{
					Required: &v1.NodeSelector{
						NodeSelectorTerms: []v1.NodeSelectorTerm{
							{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Operator: v1.NodeSelectorOpIn,
										Values:   []string{"test-label-value"},
									},
								},
							},
						},
					},
				},
			},
		}, &GooseFS{}, ufsPath},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{}
		engine.transformDatasetToVolume(test.runtime, test.dataset, test.value)
		if test.value.Master.Affinity.NodeAffinity == nil {
			t.Error("The master affinity is nil")
		}
	}
}
