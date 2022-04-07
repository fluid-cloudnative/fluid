package common

const (
	VolumeAttrFluidPath = "fluid_path"

	VolumeAttrMountType = "mount_type"

	VolumeAttrNamespace = "runtime_namespace"

	VolumeAttrName = "runtime_name"

	CSIDriver = "fuse.csi.fluid.io"

	Fluid = "fluid"
)

var (
	FluidStorageClass = Fluid

	FuseContainerName = "fluid-fuse"

	FuseMountEnv = "FLUID_FUSE_MOUNTPOINT"
)
