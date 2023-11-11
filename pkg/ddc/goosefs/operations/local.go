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
	"time"
)

// SyncLocalDir syncs local path by running command `du -sh <path>`.
// Under the circumstance where some NAS(e.g. NFS) is mounted on the `<path>`, the function will sync metadata of all files in the NAS.
// This is necessary for GooseFS to get consistent file metadata with UFS(i.e. NAS in this case).
func (a GooseFSFileUtils) SyncLocalDir(path string) (err error) {
	var (
		// command = []string{"goosefs", "fs", "-Dgoosefs.user.file.metadata.sync.interval=0", "ls", "-R", goosefsPath}
		// command = []string{"goosefs", "fs", "-Dgoosefs.user.file.metadata.sync.interval=0", "count", goosefsPath}
		command = []string{"du", "-sh", path}
		stdout  string
		stderr  string
	)

	start := time.Now()
	stdout, stderr, err = a.execWithoutTimeout(command, false)
	duration := time.Since(start)
	a.log.Info("du -sh", "path", path, "period", duration)
	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		return
	}

	return
}
