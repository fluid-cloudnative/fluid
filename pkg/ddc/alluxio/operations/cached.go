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
