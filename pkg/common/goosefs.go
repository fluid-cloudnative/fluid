/*
Copyright 2021 The Fluid Authors.

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

package common

// Runtime for Alluxio
const (
	GooseFSRuntime = "goosefs"

	GooseFSMountType = "fuse.goosefs-fuse"

	GooseFSNamespace = "goosefs-system"

	GooseFSChart = GooseFSRuntime

	GooseFSRuntimeImageEnv = "GOOSEFS_RUNTIME_IMAGE_ENV"

	GooseFSFuseImageEnv = "GOOSEFS_FUSE_IMAGE_ENV"

	DefaultGooseFSRuntimeImage = "ccr.ccs.tencentyun.com/goosefs/goosefs:v1.0.1"

	DefaultGooseFSFuseImage = "ccr.ccs.tencentyun.com/goosefs/goosefs-fuse:v1.0.1"
)

var (
	// goosefs ufs root path
	GooseFSMountPathFormat = RootDirPath + "%s"

	GooseFSLocalStorageRootPath   = "/underFSStorage"
	GooseFSLocalStoragePathFormat = GooseFSLocalStorageRootPath + "/%s"
)
