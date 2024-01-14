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

package utils

import "testing"

func TestValidateShellPipeString(t *testing.T) {
	type args struct {
		command string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// Add test cases.
		{name: "valid command with grep", args: args{command: "kubectl get pods | grep Running"}, wantErr: false},
		{name: "valid command with wc -l", args: args{command: "echo hello | wc -l"}, wantErr: false},
		{name: "invalid command", args: args{command: "rm -rf /"}, wantErr: true},
		{name: "illegal sequence in command", args: args{command: "kubectl get pods | grep Running; rm -rf /"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateShellPipeString(tt.args.command); (err != nil) != tt.wantErr {
				t.Errorf("ValidateShellPipeString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
