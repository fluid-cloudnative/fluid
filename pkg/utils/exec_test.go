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

import (
	"os/exec"
	"reflect"
	"testing"
)

func TestCheckCommandArgs(t *testing.T) {
	type args struct {
		arg []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test with illegal arguments",
			args: args{
				arg: []string{"ls", "|", "grep go"},
			},
			wantErr: true,
		},
		{
			name: "Test with legal arguments",
			args: args{
				arg: []string{"ls"},
			},
			wantErr: false,
		}, {
			name: "Test with legal arguments2",
			args: args{
				arg: []string{"echo test > /dev/null"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkCommandArgs(tt.args.arg...); (err != nil) != tt.wantErr {
				t.Errorf("checkCommandArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSimpleCommand(t *testing.T) {
	type args struct {
		name string
		arg  []string
	}
	tests := []struct {
		name    string
		args    args
		wantCmd *exec.Cmd
		wantErr bool
	}{
		{
			name: "Allow path list",
			args: args{
				name: "kubectl",
				arg:  []string{"Hello", "World"},
			},
			wantCmd: &exec.Cmd{
				Path: "echo",
				Args: []string{"Hello", "World"},
			},
			wantErr: false,
		}, {
			name: "Valid command arguments",
			args: args{
				name: "echo",
				arg:  []string{"Hello", "World"},
			},
			wantCmd: &exec.Cmd{
				Path: "echo",
				Args: []string{"Hello", "World"},
			},
			wantErr: false,
		},
		{
			name: "Invalid command arguments",
			args: args{
				name: "echo",
				arg:  []string{"Hello", "World&"},
			},
			wantCmd: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := SimpleCommand(tt.args.name, tt.args.arg...)
			if (err != nil) != tt.wantErr {
				t.Errorf("SimpleCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			tt.wantCmd = exec.Command(tt.args.name, tt.args.arg...)
			if !reflect.DeepEqual(tt.wantCmd.Args, cmd.Args) {
				t.Errorf("SimpleCommand() = %v, want %v", tt.args.arg, cmd.Args)
			}
			if !reflect.DeepEqual(tt.wantCmd.Path, cmd.Path) {
				t.Errorf("SimpleCommand() = %v, want %v", tt.args.arg, cmd.Args)
			}
		})
	}
}
