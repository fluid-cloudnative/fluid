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

//Reason for Fluid events
const (
	ErrorProcessDatasetReason = "ErrorProcessDataset"

	ErrorDeleteDataset = "ErrorDeleteDataset"

	ErrorProcessRuntimeReason = "ErrorProcessRuntime"

	ErrorHelmInstall = "ErrorHelmInstall"

	TargetDatasetNotFound = "TargetDatasetNotFound"

	TargetDatasetPathNotFound = "TargetDatasetPathNotFound"

	TargetDatasetNotReady = "TargetDatasetNotReady"

	TargetDatasetNamespaceNotSame = "TargetDatasetNamespaceNotSame"

	DataLoadCollision = "DataLoadCollision"

	RuntimeNotReady = "RuntimeNotReady"

	DataLoadJobStarted = "DataLoadJobStarted"

	DataLoadJobFailed = "DataLoadJobFailed"

	DataLoadJobComplete = "DataLoadJobComplete"

	DataBackupFailed = "DataBackupFailed"

	DataBackupComplete = "DataBackupComplete"

	RuntimeScaleInFailed = "RuntimeScaleInFailed"

	Succeed = "Succeed"
)

type CacheStoreType string

const (
	DiskCacheStore CacheStoreType = "Disk"

	MemoryCacheStore CacheStoreType = "Memory"
)

const RecommendedKubeConfigPathEnv = "KUBECONFIG"

type MediumType string

const (
	Memory MediumType = "MEM"

	SSD MediumType = "SSD"

	HDD MediumType = "HDD"
)

var tieredStoreOrderMap = map[MediumType]int{
	Memory: 0,
	SSD:    1,
	HDD:    2,
}

// GetDefaultTieredStoreOrder get the TieredStoreOrder from the default Map
// because the crd has validated the value, It's not possible to meet unknown MediumType
func GetDefaultTieredStoreOrder(MediumType MediumType) (order int) {
	order = tieredStoreOrderMap[MediumType]
	return order
}

type Category string

const (
	AccelerateCategory Category = "Accelerate"
)

var (
	ExpectedFluidAnnotations = map[string]string{
		"CreatedBy": "fluid",
	}
)

const (
	FluidExclusiveKey string = "fluid_exclusive"
)

const (
	RootDirPath            = "/"
	DefaultImagePullPolicy = "IfNotPresent"
	MyPodNamespace         = "MY_POD_NAMESPACE"
	True                   = "true"
	False                  = "false"
)
