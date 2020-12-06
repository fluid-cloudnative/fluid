package common

type ReadType string

const (
	HumanReadType ReadType = "human-"

	// rawReadType readType = "raw-"
)

type StorageType string

const (
	MemoryStorageType StorageType = "mem-"

	DiskStorageType StorageType = "disk-"

	TotalStorageType StorageType = "total-"
)
