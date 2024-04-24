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

// Fluid events related to datasets/runtimes
const (
	ErrorCreateDataset = "ErrorCreateDataset"

	ErrorProcessDatasetReason = "ErrorProcessDataset"

	ErrorDeleteDataset = "ErrorDeleteDataset"

	ErrorProcessRuntimeReason = "ErrorProcessRuntime"

	ErrorHelmInstall = "ErrorHelmInstall"

	RuntimeScaleInFailed = "RuntimeScaleInFailed"

	Succeed = "Succeed"

	FuseRecoverFailed = "FuseRecoverFailed"

	FuseRecoverSucceed = "FuseRecoverSucceed"

	FuseUmountDuplicate = "UnmountDuplicateMountpoint"

	RuntimeDeprecated = "RuntimeDeprecated"

	RuntimeWithSecretNotSupported = "RuntimeWithSecretNotSupported"
)

// Events related to all type of Data Operations
const (
	TargetDatasetNotFound = "TargetDatasetNotFound"

	TargetDatasetNotReady = "TargetDatasetNotReady"

	TargetDatasetNamespaceNotSame = "TargetDatasetNamespaceNotSame"

	DataOperationNotSupport = "DataOperationNotSupport"

	DataOperationExecutionFailed = "DataOperationExecutionFailed"

	DataOperationFailed = "DataOperationFailed"

	DataOperationSucceed = "DataOperationSucceed"

	DataOperationNotValid = "DataOperationNotValid"

	DataOperationCollision = "DataOperationCollision"

	TargetSSHSecretNameNotSet = "TargetSSHSecretNameNotSet"
)

// Events related to dataflow
const (
	DataOperationNotFound = "DataOperationNotFound"

	DataOperationWaiting = "DataOperationWaiting"
)

// Events related to DataLoad
const (
	DataLoadCollision = "DataLoadCollision"

	RuntimeNotReady = "RuntimeNotReady"

	DataLoadJobStarted = "DataLoadJobStarted"

	DataLoadJobFailed = "DataLoadJobFailed"

	DataLoadJobComplete = "DataLoadJobComplete"
)

// Events related to DataMigrate
const (
	DataMigrateCollision = "DataMigrateCollision"

	DataMigrateJobStarted = "DataMigrateJobStarted"

	DataMigrateJobFailed = "DataMigrateJobFailed"

	DataMigrateJobComplete = "DataMigrateJobComplete"
)

// Events related to DataProcess
const (
	DataProcessProcessorNotSpecified = "ProcessorNotSpecified"

	DataProcessMultipleProcessorSpecified = "MultipleProcessorSpecified"

	DataProcessConflictMountPath = "ConflictMountPath"
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

type VolumeType string

const (
	VolumeTypeDefault        VolumeType = ""
	VolumeTypeHostPath       VolumeType = "hostPath"
	VolumeTypeEmptyDir       VolumeType = "emptyDir"
	VolumeTypeVolumeTemplate VolumeType = "volumeTemplate"
)

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
	EnvFuseDeviceResourceName     string = "VFUSE_RESOURCE_NAME"
	DefaultFuseDeviceResourceName string = "fluid.io/fuse"
)

const (
	RootDirPath            = "/"
	DefaultImagePullPolicy = "IfNotPresent"
	MyPodNamespace         = "MY_POD_NAMESPACE"
	True                   = "true"
	False                  = "false"
	App                    = "app"
	JobPolicy              = "fluid.io/jobPolicy"
	CronPolicy             = "cron"
	PodRoleType            = "role"
	DataloadPod            = "dataload-pod"
	NamespaceFluidSystem   = "fluid-system"
)

const (
	inject                        = ".fluid.io/inject"
	injectSidecar                 = ".sidecar" + inject
	InjectServerless              = "serverless" + inject             // serverless.fluid.io/inject
	InjectUnprivilegedFuseSidecar = "unprivileged" + injectSidecar    // unprivileged.sidecar.fluid.io/inject
	InjectCacheDir                = "cachedir" + injectSidecar        // cachedir.sidecar.fluid.io/inject
	InjectWorkerSidecar           = "worker" + injectSidecar          // worker.sidecar.fluid.io/inject
	InjectSidecarDone             = "done" + injectSidecar            // done.sidecar.fluid.io/inject
	InjectAppPostStart            = "app.poststart" + inject          // app.poststart.fluid.io/inject
	InjectSidecarPostStart        = "fuse.sidecar.poststart" + inject // fuse.sidecar.poststart.fluid.io/inject

	injectServerful     = ".serverful" + inject
	InjectServerfulFuse = "fuse" + injectServerful

	InjectFuseSidecar = "fuse" + injectSidecar // [Deprecated] fuse.sidecar.fluid.io/inject
)

const (
	EnvServerlessPlatformKey        = "KEY_SERVERLESS_PLATFORM"
	EnvServerlessPlatformVal        = "VALUE_SERVERLESS_PLATFORM"
	EnvDisableApplicationController = "KEY_DISABLE_APP_CONTROLLER"
	EnvImagePullSecretsKey          = "IMAGE_PULL_SECRETS"
)

const (
	RuntimeFuseHostPIDKey = "runtime.fluid.io/fuse.hostpid"
)
