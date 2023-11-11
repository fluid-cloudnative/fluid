/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
