package common

const (
	// alluxio ufs root path
	RootDirPath        = "/"
	UFSMountPathFormat = RootDirPath + "%s"

	// same for Alluxio and JindoFS
	LocalStorageRootPath   = "/underFSStorage"
	LocalStoragePathFormat = LocalStorageRootPath + "/%s"
)
