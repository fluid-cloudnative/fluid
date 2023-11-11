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

package jindo

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestTransformFuseWithNoArgs(t *testing.T) {
	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     string
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Jindo{}, "true"},
	}
	for _, test := range tests {
		engine := &JindoEngine{Log: fake.NullLogger()}
		err := engine.transformFuse(test.runtime, test.jindoValue)
		if err != nil {
			t.Errorf("Got err %v", err)
		}
		if test.jindoValue.Fuse.FuseProperties["jfs.cache.data-cache.enable"] != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Fuse.FuseProperties["jfs.cache.data-cache.enable"])
		}
	}
}

func TestTransformRunAsUser(t *testing.T) {
	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     string
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				User: "user",
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Jindo{}, "user"},
	}
	for _, test := range tests {
		engine := &JindoEngine{Log: fake.NullLogger()}
		err := engine.transformRunAsUser(test.runtime, test.jindoValue)
		if err != nil {
			t.Errorf("Got err %v", err)
		}
		if test.jindoValue.Fuse.RunAs != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Fuse.RunAs)
		}
	}
}

func TestTransformSecret(t *testing.T) {
	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     string
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Secret: "secret",
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Jindo{}, "secret"},
	}
	for _, test := range tests {
		engine := &JindoEngine{Log: fake.NullLogger()}
		err := engine.transformSecret(test.runtime, test.jindoValue)
		if err != nil {
			t.Errorf("Got err %v", err)
		}
		if test.jindoValue.Secret != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Fuse.RunAs)
		}
	}
}

func TestTransformFuseArg(t *testing.T) {
	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     string
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Secret: "secret",
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
					Path:       "/",
				}},
			}}, &Jindo{}, "-ocredential_provider=secrets:///token/ -oroot_ns=jindo -okernel_cache"},
	}
	for _, test := range tests {
		engine := &JindoEngine{Log: fake.NullLogger()}
		properties := engine.transformFuseArg(test.runtime, test.dataset)
		if properties[0] != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Fuse.RunAs)
		}
	}
}

func TestParseFuseImage(t *testing.T) {
	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     string
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Secret: "secret",
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
					Path:       "/",
				}},
			}}, &Jindo{}, "registry.cn-shanghai.aliyuncs.com/jindofs/jindo-fuse:3.8.0"},
	}
	for _, test := range tests {
		engine := &JindoEngine{Log: fake.NullLogger()}
		imageR, tagR := engine.parseFuseImage()
		registryVersion := imageR + ":" + tagR
		if registryVersion != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Fuse.RunAs)
		}
	}
}
