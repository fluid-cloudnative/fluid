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

package jindocache

import (
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestTransformToken(t *testing.T) {
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
			}}, &Jindo{Secret: "test"}, "secrets:///token/"},
	}
	for _, test := range tests {
		engine := &JindoCacheEngine{Log: fake.NullLogger()}
		engine.transformToken(test.runtime, test.jindoValue)
		if test.jindoValue.Master.TokenProperties["default.credential.provider"] != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Master.MasterProperties["default.credential.provider"])
		}
	}
}

func TestTransformMasterMountPath(t *testing.T) {
	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     *Level
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
			}}, &Jindo{}, &Level{
			Path:       "/mnt/disk1",
			Type:       string(common.VolumeTypeHostPath),
			MediumType: string(common.Memory),
		}},
	}
	for _, test := range tests {
		engine := &JindoCacheEngine{Log: fake.NullLogger()}
		properties := engine.transformMasterMountPath("/mnt/disk1", common.Memory, common.VolumeTypeHostPath)
		if !reflect.DeepEqual(properties["1"], test.expect) {
			t.Errorf("expected value %v, but got %v", test.expect, properties["1"])
		}
	}
}
