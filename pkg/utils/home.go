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
	"bytes"
	"errors"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
)

// Home returns the home directory for the executing user.
// This uses an OS-specific method for discovering the home directory.
// An error is returned if a home directory cannot be detected.
func Home() (string, error) {
	user, err := user.Current()
	if nil == err {
		return user.HomeDir, nil
	}

	// cross compile support

	if runtime.GOOS == "windows" {
		return homeWindows()
	}

	// Unix-like system, so just assume Unix
	return homeUnix()
}

func homeUnix() (home string, err error) {
	// First prefer the HOME environmental variable
	home = os.Getenv("HOME")
	if home != "" {
		return home, nil
	}

	// If that fails, try the shell
	var stdout bytes.Buffer
	// cmd, err := security.Command("sh", "-c", "eval echo ~$USER")
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")

	cmd.Stdout = &stdout
	err = cmd.Run()
	if err != nil {
		return
	}

	home = strings.TrimSpace(stdout.String())
	if home == "" {
		err = errors.New("blank output when reading home directory")
		return
	}

	return
}

func homeWindows() (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		return "", errors.New("HOMEDRIVE, HOMEPATH, and USERPROFILE are blank")
	}

	return home, nil
}
