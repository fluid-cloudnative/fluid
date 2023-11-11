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

package common

// Runtime for Alluxio
const (
	GooseFSRuntime = "goosefs"

	GooseFSMountType = "fuse.goosefs-fuse"

	GooseFSNamespace = "goosefs-system"

	GooseFSChart = GooseFSRuntime

	GooseFSRuntimeImageEnv = "GOOSEFS_RUNTIME_IMAGE_ENV"

	GooseFSFuseImageEnv = "GOOSEFS_FUSE_IMAGE_ENV"

	DefaultGooseFSRuntimeImage = "ccr.ccs.tencentyun.com/qcloud/goosefs:v1.2.0"

	DefaultGooseFSFuseImage = "ccr.ccs.tencentyun.com/qcloud/goosefs-fuse:v1.2.0"
)

var (
	// goosefs ufs root path
	GooseFSMountPathFormat = RootDirPath + "%s"

	GooseFSLocalStorageRootPath   = "/underFSStorage"
	GooseFSLocalStoragePathFormat = GooseFSLocalStorageRootPath + "/%s"
)
