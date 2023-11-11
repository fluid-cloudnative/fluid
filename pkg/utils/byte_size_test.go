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

package utils

import "testing"

func TestBytesSize(t *testing.T) {
	testCases := map[string]struct {
		b    float64
		want string
	}{
		"test bytes size case 1": {
			b:    100.0,
			want: "100.00B",
		},
		"test bytes size case 2": {
			b:    1024 * 4,
			want: "4.00KiB",
		},
		"test bytes size case 3": {
			b:    1024 * 1024,
			want: "1.00MiB",
		},
		"test bytes size case 4": {
			b:    1024 * 1024 * 5,
			want: "5.00MiB",
		},
		"test bytes size case 5": {
			b:    1024 * 1024 * 1024,
			want: "1.00GiB",
		},
		"test bytes size case 6": {
			b:    1024 * 1024 * 1024 * 1024,
			want: "1.00TiB",
		},
		"test bytes size case 7": {
			b:    1024 * 1024 * 1024 * 1024 * 10,
			want: "10.00TiB",
		},
		"test bytes size case 8": {
			b:    1024 * 1024 * 1024 * 1024 * 1024,
			want: "1.00PiB",
		},
		"test bytes size case 9": {
			b:    1024 * 1024 * 1024 * 1024 * 1024 * 1024,
			want: "1.00EiB",
		},
		"test bytes size case 10": {
			b:    1024 * 1024 * 1024 * 1024 * 1024 * 1024,
			want: "1.00EiB",
		},
		"test bytes size case 11": {
			b:    1024 * 1024 * 1024 * 1024 * 1024 * 1024 * 1024 * 9,
			want: "9.00ZiB",
		},
		"test bytes size case 12": {
			b:    1024 * 1024 * 1024 * 1024 * 1024 * 1024 * 1024 * 1024 * 2,
			want: "2.00YiB",
		},
		"test bytes size case 13": {
			b:    1024 * 1024 * 1024 * 1024 * 1024 * 1024 * 1024 * 1024 * 1024,
			want: "1024.00YiB",
		},
		"test bytes size case 14": {
			b:    1024 * 1024 * 1024 * 1024 * 1024 * 1024 * 1024 * 1024 * 1024 * 1024 * 1024,
			want: "1073741824.00YiB",
		},
	}

	for k, item := range testCases {
		got := BytesSize(item.b)
		if got != item.want {
			t.Errorf("%s check failure,want:%s,got:%s", k, item.want, got)
		}
	}
}

func TestFromHumanSize(t *testing.T) {
	testCases := map[string]struct {
		s        string
		want     int64
		occurErr bool
	}{
		"test from human size case 1": {
			s:        "1KiB",
			want:     1024,
			occurErr: false,
		},
		"test from human size case 2": {
			s:        "10KiB",
			want:     10240,
			occurErr: false,
		},
		"test from human size case 3": {
			s:        "1MiB",
			want:     1048576,
			occurErr: false,
		},
		"test from human size case 4": {
			s:        "1GiB",
			want:     1073741824,
			occurErr: false,
		},
		"test from human size case 5": {
			s:        "1TiB",
			want:     1099511627776,
			occurErr: false,
		},
		"test from human size case 6": {
			s:        "1PiB",
			want:     1125899906842624,
			occurErr: false,
		},
	}

	for k, item := range testCases {
		got, err := FromHumanSize(item.s)
		if item.occurErr {
			if err == nil {
				t.Errorf("%s check failure, want err, but got nil", k)
			}
		} else {
			if got != item.want {
				t.Errorf("%s check failure, want: %d, got:%d", k, item.want, got)
			}
		}
	}
}
