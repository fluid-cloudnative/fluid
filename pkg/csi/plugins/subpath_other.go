//go:build !linux

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

	"github.com/pkg/errors"
	"k8s.io/utils/mount"
)

// resolveMountSource resolves subPath below base and returns the bind mount source.
//
// The CSI plugin only runs on Linux; this build exists so the package stays buildable and testable
// elsewhere. It checks every component of subPath for a symlink instead of pinning a descriptor, so
// unlike the Linux implementation it cannot rule out a component being swapped after the check.
func resolveMountSource(base, subPath string) (source string, closer func(), err error) {
	if subPath == "" {
		return base, func() {}, nil
	}

	if !filepath.IsLocal(subPath) {
		return "", nil, fmt.Errorf("subPath %q must be a relative path that does not escape %s", subPath, base)
	}

	current := base
	for _, segment := range strings.Split(filepath.Clean(subPath), string(filepath.Separator)) {
		current = filepath.Join(current, segment)

		fi, err := os.Lstat(current)
		if err != nil {
			if mount.IsCorruptedMnt(err) {
				return "", nil, fmt.Errorf("mount point %s is corrupted", base)
			}
			return "", nil, errors.Wrapf(err, "failed to lstat %q of subPath %q below mount point %s", segment, subPath, base)
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			return "", nil, fmt.Errorf("subPath %q of mount point %s contains a symlink at %q, which is not allowed", subPath, base, segment)
		}
	}

	return current, func() {}, nil
}
