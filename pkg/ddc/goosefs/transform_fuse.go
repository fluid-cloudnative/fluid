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

package goosefs

import (
	"fmt"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
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

	value.Fuse.NodeSelector[e.getFuseLabelname()] = "true"
	value.Fuse.HostNetwork = true
	value.Fuse.Enabled = true

	e.transformResourcesForFuse(runtime, value)

	// set critical fuse pod to avoid eviction
	value.Fuse.CriticalPod = common.CriticalFusePodEnabled()

	// transform the annotation for goosefs fuse.
	value.Fuse.Annotations = runtime.Spec.Fuse.Annotations

	return

}
