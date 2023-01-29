package common

// Runtime for Alluxio
const (
	AlluxioRuntime = "alluxio"

	AlluxioMountType = "fuse.alluxio-fuse"

	AlluxioChart = AlluxioRuntime

	DefaultInitImageEnv = "DEFAULT_INIT_IMAGE_ENV"

	AlluxioRuntimeImageEnv = "ALLUXIO_RUNTIME_IMAGE_ENV"

	AlluxioFuseImageEnv = "ALLUXIO_FUSE_IMAGE_ENV"

	DefaultAlluxioRuntimeImage = "alluxio/alluxio-dev:2.9.0"

	DefaultAlluxioFuseImage = "alluxio/alluxio-dev:2.9.0"
)

var (
	// alluxio ufs root path
	AlluxioMountPathFormat = RootDirPath + "%s"

	AlluxioLocalStorageRootPath   = "/underFSStorage"
	AlluxioLocalStoragePathFormat = AlluxioLocalStorageRootPath + "/%s"
)
