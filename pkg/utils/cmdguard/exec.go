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
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log logr.Logger

// allowedPathList is a whitelist of safe commands
var allowedPathList = map[string]bool{
	// "helm":    true,
	//"kubectl":  true,
	"ddc-helm": true,
	// add other commands as needed
}

var shellCommandList = map[string]bool{
	"bash": true,
	"sh":   true,
}

// illegalChars to check
var illegalChars = []rune{'&', '|', ';', '$', '\'', '`', '(', ')', '>'}

func init() {
	allowedPathList = buildPathList(allowedPathList)
	shellCommandList = buildPathList(shellCommandList)
	log = ctrl.Log.WithName("utils.security")
}

// buildPathList is a function that builds a map of paths for the given pathList.
func buildPathList(pathList map[string]bool) (targetPath map[string]bool) {
	targetPath = make(map[string]bool)
	for name, enabled := range pathList {
		if filepath.Base(name) == name {
			path, err := exec.LookPath(name)
			if err != nil {
				log.Info("Failed to find path %s due to %v", path, err)
			} else {
				targetPath[path] = enabled
			}
		}
		targetPath[name] = enabled
	}
	return
}

// Command checks the args before creating *exec.Cmd
func Command(name string, arg ...string) (cmd *exec.Cmd, err error) {
	commandSlice := append([]string{name}, arg...)

	if err = ValidateCommandSlice(commandSlice); err != nil {
		return
	}

	return exec.Command(name, arg...), nil
}

// ValidateCommandSlice validates all the commands in the commandSlice.
// - For command in allowedPathList, it passes validation without further checks.
// - For a possible shell command, it calls ValidateShellCommandSlice() for detailed checks.
// - For any other command, it checks all the command args to ensure no illegal chars exists.
func ValidateCommandSlice(commandSlice []string) (err error) {
	if len(commandSlice) == 0 {
		return nil
	}

	mainCmd := commandSlice[0]
	if allowedPathList[mainCmd] {
		return nil
	}

	if shellCommandList[mainCmd] {
		return ValidateShellCommandSlice(commandSlice)
	}

	// For any other mainCmd, check illegal chars in command args
	return checkCommandArgs(commandSlice[1:]...)
}

// CheckCommandArgs is check string is valid in args
func checkCommandArgs(arg ...string) (err error) {
	for _, value := range arg {
		for _, illegalChar := range illegalChars {
			if strings.ContainsRune(value, illegalChar) {
				return fmt.Errorf("args %s has illegal access with illegalChar %c", value, illegalChar)
			}
		}
	}

	return
}

// argIllegalChars is illegalChars extended with the characters that carry no meaning in a
// plain argument or filesystem path but that a shell would act on: input redirection, quoting,
// glob and brace expansion, home expansion, history expansion and comments.
var argIllegalChars = append(append([]rune{}, illegalChars...),
	'<', '"', '\\', '!', '*', '?', '{', '}', '[', ']', '~', '#')

// ValidateArg validates a single argument, such as a path, that comes from user input and is
// later embedded in a command or rendered into a script.
//
// It is deliberately stricter than the argument check ValidateCommandSlice applies: on top of
// illegalChars it rejects the characters listed in argIllegalChars and every control character,
// newline and NUL included. Call it where such a value is first parsed, so that the consumers
// which never reach ValidateCommandSlice are covered as well - values rendered into Helm charts,
// for example, are executed inside a pod and never pass through this package.
//
// Whitespace is intentionally allowed. Spaces occur in real paths and are harmless once the
// value is passed as its own element of an argv vector.
func ValidateArg(arg string) error {
	for _, illegalChar := range argIllegalChars {
		if strings.ContainsRune(arg, illegalChar) {
			return fmt.Errorf("arg %q contains illegal character %q", arg, illegalChar)
		}
	}

	for _, r := range arg {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("arg %q contains illegal control character %q", arg, r)
		}
	}

	return nil
}
