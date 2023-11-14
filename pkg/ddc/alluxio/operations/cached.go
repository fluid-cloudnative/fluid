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

package operations

import (
	"fmt"
	"strconv"
	"strings"

	units "github.com/docker/go-units"
)

func (a AlluxioFileUtils) CachedState() (cached int64, err error) {
	var (
		command = []string{"alluxio", "fsadmin", "report"}
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
func (a AlluxioFileUtils) CleanCache(path string, timeout int32) (err error) {
	var (
		// command = []string{"60", "alluxio", "fs", "free", "-f", path}
		command = []string{strconv.FormatInt(int64(timeout), 10),
			"alluxio", "fs", "free", "-f", path}
		stdout string
		stderr string
	)

	//TODO : find solution to use "timeout" or "timeout -t" in different linux release
	command = append([]string{"timeout"}, command...)

	stdout, stderr, err = a.exec(command, false)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}

	return
}
