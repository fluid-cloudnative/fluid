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
	"strings"

	"github.com/pkg/errors"
)

type CommandValidater func(str string, pattern string) bool

var (
	PrefixMatch CommandValidater = func(str, pattern string) bool { return strings.HasPrefix(str, pattern) }
	ExactMatch  CommandValidater = func(str, pattern string) bool { return str == pattern }
)

// Make sure the shell command is allowed
var AllowedShellCommands = map[string]CommandValidater{
	"bash -c": ExactMatch,
	"sh -c":   ExactMatch,
}

// allowedFirstCommands is a global map that contains all allowed prefix for the first command in a shell pipe.
var allowedFirstCommands = map[string]CommandValidater{
	"ls":       PrefixMatch,
	"df":       PrefixMatch,
	"mount":    PrefixMatch,
	"alluxio":  PrefixMatch,
	"goosefs":  PrefixMatch,
	"ddc-helm": PrefixMatch,
}

// AllowedPipedCommands is a map that contains all allowed piped commands and their validater functions.
var allowedPipedCommands = map[string]CommandValidater{
	"grep":  PrefixMatch,
	"wc -l": ExactMatch, // means exact match (wc -l is exactly the allowed command)
	// Add more commands as you see fit
}

// Define illegal sequences that may lead to command injection attack
var illegalSequences = []string{"&", ";", "$", "'", "`", "(", ")", "||", ">>"}

var allowedExpressions = []string{
	"${METAURL}", // JuiceFS community's metaurl
}

// ShellCommand is a safe wrapper of exec.Command that checks potential risks in the command.
// It requires the command follows the format like ["bash", "-c", "<shell script>"] and each part
// of the command must be valid. If no shell command is needed, use security.Command instead.
func ShellCommand(name string, arg ...string) (cmd *exec.Cmd, err error) {
	var commands = append([]string{name}, arg...)

	err = ValidateShellCommandSlice(commands)
	if err != nil {
		return nil, err
	}

	return exec.Command(name, arg...), nil
}

// ValidateShellCommandSlice takes in a slice of shell commands and returns an error if any are invalid.
// The function looks specifically for pipe commands (i.e., commands that contain a '|').
// If a pipe command is found in the slice, ValidatePipeCommandSlice is called for further validation.
func ValidateShellCommandSlice(shellCommandSlice []string) (err error) {
	shellCommand, shellScript, err := splitShellCommand(shellCommandSlice)
	if err != nil {
		return errors.Wrapf(err, "failed to split shell pipe command %v", shellCommandSlice)
	}

	if err := validateShellCommand(shellCommand); err != nil {
		return errors.Wrapf(err, "failed to validate shell command [%s]", shellCommand)
	}

	if strings.Contains(shellScript, "|") {
		// shellScript is possible a shell pipeline
		if err := validateShellPipeString(shellScript); err != nil {
			return errors.Wrapf(err, "failed to validate shell script [%s]", shellScript)
		}
	} else {
		if err := checkIllegalSequence(shellScript); err != nil {
			return errors.Wrap(err, "failed to pass illegal sequence check")
		}
	}

	return
}

func splitShellCommand(shellCommandSlice []string) (shellCommand string, pipedCommands string, err error) {
	// A shell pipeline command slice is REQUIRED to have 3 parts of commands, e.g. "bash", "-c", "<piped commands>"
	if len(shellCommandSlice) != 3 {
		err = fmt.Errorf("invalid shell command slice. Expected num of slice is exactly 3, received %d", len(shellCommandSlice))
		return
	}

	// We assume -c always directly follows the shell command
	shellCommand = strings.Join(strings.Fields(shellCommandSlice[0]+" "+shellCommandSlice[1]), " ")

	return shellCommand, shellCommandSlice[2], nil
}

func validateShellCommand(shellCommand string) (err error) {
	if _, ok := AllowedShellCommands[shellCommand]; !ok {
		return fmt.Errorf("unknown shell command: %s", shellCommand)
	}

	return nil
}

// validateShellPipeString function checks whether the input command string is safe to execute.
// It checks whether all parts of a pipeline command start with any command prefixes defined in AllowedCommands
// It also checks for any illegal sequences that may lead to command injection attack.
func validateShellPipeString(pipedCommandStr string) error {
	// Separate parts of pipeline command
	pipelineCommands := strings.Split(pipedCommandStr, "|")

	// Check each part of pipeline command
	for i, cmd := range pipelineCommands {
		cmd = strings.Join(
			strings.Fields(
				strings.TrimSpace(cmd)), " ")

		if i > 0 {
			// Check whether command starts with any allowed command prefix
			validCmd := isValidCommand(cmd, allowedPipedCommands)

			// If none of the allowed command prefix is found, throw error
			if !validCmd {
				return fmt.Errorf("full pipeline command not supported: part %d contains unsupported command '%s', the whole command %s", i+1, cmd, pipedCommandStr)
			}
		} else {
			validCmd := isValidCommand(cmd, allowedFirstCommands)
			// If none of the allowed command prefix is found, throw error
			if !validCmd {
				return fmt.Errorf("full pipeline command not supported: part %d contains unsupported command '%s', the whole command %s", i+1, cmd, pipedCommandStr)
			}
		}

		if err := checkIllegalSequence(cmd); err != nil {
			return errors.Wrap(err, "failed to pass illegal sequence check")
		}
	}

	// If no error found, return nil
	return nil
}

// Defining a function to check if the command is valid
func isValidCommand(cmd string, allowedCommands map[string]CommandValidater) bool {
	for allowedCmd, validaterFn := range allowedCommands {
		if validaterFn(cmd, allowedCmd) {
			return true
		}
	}

	return false
}

func checkIllegalSequence(script string) error {
	scriptToCheck := script
	for _, allowedEnv := range allowedExpressions {
		scriptToCheck = strings.ReplaceAll(scriptToCheck, allowedEnv, "ALLOWED_ENV")
	}

	// TODO: Simply check illegal sequence for now. Better filtered with a allowed list in future.
	for _, illegalSeq := range illegalSequences {
		if strings.Contains(scriptToCheck, illegalSeq) {
			return fmt.Errorf("unsafe shell script %s, illegal sequence detected: %s", script, illegalSeq)
		}
	}

	return nil
}
