/*

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

package alluxio

import (
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

type UFSPathBuilder struct{}

// dataset.spec.mounts mount to alluxio instance strategy:
//
// strategy && priority:
// 1. if set dataset.spec.mounts[x].path
// 2. if only one item use default root path "/"
// 3. "/" + dataset.spec.mounts[x].name
func (u UFSPathBuilder) GenAlluxioMountPath(curMount datav1alpha1.Mount, mounts []datav1alpha1.Mount) string {

	// if the user defines mount.path, use it
	if len(curMount.Path) > 0 {
		return curMount.Path
	}
	// if dataset only has one mount item
	if len(mounts) == 1 {
		return common.RootDirPath
	}

	return fmt.Sprintf(common.AlluxioMountPathFormat, curMount.Name)
}

// value for alluxio instance configuration :
//
//  alluxio.master.mount.table.root.ufs
//
// two situations
//	1. mount local storage root path as alluxio root path
//     e.g. : alluxio fs mount
//            /underFSStorage /
// 	2. direct mount ufs endpoint as alluxio root path
//     e.g. : alluxio fs mount
//            http://fluid.io/apache/spark/spark-3.0.2 /
func (u UFSPathBuilder) GenAlluxioUFSRootPath(items []datav1alpha1.Mount) (string, *datav1alpha1.Mount) {
	// if have multi ufs mount point or empty
	// use local storage root path by default
	if len(items) > 1 || len(items) == 0 {
		return u.GetLocalStorageRootDir(), nil
	}

	m := items[0]

	// if fluid native scheme : use local storage root path
	if common.IsFluidNativeScheme(m.MountPoint) {
		return u.GetLocalStorageRootDir(), nil
	}

	// if user define mount.path : use local storage root path
	if m.Path != "" && m.Path != common.RootDirPath {
		return u.GetLocalStorageRootDir(), nil
	}

	// use ufs path as alluxio root path
	return m.MountPoint, &m

}

// this value will be the default value for the alluxio configuration:
//   alluxio.master.mount.table.root.ufs
//
// e.g. :
//   $ alluxio fs mount
//   /underFSStorage  on  /  (local, capacity=0B, used=-1B, not read-only, not shared, properties={})
func (u UFSPathBuilder) GetLocalStorageRootDir() string {
	return common.AlluxioLocalStorageRootPath
}

// generate local storage path by mount info
func (u UFSPathBuilder) GenLocalStoragePath(curMount datav1alpha1.Mount) string {
	return fmt.Sprintf(common.AlluxioLocalStoragePathFormat, curMount.Name)
}
