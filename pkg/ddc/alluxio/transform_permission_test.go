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

func TestTransformPermission(t *testing.T) {

	keys := []string{
		"alluxio.master.security.impersonation.root.users",
		"alluxio.master.security.impersonation.root.groups",
		"alluxio.security.authorization.permission.enabled",
	}

	var tests = []struct {
		runtime *datav1alpha1.AlluxioRuntime
		value   *Alluxio
		expect  map[string]string
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Fuse: datav1alpha1.AlluxioFuseSpec{},
			},
		}, &Alluxio{}, map[string]string{
			"alluxio.master.security.impersonation.root.users":  "*",
			"alluxio.master.security.impersonation.root.groups": "*",
			"alluxio.security.authorization.permission.enabled": "false",
		}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{}
		engine.transformPermission(test.runtime, test.value)
		for _, key := range keys {
			if test.value.Properties[key] != test.expect[key] {
				t.Errorf("The key %s expected %s, got %s", key, test.value.Properties[key], test.expect[key])
			}
		}

	}
}
