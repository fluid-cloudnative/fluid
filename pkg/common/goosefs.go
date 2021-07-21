package common

// Runtime for Alluxio
const (
	GooseFSRuntime = "goosefs"

	GooseFSMountType = "fuse.goosefs-fuse"

	GooseFSNamespace = "goosefs-system"

	GooseFSChart = GooseFSRuntime

	GooseFSRuntimeImageEnv = "GOOSEFS_RUNTIME_IMAGE_ENV"

	GooseFSFuseImageEnv = "GOOSEFS_FUSE_IMAGE_ENV"

	DefaultGooseFSRuntimeImage = "ccr.ccs.tencentyun.com/goosefs/goosefs:v1.0.1"

	DefaultGooseFSFuseImage = "ccr.ccs.tencentyun.com/goosefs/goosefs-fuse:v1.0.1"
)

var (
	// goosefs ufs root path
	GooseFSMountPathFormat = RootDirPath + "%s"

	GooseFSLocalStorageRootPath   = "/underFSStorage"
	GooseFSLocalStoragePathFormat = GooseFSLocalStorageRootPath + "/%s"
)
