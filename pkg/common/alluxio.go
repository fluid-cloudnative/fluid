package common

// Runtime for Alluxio
const (
	ALLUXIO_RUNTIME = "alluxio"

	ALLUXIO_MOUNT_TYPE = "fuse.alluxio-fuse"

	ALLUXIO_NAMESPACE = "alluxio-system"

	ALLUXIO_CHART = ALLUXIO_RUNTIME

	DEFAULT_INIT_IMAGE_ENV = "DEFAULT_INIT_IMAGE_ENV"

	ALLUXIO_RUNTIME_IMAGE_ENV = "ALLUXIO_RUNTIME_IMAGE_ENV"

	ALLUXIO_FUSE_IMAGE_ENV = "ALLUXIO_FUSE_IMAGE_ENV"

	DEFAULT_ALLUXIO_RUNTIME_IMAGE = "registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio:2.3.0-SNAPSHOT-2c41226"

	DEFAULT_ALLUXIO_FUSE_IMAGE = "registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio-fuse:2.3.0-SNAPSHOT-2c41226"
)

var (
	// alluxio ufs root path
	AlluxioMountPathFormat = RootDirPath + "%s"

	AlluxioLocalStorageRootPath   = "/underFSStorage"
	AlluxioLocalStoragePathFormat = AlluxioLocalStorageRootPath + "/%s"
)
