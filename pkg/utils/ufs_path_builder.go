/*
Copyright 2023 The Fluid Author.

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
	"path/filepath"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

type UFSPathBuilder struct{}

// GenUFSPathInUnifiedNamespace generates a path in the cache engine's [unified namespace] for
// the given mount. It follows the convention defined internally by Fluid:
//
// 1. if `dataset.spec.mounts[*].path` is set to a absolute path, pick `path`.
// 2. otherwise, pick `/{dataset.spec.mounts[*].name}`
//
// [unified namespace]: https://docs.alluxio.io/os/user/stable/en/core-services/Unified-Namespace.html
func (u UFSPathBuilder) GenUFSPathInUnifiedNamespace(mount datav1alpha1.Mount) string {

	// if the user defines mount.path, use it
	if filepath.IsAbs(mount.Path) {
		return mount.Path
	}

	return fmt.Sprintf(common.UFSMountPathFormat, mount.Name)
}

// GenAlluxioUFSRootPath determines which mount point should be mounted on the root path of
// the [unified namespace] in Alluxio engine. Commonly there are two cases:
//
//  1. If a `mount` item is the only item defined in `dataset.sepc.mounts[*]` and its ufs path equals to "/", its `mountpoint` should be on the root path.
//     e.g. alluxio fs mount s3://mybucket /
//  2. Otherwise, pick `/underFSStorage` as the default root path.
//     e.g. alluxio fs mount /underFSStorage / && alluxio fs mount s3://mybucket /mybucket
//
// [unified namespace]: https://docs.alluxio.io/os/user/stable/en/core-services/Unified-Namespace.html
func (u UFSPathBuilder) GenAlluxioUFSRootPath(items []datav1alpha1.Mount) (string, *datav1alpha1.Mount) {
	if len(items) == 1 {
		m := items[0]
		// only iff m matches all of the two following conditions (1) m is not a fluid-native mount point; (2) m.Path is "/", it should be the root path in cache UFS.
		if !common.IsFluidNativeScheme(m.MountPoint) && u.GenUFSPathInUnifiedNamespace(m) == common.RootDirPath {
			return m.MountPoint, &m
		}
	}

	return u.GetLocalStorageRootDir(), nil
}

// this value will be the default value for the alluxio configuration:
//
//	alluxio.master.mount.table.root.ufs
//
// e.g. :
//
//	$ alluxio fs mount
//	/underFSStorage  on  /  (local, capacity=0B, used=-1B, not read-only, not shared, properties={})
func (u UFSPathBuilder) GetLocalStorageRootDir() string {
	return common.LocalStorageRootPath
}

// generate local storage path by mount info
func (u UFSPathBuilder) GenLocalStoragePath(curMount datav1alpha1.Mount) string {

	if filepath.IsAbs(curMount.Path) {
		return filepath.Join(common.LocalStorageRootPath, curMount.Path)
	}

	return filepath.Join(common.LocalStorageRootPath, curMount.Name)
}
