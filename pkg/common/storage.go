package common

type ReadType string

const (
	HumanReadType ReadType = "h-"

	// rawReadType readType = "raw-"
)

type StorageType string

const (
	MemoryStorageType StorageType = "m-"

	DiskStorageType StorageType = "d-"

	TotalStorageType StorageType = "t-"
)
