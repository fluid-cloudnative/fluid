package common

// Runtime for Alluxio
const (
	AlluxioRuntime = "alluxio"

	AlluxioMountType = "fuse.alluxio-fuse"

	AlluxioChart = AlluxioRuntime

	DefaultInitImageEnv = "DEFAULT_INIT_IMAGE_ENV"

	AlluxioRuntimeImageEnv = "ALLUXIO_RUNTIME_IMAGE_ENV"

	AlluxioFuseImageEnv = "ALLUXIO_FUSE_IMAGE_ENV"

	DefaultAlluxioRuntimeImage = "registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio:2.3.0-SNAPSHOT-2c41226"

	DefaultAlluxioFuseImage = "registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio-fuse:2.3.0-SNAPSHOT-2c41226"
)

var (
	// alluxio ufs root path
	AlluxioMountPathFormat = RootDirPath + "%s"

	AlluxioLocalStorageRootPath   = "/underFSStorage"
	AlluxioLocalStoragePathFormat = AlluxioLocalStorageRootPath + "/%s"
)
