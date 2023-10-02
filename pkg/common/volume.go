/*
Copyright 2023 The Fluid Author.

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

const (
	VolumeAttrFluidPath = "fluid_path"

	VolumeAttrFluidSubPath = "fluid_sub_path"

	VolumeAttrMountType = "mount_type"

	VolumeAttrNamespace = "runtime_namespace"

	VolumeAttrName = "runtime_name"

	CSIDriver = "fuse.csi.fluid.io"

	Fluid = "fluid"

	NodePublishMethod = "node_publish_method"

	NodePublishMethodSymlink = "symlink"
)

var (
	FluidStorageClass = Fluid

	FuseContainerName = "fluid-fuse"

	InitFuseContainerName = "init-fluid-fuse"

	FuseMountEnv = "FLUID_FUSE_MOUNTPOINT"
)
