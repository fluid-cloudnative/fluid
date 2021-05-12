package common

// Runtime for Alluxio
const (
	ALLUXIO_RUNTIME = "alluxio"

	ALLUXIO_MOUNT_TYPE = "fuse.alluxio-fuse"

	ALLUXIO_NAMESPACE = "alluxio-system"

	ALLUXIO_CHART = ALLUXIO_RUNTIME

	ALLUXIO_DATA_LOADER_IMAGE_ENV = "AlluxioDataLoaderImage"

	// DEFAULT_ALLUXIO_DATA_LOADER_IMAGE = "registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio-data-loader:v2.2.0"

	ALLUXIO_INIT_IMAGE_ENV = "ALLUXIO_INIT_IMAGE_ENV"

	ALLUXIO_RUNTIME_IMAGE_ENV = "ALLUXIO_RUNTIME_IMAGE_ENV"

	ALLUXIO_FUSE_IMAGE_ENV = "ALLUXIO_FUSE_IMAGE_ENV"

	DEFAULT_ALLUXIO_INIT_IMAGE = "registry.cn-hangzhou.aliyuncs.com/fluid/init-users:v0.3.0-1467caa"
)

var (
	// alluxio ufs root path
	AlluxioMountPathFormat = RootDirPath + "%s"

	AlluxioLocalStorageRootPath   = "/underFSStorage"
	AlluxioLocalStoragePathFormat = AlluxioLocalStorageRootPath + "/%s"
)
