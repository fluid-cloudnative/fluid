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

package goosefs

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func TestTransformAPIGateway(t *testing.T) {
	var engine = &GooseFSEngine{}
	var tests = []struct {
		runtime *datav1alpha1.GooseFSRuntime
		value   *GooseFS
	}{
		{
			runtime: &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					APIGateway: datav1alpha1.GooseFSCompTemplateSpec{
						Enabled: true,
					},
				},
			},
			value: &GooseFS{
				APIGateway: APIGateway{
					Enabled: false,
				},
			},
		},
		{
			runtime: &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					APIGateway: datav1alpha1.GooseFSCompTemplateSpec{
						Enabled: false,
					},
				},
			},
			value: &GooseFS{
				APIGateway: APIGateway{
					Enabled: true,
				},
			},
		},
		{
			runtime: nil,
			value: &GooseFS{
				APIGateway: APIGateway{
					Enabled: false,
				},
			},
		},
		{
			runtime: &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					APIGateway: datav1alpha1.GooseFSCompTemplateSpec{
						Enabled: true,
					},
				},
			},
			value: nil,
		},
	}
	for _, test := range tests {
		err := engine.transformAPIGateway(test.runtime, test.value)
		if test.runtime == nil || test.value == nil {
			if err == nil {
				t.Errorf("should return err if it's possible to lead to nil pointer")
			}
		} else if test.runtime.Spec.APIGateway.Enabled != test.value.APIGateway.Enabled {
			t.Errorf("testcase cannot paas beacuse of wrong result,%t != %t",
				test.runtime.Spec.APIGateway.Enabled,
				test.value.APIGateway.Enabled)
		}
	}

}
