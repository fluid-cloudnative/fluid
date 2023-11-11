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

// Runtime for Juice
const (
	JuiceFSRuntime = "juicefs"

	JuiceFSMountType = "JuiceFS"

	JuiceFSNamespace = "juicefs-system"

	JuiceFSChart = JuiceFSRuntime

	JuiceFSCEImageEnv = "JUICEFS_CE_IMAGE_ENV"
	JuiceFSEEImageEnv = "JUICEFS_EE_IMAGE_ENV"

	DefaultCEImage = "juicedata/juicefs-fuse:ce-v1.1.0-beta2"
	DefaultEEImage = "juicedata/juicefs-fuse:ee-4.9.14"

	NightlyTag = "nightly"

	JuiceFSCeMountPath = "/bin/mount.juicefs"
	JuiceFSMountPath   = "/sbin/mount.juicefs"
	JuiceCeCliPath     = "/usr/local/bin/juicefs"
	JuiceCliPath       = "/usr/bin/juicefs"

	JuiceFSFuseContainer = "juicefs-fuse"

	JuiceFSWorkerContainer = "juicefs-worker"

	JuiceFSDefaultCacheDir = "/var/jfsCache"
)
