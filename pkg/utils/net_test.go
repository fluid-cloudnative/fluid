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

import (
	"reflect"
	"testing"
)

func TestSortIpAddresses(t *testing.T) {
	tests := []struct {
		name string
		ips  []string
		want []string
	}{
		{
			name: "empty_input",
			ips:  []string{},
			want: nil,
		},
		{
			name: "single_ip",
			ips:  []string{"10.0.0.1"},
			want: []string{"10.0.0.1"},
		},
		{
			name: "already_sorted",
			ips:  []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"},
			want: []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"},
		},
		{
			name: "unsorted_ips",
			ips:  []string{"10.0.0.3", "10.0.0.1", "10.0.0.2"},
			want: []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"},
		},
		{
			name: "duplicate_ips_are_deduplicated",
			ips:  []string{"10.0.0.1", "10.0.0.2", "10.0.0.1"},
			want: []string{"10.0.0.1", "10.0.0.2"},
		},
		{
			name: "cross_subnet_ordering",
			ips:  []string{"192.168.1.1", "10.0.0.1", "172.16.0.1"},
			want: []string{"10.0.0.1", "172.16.0.1", "192.168.1.1"},
		},
		{
			name: "last_octet_ordering",
			ips:  []string{"10.0.0.100", "10.0.0.20", "10.0.0.3"},
			want: []string{"10.0.0.3", "10.0.0.20", "10.0.0.100"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SortIpAddresses(tt.ips)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SortIpAddresses() = %v, want %v", got, tt.want)
			}
		})
	}
}
