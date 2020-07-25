package common

const (
	// LabelAnnotationPrefix is the prefix of every labels and annotations added by the controller.
	LabelAnnotationPrefix = "data.fluid.io/"
	// The format is data.fluid.io/storage-{runtime_type}-{data_set_name}
	LabelAnnotationStorageCapacityPrefix = LabelAnnotationPrefix + "storage-"
	// The dataset annotation
	LabelAnnotationDataset = LabelAnnotationPrefix + "dataset"
)

//Reason for Pillar events
const (
	ErrorProcessDatasetReason = "ErrorProcessDataset"

	ErrorProcessRuntimeReason = "ErrorProcessRuntime"
)

// Runtime for Alluxio
const (
	ALLUXIO_RUNTIME = "alluxio"

	ALLUXIO_NAMESPACE = "alluxio-system"

	ALLUXIO_CHART = ALLUXIO_RUNTIME

	ALLUXIO_DATA_LOADER_IMAGE_ENV = "AlluxioDataLoaderImage"

	DEFAULT_ALLUXIO_DATA_LOADER_IMAGE = "registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio-data-loader:v2.2.0"
)

type CacheStoreType string

const (
	DiskCacheStore CacheStoreType = "Disk"

	MemoryCacheStore CacheStoreType = "Memory"

	NoneCacheStore CacheStoreType = ""
)

const RecommendedKubeConfigPathEnv = "KUBECONFIG"

type MediumType string

const (
	Memory MediumType = "MEM"

	SSD MediumType = "SSD"

	HDD MediumType = "HDD"
)

type Category string

const (
	AccelerateCategory Category = "Accelerate"
)
