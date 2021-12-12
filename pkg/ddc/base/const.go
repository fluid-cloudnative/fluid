package base

type FuseCleanPolicy string

const (
	// OnDemandCleanPolicy cleans fuse pod once th fuse pod on some node is not needed
	OnDemandCleanPolicy FuseCleanPolicy = "OnDemand"

	// OnRuntimeDeletedCleanPolicy cleans fuse pod only when the cache runtime is deleted
	OnRuntimeDeletedCleanPolicy FuseCleanPolicy = "OnRuntimeDeleted"
)
