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

package cmdguard

import (
	"fmt"
	"os/exec"
	"reflect"
	"testing"

	"github.com/brahma-adshonor/gohook"
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
			wantErr: true,
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

func TestCommand(t *testing.T) {
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
			name: "Allowed path list",
			args: args{
				name: "kubectl",
				arg:  []string{"create", "configmap"},
			},
			wantCmd: exec.Command("kubectl", "create", "configmap"),
			wantErr: false,
		}, {
			name: "Valid command arguments",
			args: args{
				name: "echo",
				arg:  []string{"Hello", "World"},
			},
			wantCmd: exec.Command("echo", "Hello", "World"),
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
		{
			name: "Valid shell command",
			args: args{
				name: "bash",
				arg:  []string{"-c", "echo hello world"},
			},
			wantCmd: exec.Command("bash", "-c", "echo hello world"),
			wantErr: false,
		},
		{
			name: "Invalid shell command",
			args: args{
				name: "sh",
				arg:  []string{"/entrypoint.sh"},
			},
			wantCmd: nil,
			wantErr: true,
		},
		{
			name: "Valid shell pipeline command",
			args: args{
				name: "sh",
				arg:  []string{"-c", "ls -lh | grep -c test"},
			},
			wantCmd: exec.Command("sh", "-c", "ls -lh | grep -c test"),
			wantErr: false,
		},
		{
			name: "Invalid shell pipeline command (invalid pipelined command)",
			args: args{
				name: "bash",
				arg:  []string{"-c", "du -sh ./ | xargs echo"},
			},
			wantCmd: nil,
			wantErr: true,
		},
		{
			name: "Invalid shell pipeline command (invalid first command)",
			args: args{
				name: "bash",
				arg:  []string{"-c", "echo hello | grep hello"},
			},
			wantCmd: nil,
			wantErr: true,
		},
		{
			name: "Invalid shell pipeline command (illegal sequence)",
			args: args{
				name: "bash",
				arg:  []string{"-c", "ls -lh $(cat myfile) | grep hello"},
			},
			wantCmd: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := Command(tt.args.name, tt.args.arg...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Command() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if !reflect.DeepEqual(tt.wantCmd, cmd) {
				t.Fatalf("Command() = %v, want %v", cmd, tt.wantCmd)
			}

			// tt.wantCmd = exec.Command(tt.args.name, tt.args.arg...)
			// if !reflect.DeepEqual(tt.wantCmd.Args, cmd.Args) {
			// 	t.Errorf("SimpleCommand() = %v, want %v", tt.args.arg, cmd.Args)
			// }
			// if !reflect.DeepEqual(tt.wantCmd.Path, cmd.Path) {
			// 	t.Errorf("SimpleCommand() = %v, want %v", tt.args.arg, cmd.Args)
			// }
		})
	}
}

func Test_buildPathList(t *testing.T) {
	type args struct {
		pathList map[string]bool
	}
	tests := []struct {
		name             string
		args             args
		mockLookpathFunc func(file string) (string, error)
		want             map[string]bool
	}{
		{
			name: "Test with command 'kubectl'",
			args: args{
				pathList: map[string]bool{"kubectl": true},
			},
			mockLookpathFunc: func(file string) (string, error) {
				return "/path/to/" + file, nil // Mocked path
			},
			want: map[string]bool{"kubectl": true, "/path/to/kubectl": true}, // assuming '/path/to/kubectl' is the path of the 'kubectl' command
		},
		{
			name: "Test with nonexistent command",
			args: args{
				pathList: map[string]bool{"nonexistent": true},
			}, mockLookpathFunc: func(file string) (string, error) {
				return "", fmt.Errorf("Failed to find path")
			},
			want: map[string]bool{"nonexistent": true}, // as 'nonexistent' command does not exist, so the result should be same as initial
		},
		{
			name: "Test with full path command",
			args: args{
				pathList: map[string]bool{"/usr/local/bin/kubectl": true},
			},
			mockLookpathFunc: func(file string) (string, error) {
				return "/path/to/" + file, nil // Mocked path
			},
			want: map[string]bool{"/usr/local/bin/kubectl": true}, // since '/usr/local/bin/kubectl' command already has full path, so the result should be same as initial
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gohook.Hook(exec.LookPath, tt.mockLookpathFunc, nil)
			if err != nil {
				t.Fatalf("failed to hook function: %v", err)
			}
			got := buildPathList(tt.args.pathList)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildPathList() = %v, want %v", got, tt.want)
			}
			_ = gohook.UnHook(tt.mockLookpathFunc)

		})
	}
}

func TestValidateCommandSlice(t *testing.T) {
	type args struct {
		commandSlice []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "allowed path list", args: args{[]string{"kubectl", "create", "foobar"}}, wantErr: false},
		{name: "valid command args", args: args{[]string{"echo", "hello"}}, wantErr: false},
		{name: "invalid command args", args: args{[]string{"echo", "$MYENV"}}, wantErr: true},
		{name: "valid shell command", args: args{[]string{"sh", "-c", "echo test"}}, wantErr: false},
		{name: "invalid shell command", args: args{[]string{"sh", "-c", "echo $MYENV"}}, wantErr: true},
		{name: "invalid shell command", args: args{[]string{"sh", "myscript.sh"}}, wantErr: true},
		{name: "valid shell piped command", args: args{[]string{"bash", "-c", "ls -lh ./ | wc -l"}}, wantErr: false},
		{name: "invalid shell piped command(invalid pipe command)", args: args{[]string{"bash", "-c", "ls -lh ./ | xargs echo"}}, wantErr: true},
		{name: "invalid shell piped command(invalid first command)", args: args{[]string{"bash", "-c", "echo foobar | grep foo"}}, wantErr: true},
		{name: "invalid shell piped command(illegal sequence)", args: args{[]string{"bash", "-c", "du -sh ./ | grep $(foo) | wc -l"}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateCommandSlice(tt.args.commandSlice); (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommandSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
