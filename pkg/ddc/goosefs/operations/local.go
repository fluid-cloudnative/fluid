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
