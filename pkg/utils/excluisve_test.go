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
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
)

func TestGetExclusiveKey(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "test for GetExclusiveKey",
			want: common.FluidExclusiveKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetExclusiveKey(); got != tt.want {
				t.Errorf("GetExclusiveKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetExclusiveValue(t *testing.T) {
	type args struct {
		namespace string
		name      string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "default test-dataset-1",
			args: args{
				name:      "test-dataset-1",
				namespace: "default",
			},
			want: "default_test-dataset-1",
		},
		{
			name: "otherns test-dataset-2",
			args: args{
				name:      "test-dataset-2",
				namespace: "otherns",
			},
			want: "otherns_test-dataset-2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetExclusiveValue(tt.args.namespace, tt.args.name); got != tt.want {
				t.Errorf("GetExclusiveValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
