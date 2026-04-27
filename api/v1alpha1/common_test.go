/*
Copyright 2026 The Fluid Authors.

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

package v1alpha1

import "testing"

func TestMetadataSyncPolicy_AutoSyncEnabled(t *testing.T) {
	trueVal := true
	falseVal := false

	testCases := map[string]struct {
		policy *MetadataSyncPolicy
		want   bool
	}{
		"nil policy defaults to true": {
			policy: nil,
			want:   true,
		},
		"nil autoSync defaults to true": {
			policy: &MetadataSyncPolicy{},
			want:   true,
		},
		"explicit true returns true": {
			policy: &MetadataSyncPolicy{AutoSync: &trueVal},
			want:   true,
		},
		"explicit false returns false": {
			policy: &MetadataSyncPolicy{AutoSync: &falseVal},
			want:   false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := tc.policy.AutoSyncEnabled()
			if got != tc.want {
				t.Fatalf("AutoSyncEnabled() = %v, want %v", got, tc.want)
			}
		})
	}
}
