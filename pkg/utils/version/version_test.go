/*
Copyright 2023 The Fluid Authors.

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

package version

import "testing"

func TestParse(t *testing.T) {
	validVersions := []string{
		"2.7.2-SNAPSHOT-3714f2b",
		"release-2.7.2-SNAPSHOT-3714f2b",
		"2.8.0",
	}

	for _, s := range validVersions {
		t.Run(s, func(t *testing.T) {
			ver, err := RuntimeVersion(s)
			t.Log("Valid: ", s, ver, err)
			if err != nil {
				t.Errorf("RuntimeVersion unexpected error for version %q: %v", s, err)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name      string
		current   string
		other     string
		wantError bool
		want      int
	}{
		{
			name:      "lessThan",
			current:   "release-2.7.2-SNAPSHOT-3714f2b",
			other:     "2.8.0",
			wantError: false,
			want:      -1,
		},
		{
			name:      "error",
			current:   "test-2.7.2-SNAPSHOT-3714f2b",
			other:     "2.8.0",
			wantError: true,
			want:      0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Compare(tt.current, tt.other)
			gotErr := err != nil
			if gotErr != tt.wantError {
				t.Errorf("testcase %v compare()'s expected error is %v, result is %v", tt.name, tt.wantError, err)
			}

			if got != tt.want {
				t.Errorf("testcase %v compare()'s expected value is %v, result is %v", tt.name, tt.want, got)
			}

		})
	}
}
