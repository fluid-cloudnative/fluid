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
	"fmt"
	"os/exec"
	"strings"
)

// allowPathlist of safe commands
var allowPathlist = map[string]bool{
	// "helm":    true,
	"kubectl":  true,
	"ddc-helm": true,
	// add other commands as needed
}

// illegalChars to check
var illegalChars = []string{"&", "|", ";", "$", "'", "`", "(", ")", ">>"}

// SimpleCommand checks the args before creating *exec.Cmd
func SimpleCommand(name string, arg ...string) (cmd *exec.Cmd, err error) {
	if allowPathlist[name] {
		cmd = exec.Command(name, arg...)
	} else {
		err = checkCommandArgs(arg...)
		if err != nil {
			return
		}
		cmd = exec.Command(name, arg...)
	}

	return
}

// CheckCommandArgs is check string is valid in args
func checkCommandArgs(arg ...string) (err error) {
	for _, value := range arg {
		for _, illegalChar := range illegalChars {
			if strings.Contains(value, illegalChar) {
				return fmt.Errorf("args %s has illegal access with illegalChar %s", value, illegalChar)
			}
		}
	}

	return
}
