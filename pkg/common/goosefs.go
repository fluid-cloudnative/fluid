package common

// Runtime for Alluxio
const (
	GOOSEFS_RUNTIME = "goosefs"

	GOOSEFS_MOUNT_TYPE = "fuse.goosefs-fuse"

	GOOSEFS_NAMESPACE = "goosefs-system"

	GOOSEFS_CHART = GOOSEFS_RUNTIME

	GOOSEFS_RUNTIME_IMAGE_ENV = "GOOSEFS_RUNTIME_IMAGE_ENV"

	GOOSEFS_FUSE_IMAGE_ENV = "GOOSEFS_FUSE_IMAGE_ENV"

	DEFAULT_GOOSEFS_RUNTIME_IMAGE = "ccr.ccs.tencentyun.com/fluid/goosefs:v1.0.1"

	DEFAULT_GOOSEFS_FUSE_IMAGE = "ccr.ccs.tencentyun.com/fluid/goosefs-fuse:v1.0.1"
)

var (
	// goosefs ufs root path
	GooseFSMountPathFormat = RootDirPath + "%s"

	GooseFSLocalStorageRootPath   = "/underFSStorage"
	GooseFSLocalStoragePathFormat = GooseFSLocalStorageRootPath + "/%s"
)
