package common

const (
	FluidPath = "fluid_path"

	MountType = "mount_type"

	CSIDriver = "fuse.csi.fluid.io"

	FuseModeKey = "fuse_mode"
)

type FuseMode string

const (
	PodMode FuseMode = "pod"

	ContainerMode FuseMode = "container"
)

func (mode FuseMode) String() string {
	return string(mode)
}

var (
	FLUID_STORAGECLASS = "fluid"
)
