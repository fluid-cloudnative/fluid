/*
Copyright 2020 The Fluid Authors.

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
	AlluxioRuntime = "alluxio"

	AlluxioMountType = "fuse.alluxio-fuse"

	AlluxioChart = AlluxioRuntime

	AlluxioEngineImpl = AlluxioRuntime
)

// Constants for Alluxio Images
const (
	DefaultInitImageEnv = "DEFAULT_INIT_IMAGE_ENV"

	AlluxioRuntimeImageEnv = "ALLUXIO_RUNTIME_IMAGE_ENV"

	AlluxioFuseImageEnv = "ALLUXIO_FUSE_IMAGE_ENV"

	DefaultAlluxioRuntimeImage = "registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio:2.3.0-SNAPSHOT-2c41226"

	DefaultAlluxioFuseImage = "registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio-fuse:2.3.0-SNAPSHOT-2c41226"
)
