/*
Copyright 2022 The Fluid Authors.

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

package jindofsx

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
		engine := &JindoFSxEngine{Log: fake.NullLogger()}
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
		engine := &JindoFSxEngine{Log: fake.NullLogger()}
		properties := engine.transformMasterMountPath("/mnt/disk1", common.Memory, common.VolumeTypeHostPath)
		if !reflect.DeepEqual(properties["1"], test.expect) {
			t.Errorf("expected value %v, but got %v", test.expect, properties["1"])
		}
	}
}
