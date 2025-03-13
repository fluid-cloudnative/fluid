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

package goosefs

import (
	"fmt"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// 4. Transform the fuse
func (e *GooseFSEngine) transformFuse(runtime *datav1alpha1.GooseFSRuntime, dataset *datav1alpha1.Dataset, value *GooseFS) (err error) {
	value.Fuse = Fuse{}

	image := runtime.Spec.Fuse.Image
	tag := runtime.Spec.Fuse.ImageTag
	imagePullPolicy := runtime.Spec.Fuse.ImagePullPolicy

	value.Fuse.Image, value.Fuse.ImageTag, value.Fuse.ImagePullPolicy = e.parseFuseImage(image, tag, imagePullPolicy)

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
		value.Properties["goosefs.fuse.user.group.translation.enabled"] = "true"
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

	if len(runtime.Spec.Fuse.NodeSelector) > 0 {
		value.Fuse.NodeSelector = runtime.Spec.Fuse.NodeSelector
	} else {
		value.Fuse.NodeSelector = map[string]string{}
	}
	if runtime.Spec.Fuse.LaunchMode != datav1alpha1.EagerMode {
		// The label will be added by CSI Plugin when any workload pod is scheduled on the node.
		value.Fuse.NodeSelector[utils.GetFuseLabelName(runtime.Namespace, runtime.Name, e.runtimeInfo.GetOwnerDatasetUID())] = "true"
	}
	value.Fuse.HostNetwork = true
	value.Fuse.HostPID = common.HostPIDEnabled(runtime.Annotations)
	value.Fuse.Enabled = true

	e.transformResourcesForFuse(runtime, value)

	// set critical fuse pod to avoid eviction
	value.Fuse.CriticalPod = common.CriticalFusePodEnabled()

	// transform the annotation for goosefs fuse.
	value.Fuse.Annotations = runtime.Spec.Fuse.Annotations

	return

}
