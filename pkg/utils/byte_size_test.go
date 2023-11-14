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
