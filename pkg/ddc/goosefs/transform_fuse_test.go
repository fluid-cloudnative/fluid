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
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestTransformFuseWithNoArgs(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		dataset      *datav1alpha1.Dataset
		goosefsValue *GooseFS
		expect       string
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &GooseFS{}, "--fuse-opts=rw,direct_io,allow_other"},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{Log: fake.NullLogger()}
		err := engine.transformFuse(test.runtime, test.dataset, test.goosefsValue)
		if err != nil {
			t.Errorf("Got err %v", err)
		}
		if test.goosefsValue.Fuse.Args[1] != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.goosefsValue.Fuse.Args[1])
		}
	}
}

func TestTransformFuseWithArgs(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		dataset      *datav1alpha1.Dataset
		goosefsValue *GooseFS
		expect       string
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Fuse: datav1alpha1.GooseFSFuseSpec{
					Args: []string{
						"fuse",
						"--fuse-opts=kernel_cache",
					},
				},
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &GooseFS{}, "--fuse-opts=kernel_cache,allow_other"},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{Log: fake.NullLogger()}
		err := engine.transformFuse(test.runtime, test.dataset, test.goosefsValue)
		if err != nil {
			t.Errorf("Got err %v", err)
		}
		if test.goosefsValue.Fuse.Args[1] != test.expect {
			t.Errorf("expected fuse %v, but got %v", test.expect, test.goosefsValue.Fuse.Args[1])
		}
	}
}
