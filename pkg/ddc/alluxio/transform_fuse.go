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
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

// 4. Transform the fuse
func (e *AlluxioEngine) transformFuse(runtime *datav1alpha1.AlluxioRuntime, dataset *datav1alpha1.Dataset, value *Alluxio) (err error) {
	value.Fuse = Fuse{}

	value.Fuse.Image, value.Fuse.ImageTag = e.parseFuseImage()
	if runtime.Spec.Fuse.Image != "" {
		value.Fuse.Image = runtime.Spec.Fuse.Image
	}

	if runtime.Spec.Fuse.ImageTag != "" {
		value.Fuse.ImageTag = runtime.Spec.Fuse.ImageTag
	}

	value.Fuse.ImagePullPolicy = "IfNotPresent"
	if runtime.Spec.Fuse.ImagePullPolicy != "" {
		value.Fuse.ImagePullPolicy = runtime.Spec.Fuse.ImagePullPolicy
	}

	if len(runtime.Spec.Fuse.Properties) > 0 {
		value.Fuse.Properties = runtime.Spec.Fuse.Properties
	}

	if len(runtime.Spec.Fuse.Env) > 0 {
		value.Fuse.Env = runtime.Spec.Fuse.Env
	} else {
		value.Fuse.Env = map[string]string{}
	}

	// if runtime.Spec.Fuse.MountPath != "" {
	// 	value.Fuse.MountPath = runtime.Spec.Fuse.MountPath
	// } else {
	// 	value.Fuse.MountPath = fmt.Sprintf("format", a)
	// }

	value.Fuse.MountPath = e.getMountPoint()
	value.Fuse.Env["MOUNT_POINT"] = value.Fuse.MountPath

	// if len(runtime.Spec.Fuse.Args) > 0 {
	// 	value.Fuse.Args = runtime.Spec.Fuse.Args
	// } else {
	// 	value.Fuse.Args = []string{"fuse", "--fuse-opts=kernel_cache"}
	// }
	e.optimizeDefaultFuse(runtime, value)

	if dataset.Spec.Owner != nil {
		value.Fuse.Args[len(value.Fuse.Args)-1] = strings.Join([]string{value.Fuse.Args[len(value.Fuse.Args)-1], fmt.Sprintf("uid=%d,gid=%d", *dataset.Spec.Owner.UID, *dataset.Spec.Owner.GID)}, ",")
	} else {
		if len(value.Properties) == 0 {
			value.Properties = map[string]string{}
		}
		value.Properties["alluxio.fuse.user.group.translation.enabled"] = "true"
	}
	// value.Fuse.Args[-1]

	// Allow root: only the RunAs user and root can access fuse
	//if !strings.Contains(value.Fuse.Args[len(value.Fuse.Args)-1], "allow_") {
	//	value.Fuse.Args[len(value.Fuse.Args)-1] = strings.Join([]string{value.Fuse.Args[len(value.Fuse.Args)-1], "allow_root"}, ",")
	//}

	// Allow others: all users(including root) can access fuse
	if !strings.Contains(value.Fuse.Args[len(value.Fuse.Args)-1], "allow_") {
		value.Fuse.Args[len(value.Fuse.Args)-1] = strings.Join([]string{value.Fuse.Args[len(value.Fuse.Args)-1], "allow_other"}, ",")
	}

	value.Fuse.NodeSelector = map[string]string{}

	if runtime.Spec.Fuse.Global {
		value.Fuse.Global = true
		if len(runtime.Spec.Fuse.NodeSelector) > 0 {
			value.Fuse.NodeSelector = runtime.Spec.Fuse.NodeSelector
		}
		value.Fuse.NodeSelector[common.FLUID_FUSE_BALLOON_KEY] = common.FLUID_FUSE_BALLOON_VALUE
		e.Log.Info("Enable Fuse's global mode")
	} else {
		labelName := e.getCommonLabelname()
		value.Fuse.NodeSelector[labelName] = "true"
		e.Log.Info("Disable Fuse's global mode")
	}

	value.Fuse.HostNetwork = true
	value.Fuse.Enabled = true

	e.transformResourcesForFuse(runtime, value)

	return

}
