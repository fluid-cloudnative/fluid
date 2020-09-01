/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

	ErrorHelmInstall = "ErrorHelmInstall"

	DatasetNotReady = "DatasetNotReady"

	RuntimeNotReady = "RuntimeNotReady"

	DataLoadCollision = "DataLoadCollision"

	PrefetchJobStarted = "Prefetch Started"

	PrefetchJobInterrupted = "PrefetchJobInterrupted"

	PrefetchJobComplete = "Prefetch Complete"

	PrefetchJobFailed = "Prefetch Failed"
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

var tieredStoreOrderMap map[MediumType]int = map[MediumType]int{
	Memory: 0,
	SSD:    1,
	HDD:    2,
}

func GetDefaultTieredStoreOrder(MediumType MediumType) (order int) {
	order = tieredStoreOrderMap[MediumType]
	return order
}

type Category string

const (
	AccelerateCategory Category = "Accelerate"
)
