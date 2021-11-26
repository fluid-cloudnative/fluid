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
	JINDO_RUNTIME = "jindo"

	JINDO_MOUNT_TYPE = "fuse.jindofs-fuse"

	JINDO_SMARTDATA_IMAGE_ENV = "JINDO_SMARTDATA_IMAGE_ENV"

	JINDO_FUSE_IMAGE_ENV = "JINDO_FUSE_IMAGE_ENV"

	DEFAULT_JINDO_RUNTIME_IMAGE = "registry.cn-shanghai.aliyuncs.com/jindofs/smartdata:3.7.0"

	JINDO_DNS_SERVER = "JINDO_DNS_SERVER_ENV"
)
