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
