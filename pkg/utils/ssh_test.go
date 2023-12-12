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
)

func TestGenerateSSHConfig1(t *testing.T) {
	type args struct {
		releaseName string
		parallelism int32
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test ssh config",
			args: args{
				releaseName: "demo-migrate",
				parallelism: 2,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GenerateSSHConfig(tt.args.releaseName, tt.args.parallelism)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSSHConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
