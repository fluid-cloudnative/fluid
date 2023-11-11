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
