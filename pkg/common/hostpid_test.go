/*
Copyright 2024 The Fluid Authors.

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

package common

import "testing"

func TestHostPIDEnabled(t *testing.T) {
	type args struct {
		annotations map[string]string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil, return false",
			args: args{
				annotations: nil,
			},
			want: false,
		},
		{
			name: "not exist, return false",
			args: args{
				annotations: map[string]string{},
			},
			want: false,
		},
		{
			name: "wrong value, return false",
			args: args{
				annotations: map[string]string{
					RuntimeFuseHostPIDKey: "sss",
				},
			},
			want: false,
		},
		{
			name: "exist, return true",
			args: args{
				annotations: map[string]string{
					RuntimeFuseHostPIDKey: "true",
				},
			},
			want: true,
		},
		{
			name: "exist True, return true",
			args: args{
				annotations: map[string]string{
					RuntimeFuseHostPIDKey: "True",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HostPIDEnabled(tt.args.annotations); got != tt.want {
				t.Errorf("HostPIDEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}
