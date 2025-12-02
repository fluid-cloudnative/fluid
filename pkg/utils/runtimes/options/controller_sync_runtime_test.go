/*
Copyright 2025 The Fluid Authors.

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

import "testing"

func Test_setControllerSkipSyncingRuntime(t *testing.T) {
	tests := []struct {
		name   string
		env    map[string]string
		expect bool
	}{
		{
			name:   "not set",
			env:    map[string]string{},
			expect: false,
		},
		{
			name:   "set true",
			env:    map[string]string{EnvControllerSkipSyncingRuntime: "true"},
			expect: true,
		},
		{
			name:   "set false",
			env:    map[string]string{EnvControllerSkipSyncingRuntime: "false"},
			expect: false,
		},
	}
	for _, tt := range tests {
		for k, v := range tt.env {
			t.Setenv(k, v)
		}
		t.Run(tt.name, func(t *testing.T) {
			setControllerSkipSyncingRuntime()
			got := ShouldSkipSyncingRuntime()
			if got != tt.expect {
				t.Errorf("ControllerSkipSyncingRuntime() = %v, want %v", got, tt.expect)
			}
		})
	}
}
