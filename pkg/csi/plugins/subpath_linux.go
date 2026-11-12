/*
Copyright 2026 The Fluid Authors.

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

package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

// openFDFlags names a path component without following it if it turns out to be a symlink.
// O_PATH keeps the open cheap and side effect free: the descriptor only names the inode and is
// never read from. Note that O_PATH|O_NOFOLLOW does not fail on a symlink, it returns a descriptor
// for the link itself, so every component is stat'ed to reject one.
const openFDFlags = unix.O_NOFOLLOW | unix.O_PATH | unix.O_CLOEXEC

// resolveMountSource resolves subPath below base and returns a path that is safe to hand to
// mount(8) as the bind mount source, together with a closer to call once the mount is done.
//
// Every component of subPath is opened with O_NOFOLLOW relative to its parent descriptor, so a
// symlink anywhere along the way - not only in the last component - is rejected. The returned
// source is a /proc/self/fd entry for the descriptor that was verified, which pins the resolved
// inode: swapping a component for a symlink after the check no longer changes what gets mounted.
func resolveMountSource(base, subPath string) (source string, closer func(), err error) {
	if subPath == "" {
		return base, func() {}, nil
	}

	fd, err := openBeneathNoSymlinks(base, subPath)
	if err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("/proc/%d/fd/%d", os.Getpid(), fd), func() { _ = syscall.Close(fd) }, nil
}

// openBeneathNoSymlinks walks subPath one component at a time below base and returns a descriptor
// for the final component. It fails if base or any component is a symlink. The caller owns the
// returned descriptor.
func openBeneathNoSymlinks(base, subPath string) (int, error) {
	if !filepath.IsLocal(subPath) {
		return -1, fmt.Errorf("subPath %q must be a relative path that does not escape %s", subPath, base)
	}

	parentFD, err := syscall.Open(base, openFDFlags, 0)
	if err != nil {
		return -1, errors.Wrapf(err, "failed to open mount point %s", base)
	}

	succeeded := false
	defer func() {
		if !succeeded {
			_ = syscall.Close(parentFD)
		}
	}()

	if err := rejectSymlinkFD(parentFD, base); err != nil {
		return -1, err
	}

	// filepath.Clean collapses the "." and ".." elements IsLocal tolerates as long as they resolve
	// below base, so the remaining segments are all plain directory entries.
	for _, segment := range strings.Split(filepath.Clean(subPath), string(filepath.Separator)) {
		childFD, err := syscall.Openat(parentFD, segment, openFDFlags, 0)
		if err != nil {
			return -1, errors.Wrapf(err, "failed to open %q of subPath %q below mount point %s", segment, subPath, base)
		}

		_ = syscall.Close(parentFD)
		parentFD = childFD

		if err := rejectSymlinkFD(parentFD, filepath.Join(base, segment)); err != nil {
			return -1, err
		}
	}

	succeeded = true
	return parentFD, nil
}

// rejectSymlinkFD returns an error if fd names a symlink. O_PATH|O_NOFOLLOW succeeds on a symlink
// and yields a descriptor for the link itself, so the check has to be explicit.
func rejectSymlinkFD(fd int, name string) error {
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		return errors.Wrapf(err, "failed to stat %s", name)
	}

	if stat.Mode&syscall.S_IFMT == syscall.S_IFLNK {
		return fmt.Errorf("%s is a symlink, which is not allowed for mounting", name)
	}

	return nil
}
