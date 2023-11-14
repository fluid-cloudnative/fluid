/*
Copyright 2022 The Fluid Authors.

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

package operations

import (
	"fmt"
	"strings"

	units "github.com/docker/go-units"
)

func (a GooseFSFileUtils) CachedState() (cached int64, err error) {
	var (
		command = []string{"goosefs", "fsadmin", "report"}
		stdout  string
		stderr  string
	)

	found := false
	stdout, stderr, err = a.exec(command, false)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}
	str := strings.Split(stdout, "\n")

	for _, s := range str {
		if strings.Contains(s, "Used Capacity:") {
			values := strings.Fields(s)
			if len(values) == 0 {
				return cached, fmt.Errorf("failed to parse %s", s)
			}
			cached, err = units.RAMInBytes(values[len(values)-1])
			if err != nil {
				return
			}
			found = true
		}
	}

	if !found {
		err = fmt.Errorf("failed to find the cache in output %v", stdout)
	}

	return
}

// clean cache with a preset timeout of 60s
func (a GooseFSFileUtils) CleanCache(path string) (err error) {
	var (
		releaseVersion = []string{"cat", "/etc/issue"}
		command        = []string{"60", "goosefs", "fs", "free", "-f", path}
		stdout         string
		stderr         string
	)

	stdout, stderr, err = a.exec(releaseVersion, false)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", releaseVersion, err, stdout, stderr)
		return
	}

	if strings.Contains(stdout, "Ubuntu") {
		command = append([]string{"timeout", "-k"}, command...)
	} else if strings.Contains(stdout, "Alpine") {
		command = append([]string{"timeout"}, command...)
	} else {
		err = fmt.Errorf("unknow release version for linux")
		return
	}

	stdout, stderr, err = a.exec(command, false)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}

	return
}
