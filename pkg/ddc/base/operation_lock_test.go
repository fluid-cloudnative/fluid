/*
Copyright 2022 The Fluid Authors.

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

package base

import "testing"

func TestGetDataBackupRef(t *testing.T) {
	type args struct {
		name      string
		namespace string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				name:      "test",
				namespace: "default",
			},
			want: "default/test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDataOperationRef(tt.args.name, tt.args.namespace); got != tt.want {
				t.Errorf("GetDataBackupRef() = %v, want %v", got, tt.want)
			}
		})
	}
}
