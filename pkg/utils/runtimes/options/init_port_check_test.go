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

package options

import (
	"testing"
)

func TestPortCheckEnabled(t *testing.T) {
	type testCase struct {
		name   string
		env    map[string]string
		expect bool
	}

	testCases := []testCase{
		{
			name:   "not_set",
			env:    map[string]string{},
			expect: false,
		}, {
			name: "set_true",
			env: map[string]string{
				EnvPortCheckEnabled: "true",
			},
			expect: true,
		}, {
			name: "set_false",
			env: map[string]string{
				EnvPortCheckEnabled: "false",
			},
			expect: false,
		},
	}

	for _, test := range testCases {
		for k, v := range test.env {
			t.Setenv(k, v)
		}
		setPortCheckOption()
		got := PortCheckEnabled()
		if got != test.expect {
			t.Errorf("testcase %s is failed due to expect %v, but got %v", test.name, test.expect, got)
		}
	}
}
