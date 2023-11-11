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

package alluxio

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func TestTransformAPIGateway(t *testing.T) {
	var engine = &AlluxioEngine{}
	var tests = []struct {
		runtime *datav1alpha1.AlluxioRuntime
		value   *Alluxio
	}{
		{
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					APIGateway: datav1alpha1.AlluxioCompTemplateSpec{
						Enabled: true,
					},
				},
			},
			value: &Alluxio{
				APIGateway: APIGateway{
					Enabled: false,
				},
			},
		},
		{
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					APIGateway: datav1alpha1.AlluxioCompTemplateSpec{
						Enabled: false,
					},
				},
			},
			value: &Alluxio{
				APIGateway: APIGateway{
					Enabled: true,
				},
			},
		},
		{
			runtime: nil,
			value: &Alluxio{
				APIGateway: APIGateway{
					Enabled: false,
				},
			},
		},
		{
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					APIGateway: datav1alpha1.AlluxioCompTemplateSpec{
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
