package cmdguard

import (
	"os/exec"
	"reflect"
	"testing"
)

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

func TestValidateShellPipeString(t *testing.T) {
	type args struct {
		command string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "valid command with grep", args: args{command: "echo hello world | grep hello"}, wantErr: true},
		{name: "valid command with wc -l", args: args{command: "kubectl hello world | wc -l"}, wantErr: false},
		{name: "invalid command with xyz", args: args{command: "echo hello world | xyz"}, wantErr: true},
		{name: "illegal sequence in command with &", args: args{command: "echo hello world & echo y"}, wantErr: true},
		{name: "illegal sequence in command with ;", args: args{command: "ls ; echo y"}, wantErr: true},
		{name: "command with $", args: args{command: "kubectl $HOME"}, wantErr: true},
		{name: "command with absolute path", args: args{command: "ls /etc"}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateShellPipeString(tt.args.command); (err != nil) != tt.wantErr {
				t.Errorf("Testcase '%s' ValidateShellPipeString() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestShellCommand(t *testing.T) {
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
		{name: "valid simple command", args: args{name: "bash", arg: []string{"-c", "ls"}}, wantCmd: exec.Command("bash", "-c", "ls"), wantErr: false},
		{name: "insufficient arguments", args: args{name: "bash", arg: []string{"-c"}}, wantCmd: nil, wantErr: true},
		{name: "unknown shell command", args: args{name: "zsh", arg: []string{"-c", "ls"}}, wantCmd: nil, wantErr: true},
		{name: "valid piped command", args: args{name: "bash", arg: []string{"-c", "ls | grep something"}}, wantCmd: exec.Command("bash", "-c", "ls | grep something"), wantErr: false},
		{name: "invalid piped command", args: args{name: "bash", arg: []string{"-c", "ls | random-command"}}, wantCmd: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, err := ShellCommand(tt.args.name, tt.args.arg...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Testcase '%s': PipeCommand()  error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if gotCmd != nil && !reflect.DeepEqual(gotCmd.Path, tt.wantCmd.Path) {
				t.Errorf("Testcase '%s': PipeCommand()  = %v, want %v", tt.name, gotCmd, tt.wantCmd)
			}
			if gotCmd != nil && !reflect.DeepEqual(gotCmd.Args, tt.wantCmd.Args) {
				t.Errorf("Testcase '%s': PipeCommand()   = %v, want %v", tt.name, gotCmd, tt.wantCmd)
			}
		})
	}
}

// func TestValidateCommandSlice(t *testing.T) {
// 	type args struct {
// 		shellCommandSlice []string
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantErr bool
// 	}{
// 		{name: "valid bash command", args: args{shellCommandSlice: []string{"bash", "-c", "ls"}}, wantErr: false},
// 		{name: "valid sh command", args: args{shellCommandSlice: []string{"sh", "-c", "ls"}}, wantErr: false},
// 		{name: "unknown shell command", args: args{shellCommandSlice: []string{"zsh", "-c", "ls"}}, wantErr: true},
// 		{name: "invalid bash command", args: args{shellCommandSlice: []string{"bash", "-c", "wrong_command"}}, wantErr: true},
// 		{name: "insufficient arguments", args: args{shellCommandSlice: []string{"bash", "-c"}}, wantErr: true},
// 		{name: "empty command slice", args: args{shellCommandSlice: []string{}}, wantErr: true},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := ValidateShellCommandSlice(tt.args.shellCommandSlice); (err != nil) != tt.wantErr {
// 				t.Errorf("Testcase '%s': ValidatePipeCommandSlice() error = %v, wantErr %v", tt.name, err, tt.wantErr)
// 			}
// 		})
// 	}
// }

func TestIsValidCommand(t *testing.T) {
	type args struct {
		cmd             string
		allowedCommands map[string]CommandValidater
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "valid bash command", args: args{cmd: "bash", allowedCommands: map[string]CommandValidater{"bash": ExactMatch}}, want: true},
		{name: "valid sh command", args: args{cmd: "sh", allowedCommands: map[string]CommandValidater{"bash": ExactMatch, "sh": ExactMatch}}, want: true},
		{name: "invalid zsh command", args: args{cmd: "zsh", allowedCommands: map[string]CommandValidater{"bash": ExactMatch}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidCommand(tt.args.cmd, tt.args.allowedCommands); got != tt.want {
				t.Errorf("Testcase '%s': isValidCommand() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
