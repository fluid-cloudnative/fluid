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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
)

// transform dataset which has ufsPaths and ufsVolumes
func (e *GooseFSEngine) transformInitUsers(runtime *datav1alpha1.GooseFSRuntime, value *GooseFS) {

	value.InitUsers = common.InitUsers{
		Enabled: false,
	}

	if runtime.Spec.RunAs != nil {
		value.UserInfo.User = int(*runtime.Spec.RunAs.UID)
		value.UserInfo.Group = int(*runtime.Spec.RunAs.GID)
		value.InitUsers = common.InitUsers{
			Enabled:        true,
			Dir:            e.getInitUserDir(),
			EnvUsers:       utils.GetInitUserEnv(runtime.Spec.RunAs),
			EnvTieredPaths: e.getInitTierPathsEnv(runtime),
		}
	}

	image := runtime.Spec.InitUsers.Image
	tag := runtime.Spec.InitUsers.ImageTag
	imagePullPolicy := runtime.Spec.InitUsers.ImagePullPolicy

	value.InitUsers.Image, value.InitUsers.ImageTag, value.InitUsers.ImagePullPolicy = docker.ParseInitImage(image, tag, imagePullPolicy, common.DefaultInitImageEnv)

	e.Log.Info("Check InitUsers", "InitUsers", value.InitUsers)

}
