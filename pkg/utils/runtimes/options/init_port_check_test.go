/*
Copyright 2021 The Fluid Authors.

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
