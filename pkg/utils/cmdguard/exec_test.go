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

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("cmdguard", func() {
	Describe("checkCommandArgs", func() {
		DescribeTable("argument validation",
			func(arg []string, wantErr bool) {
				err := checkCommandArgs(arg...)
				if wantErr {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			},
			Entry("illegal arguments", []string{"ls", "|", "grep go"}, true),
			Entry("legal arguments", []string{"ls"}, false),
			Entry("illegal redirection", []string{"echo test > /dev/null"}, true),
		)
	})

	Describe("Command", func() {
		DescribeTable("command creation",
			func(name string, arg []string, wantCmd *exec.Cmd, wantErr bool) {
				cmd, err := Command(name, arg...)
				if wantErr {
					Expect(err).To(HaveOccurred())
					Expect(cmd).To(BeNil())
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(reflect.DeepEqual(cmd, wantCmd)).To(BeTrue())
				}
			},
			Entry("Allowed path list", "kubectl", []string{"create", "configmap"}, exec.Command("kubectl", "create", "configmap"), false),
			Entry("Valid command arguments", "echo", []string{"Hello", "World"}, exec.Command("echo", "Hello", "World"), false),
			Entry("Invalid command arguments", "echo", []string{"Hello", "World&"}, nil, true),
			Entry("Valid shell command", "bash", []string{"-c", "echo hello world"}, exec.Command("bash", "-c", "echo hello world"), false),
			Entry("Invalid shell command", "sh", []string{"/entrypoint.sh"}, nil, true),
			Entry("Valid shell pipeline command", "sh", []string{"-c", "ls -lh | grep -c test"}, exec.Command("sh", "-c", "ls -lh | grep -c test"), false),
			Entry("Invalid shell pipeline command (invalid pipelined command)", "bash", []string{"-c", "du -sh ./ | xargs echo"}, nil, true),
			Entry("Invalid shell pipeline command (invalid first command)", "bash", []string{"-c", "echo hello | grep hello"}, nil, true),
			Entry("Invalid shell pipeline command (illegal sequence)", "bash", []string{"-c", "ls -lh $(cat myfile) | grep hello"}, nil, true),
		)
	})

	Describe("buildPathList", func() {
		var (
			patches *gomonkey.Patches
		)
		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})
		It("should add found path for kubectl", func() {
			mockLookpathFunc := func(file string) (string, error) {
				return "/path/to/" + file, nil
			}
			patches = gomonkey.ApplyFunc(exec.LookPath, mockLookpathFunc)
			got := buildPathList(map[string]bool{"kubectl": true})
			Expect(got).To(Equal(map[string]bool{"kubectl": true, "/path/to/kubectl": true}))
		})
		It("should not add path for nonexistent command", func() {
			mockLookpathFunc := func(file string) (string, error) {
				return "", fmt.Errorf("Failed to find path")
			}
			patches = gomonkey.ApplyFunc(exec.LookPath, mockLookpathFunc)
			got := buildPathList(map[string]bool{"nonexistent": true})
			Expect(got).To(Equal(map[string]bool{"nonexistent": true}))
		})
		It("should keep full path command as is", func() {
			mockLookpathFunc := func(file string) (string, error) {
				return "/path/to/" + file, nil
			}
			patches = gomonkey.ApplyFunc(exec.LookPath, mockLookpathFunc)
			got := buildPathList(map[string]bool{"/usr/local/bin/kubectl": true})
			Expect(got).To(Equal(map[string]bool{"/usr/local/bin/kubectl": true}))
		})
	})

	Describe("ValidateCommandSlice", func() {
		DescribeTable("command slice validation",
			func(commandSlice []string, wantErr bool) {
				err := ValidateCommandSlice(commandSlice)
				if wantErr {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			},
			Entry("allowed path list", []string{"kubectl", "create", "foobar"}, false),
			Entry("valid command args", []string{"echo", "hello"}, false),
			Entry("invalid command args", []string{"echo", "$MYENV"}, true),
			Entry("valid shell command", []string{"sh", "-c", "echo test"}, false),
			Entry("invalid shell command", []string{"sh", "-c", "echo $MYENV"}, true),
			Entry("invalid shell command", []string{"sh", "myscript.sh"}, true),
			Entry("valid shell piped command", []string{"bash", "-c", "ls -lh ./ | wc -l"}, false),
			Entry("invalid shell piped command(invalid pipe command)", []string{"bash", "-c", "ls -lh ./ | xargs echo"}, true),
			Entry("invalid shell piped command(invalid first command)", []string{"bash", "-c", "echo foobar | grep foo"}, true),
			Entry("invalid shell piped command(illegal sequence)", []string{"bash", "-c", "du -sh ./ | grep $(foo) | wc -l"}, true),
		)
	})
})
