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

func PipeCommand(name string, arg ...string) (cmd *exec.Cmd, err error) {
	// prepare the slice for ValidatePipeCommandSlice
	var commands []string
	commands = append(commands, name)
	commands = append(commands, arg...)

	// validate commands
	err = ValidatePipeCommandSlice(commands)
	if err != nil {
		return nil, err
	}

	return exec.Command(name, arg...), nil
}

func ValidatePipeCommandSlice(shellCommandSlice []string) (err error) {
	// Make sure the shell command is allowed
	var AllowedShellCommands = map[string]bool{
		"bash -c": true,
		"sh -c":   true,
	}

	// check if shellCommandSlice has enough arguments
	if len(shellCommandSlice) < 3 {
		return fmt.Errorf("insufficient arguments. Expected at least 3, received %d", len(shellCommandSlice))
	}
	// We assume -c always directly follows the shell command
	shellCommand := strings.Join(strings.Fields(shellCommandSlice[0]+" "+shellCommandSlice[1]), " ")
	if _, ok := AllowedShellCommands[shellCommand]; !ok {
		return fmt.Errorf("unknown shell command: %s", shellCommand)
	}

	for _, command := range shellCommandSlice[2:] {
		if err := ValidateShellPipeString(command); err != nil {
			return err
		}
	}
	return
}

// ValidateShellPipeString function checks whether the input command string is safe to execute.
// It checks whether all parts of a pipeline command start with any command prefixes defined in AllowedCommands
// It also checks for any illegal sequences that may lead to command injection attack.
func ValidateShellPipeString(command string) error {
	// Define illegal sequences that may lead to command injection attack
	illegalSequences := []string{"&", ";", "$", "'", "`", "(", ")", "||", ">>"}
	// Separate parts of pipeline command
	pipelineCommands := strings.Split(command, "|")

	// AllowedCommands is a global map that contains all allowed command prefixes.
	var AllowedCommands = map[string]bool{
		"ls":      false,
		"df":      false,
		"mount":   false,
		"alluxio": false,
		"goosefs": false,
		"kubectl": false,
		"helm":    false,
	}

	// AllowedPipeCommands is a map that contains all allowed pipe command prefixes.
	var allowedPipeCommands = map[string]bool{
		"grep":  false, // false means partial match
		"wc -l": true,  // true means full match (wc -l is exactly the allowed command)
		// Add more commands as you see fit
	}

	// Check each part of pipeline command
	for i, cmd := range pipelineCommands {
		cmd = strings.Join(
			strings.Fields(
				strings.TrimSpace(cmd)), " ")

		if i > 0 {
			// Check whether command starts with any allowed command prefix
			validCmd := isValidCommand(cmd, allowedPipeCommands)

			// If none of the allowed command prefix is found, throw error
			if !validCmd {
				return fmt.Errorf("full pipeline command not supported: part %d contains unsupported command '%s', the whole command %s", i+1, cmd, command)
			}
		} else {
			validCmd := isValidCommand(cmd, AllowedCommands)
			// If none of the allowed command prefix is found, throw error
			if !validCmd {
				return fmt.Errorf("full pipeline command not supported: part %d contains unsupported command '%s', the whole command %s", i+1, cmd, command)
			}
		}

		// Check for illegal sequences in command
		for _, illegalSeq := range illegalSequences {
			if strings.Contains(cmd, illegalSeq) {
				return fmt.Errorf("unsafe pipeline command %s, illegal sequence detected: %s in part %d: '%s'", command, illegalSeq, i+1, cmd)
			}
		}
	}

	// If no error found, return nil
	return nil
}

// Defining a function to check if the command is valid
func isValidCommand(cmd string, allowedCommands map[string]bool) bool {
	for cmdPrefix, exactMatch := range allowedCommands {
		if exactMatch && cmd == cmdPrefix {
			return true
		} else if !exactMatch && strings.HasPrefix(cmd, cmdPrefix) {
			return true
		}
	}
	return false
}
