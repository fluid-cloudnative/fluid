/*
Copyright 2022 The Fluid Authors.

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

package base

import (
	"encoding/json"
	"fmt"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Conventions defines naming convention for all runtime.
// Conventions includes all the literal string used in Fluid
// to identify the relationship between a dataset(runtime) and its component(master, worker, fuse, persistentVolume, etc.).
type Conventions interface {
	GetPersistentVolumeName() string

	GetLabelNameForMemory() string

	GetLabelNameForDisk() string

	GetLabelNameForTotal() string

	GetCommonLabelName() string

	GetFuseLabelName() string

	GetRuntimeLabelName() string

	GetDatasetNumLabelName() string

	GetExclusiveLabelValue() string
}

// Runtime Information interface defines the interfaces that should be implemented
// by Alluxio Runtime or other implementation .
// Thread safety is required from implementations of this interface.
type RuntimeInfoInterface interface {
	Conventions

	// GetWorkerPods returns the worker object and selector for runtime workers.
	GetWorkerPods(client client.Client) ([]corev1.Pod, error)

	GetTieredStoreInfo() TieredStoreInfo

	GetName() string

	GetNamespace() string

	GetOwnerDatasetUID() string

	GetRuntimeType() string

	GetPlacementModeWithDefault(defaultMode datav1alpha1.PlacementMode) datav1alpha1.PlacementMode

	IsPlacementModeSet() bool

	SetFuseNodeSelector(nodeSelector map[string]string)

	SetFuseName(fuseName string)

	SetupFuseCleanPolicy(policy datav1alpha1.FuseCleanPolicy)

	SetupWithDataset(dataset *datav1alpha1.Dataset)

	SetOwnerDatasetUID(alias types.UID)

	GetFuseNodeSelector() (nodeSelector map[string]string)

	GetFuseName() string

	GetFuseCleanPolicy() datav1alpha1.FuseCleanPolicy

	GetFuseContainerTemplate() (template *common.FuseInjectionTemplate, err error)

	SetAPIReader(apiReader client.Reader)

	GetMetadataList() []datav1alpha1.Metadata

	GetAnnotations() map[string]string

	GetFuseMetricsScrapeTarget() mountModeSelector
}

var _ RuntimeInfoInterface = &RuntimeInfo{}

// The real Runtime Info should implement
type RuntimeInfo struct {
	name      string
	namespace string
	// Use owner dataset's UID as ownerDatasetUID,
	// ownerDatasetUID is used to identify the owner dataset of the runtime
	// when the namespacedName of dataset is over length limit.
	ownerDatasetUID string
	runtimeType     string

	//tieredstore datav1alpha1.TieredStore
	tieredstoreInfo TieredStoreInfo

	placementMode *datav1alpha1.PlacementMode
	// Check if the runtime info is already setup by the dataset
	//setup bool

	// Fuse configuration
	fuse Fuse

	apiReader client.Reader

	annotations map[string]string

	metadataList []datav1alpha1.Metadata
}

type Fuse struct {
	Name string

	NodeSelector map[string]string

	// CleanPolicy decides when to clean fuse pods.
	CleanPolicy datav1alpha1.FuseCleanPolicy

	// Metrics
	MetricsScrapeTarget mountModeSelector
}

type TieredStoreInfo struct {
	Levels []Level
}

type Level struct {
	MediumType common.MediumType

	VolumeType common.VolumeType

	VolumeSource datav1alpha1.VolumeSource

	CachePaths []CachePath

	High string

	Low string
}

type CachePath struct {
	Path string

	Quota *resource.Quantity
}

// BuildRuntimeInfo creates and initializes a RuntimeInfoInterface instance with the specified name, namespace, and runtime type.
// It applies any number of optional configuration functions (opts) to customize the runtime information before returning.
//
// Parameters:
// - name (string): The name of the runtime.
// - namespace (string): The namespace of the runtime.
// - runtimeType (string): The type of the runtime (e.g., Alluxio, JuiceFS).
// - opts (...RuntimeInfoOption): Optional configuration functions that modify the RuntimeInfo struct.
//
// Returns:
// - runtime (RuntimeInfoInterface): A fully configured runtime information object.
// - err (error): Returns an error if any of the provided options fails to apply.
func BuildRuntimeInfo(name string,
	namespace string,
	runtimeType string,
	opts ...RuntimeInfoOption) (runtime RuntimeInfoInterface, err error) {

	runtime = &RuntimeInfo{
		name:        name,
		namespace:   namespace,
		runtimeType: runtimeType,
	}
	for _, fn := range opts {
		if err := fn(runtime.(*RuntimeInfo)); err != nil {
			return nil, errors.Wrapf(err, "fail to build runtime info \"%s/%s\"", namespace, name)
		}
	}
	return
}

type RuntimeInfoOption func(info *RuntimeInfo) error

// GetMetadataListFromAnnotation retrieves the metadata list from the annotations of an object.
// This function is primarily responsible for extracting the "data.fluid.io/metadataList" annotation,
// unmarshaling it into a slice of Metadata structs, and returning the result.
//
// Parameters:
// - accessor (metav1.ObjectMetaAccessor): An accessor to retrieve the object's metadata annotations.
//
// Returns:
// - ret ([]datav1alpha1.Metadata): Returns the unmarshaled metadata list if successful, otherwise an empty slice.
func GetMetadataListFromAnnotation(accessor metav1.ObjectMetaAccessor) (ret []datav1alpha1.Metadata) {
	annotations := accessor.GetObjectMeta().GetAnnotations()
	if annotations == nil {
		return
	}
	m := annotations["data.fluid.io/metadataList"]
	if m == "" {
		return
	}
	if err := json.Unmarshal([]byte(m), &ret); err != nil {
		log := ctrl.Log.WithName("base")
		log.V(5).Error(err, "failed to unmarshal metadataList from annotations", "data.fluid.io/metadataList", m)
	}
	return
}

// WithMetadataList returns a RuntimeInfoOption that sets the metadataList field
// on a RuntimeInfo instance.
//
// Parameters:
//   - metadataList: a slice of Metadata objects to associate with the RuntimeInfo.
//
// Returns:
//   - A RuntimeInfoOption function that, when applied, assigns the provided
//     metadataList to info.metadataList and returns nil (no error).
func WithMetadataList(metadataList []datav1alpha1.Metadata) RuntimeInfoOption {
	return func(info *RuntimeInfo) error {
		// Assign the provided metadataList to the RuntimeInfo instance.
		info.metadataList = metadataList
		return nil
	}
}

// GetMetadataList returns the metadata list of the runtime
func (info *RuntimeInfo) GetMetadataList() []datav1alpha1.Metadata {
	return info.metadataList
}

// WithAnnotations creates a RuntimeInfoOption that sets the annotations field of RuntimeInfo.
// The annotations map usually comes from the Kubernetes runtime object, such as AlluxioRuntime,
// JindoRuntime, GooseFSRuntime, JuiceFSRuntime, ThinRuntime, EFCRuntime, VineyardRuntime, or CacheRuntime.
// These annotations are stored in RuntimeInfo so that other components can later retrieve them through
// GetAnnotations and use them when building, reconciling, or processing runtime-related resources.
// The function itself does not validate, copy, or transform the input annotations map; it directly assigns
// the provided map to RuntimeInfo.annotations.
//
// Parameters:
//   - annotations: A map containing annotation key-value pairs associated with the runtime object.
//
// Returns:
//   - RuntimeInfoOption: An option function that writes the given annotations into RuntimeInfo.
func WithAnnotations(annotations map[string]string) RuntimeInfoOption {
	return func(info *RuntimeInfo) error {
		info.annotations = annotations
		return nil
	}
}

// GetAnnotations returns the annotations map associated with the runtime.
//
// The annotations are typically copied from the underlying Kubernetes runtime
// object (for example, AlluxioRuntime, JindoRuntime, etc.) via the
// `WithAnnotations` option when building the `RuntimeInfo`.
//
// Returns:
//   - map[string]string: the annotations map stored on `RuntimeInfo`. The
//     returned map is not deep-copied; callers should avoid modifying it if
//     it may be shared.
func (info *RuntimeInfo) GetAnnotations() map[string]string {
	return info.annotations
}

// WithClientMetrics sets the client metrics for the RuntimeInfo.
// This function returns a RuntimeInfoOption that configures how client metrics
// are collected, including setting a default scrape target if none is provided
// and parsing the scrape target string into a valid selector.
//
// Parameters:
//   - clientMetrics (datav1alpha1.ClientMetrics): The client metrics configuration to be applied.
//
// Returns:
//   - (RuntimeInfoOption): A function that updates the RuntimeInfo with the specified metrics configuration.
func WithClientMetrics(clientMetrics datav1alpha1.ClientMetrics) RuntimeInfoOption {
	return func(info *RuntimeInfo) error {
		if len(clientMetrics.ScrapeTarget) == 0 {
			// When scrape target is not set, default it to None
			clientMetrics.ScrapeTarget = MountModeSelectNone
		}
		metricsScrapeTarget, err := ParseMountModeSelectorFromStr(clientMetrics.ScrapeTarget)
		if err != nil {
			return err
		}
		info.fuse.MetricsScrapeTarget = metricsScrapeTarget
		return nil
	}
}

func (info *RuntimeInfo) GetFuseMetricsScrapeTarget() mountModeSelector {
	return info.fuse.MetricsScrapeTarget
}

// WithTieredStore converts datav1alpha1.TieredStore to TieredStoreInfo and sets it to RuntimeInfo
// The conversion is needed because datav1alpha1.TieredStore contains some fields in string type which are not convenient to use, such as Quota and QuotaList, and we want to convert them to more structured type in RuntimeInfo.
// The conversion logic is as follows:
// 1. If the length of Levels in datav1alpha1.TieredStore is 0, return an empty TieredStoreInfo.
// 2. For each level in datav1alpha1.TieredStore, split the Path by comma to get multiple paths.
// 3. If QuotaList is not set, divide Quota equally to multiple paths and set the quota for each path.
// 4. If QuotaList is set, split the QuotaList by comma to get multiple quotas, and set the corresponding quota for each path. The value in QuotaList will overwrite the value in Quota if both of them are set.
// 5. Set other fields in TieredStoreInfo according to the values in datav1alpha1.TieredStore.
func WithTieredStore(tieredStore datav1alpha1.TieredStore) RuntimeInfoOption {
	return func(info *RuntimeInfo) error {
		tieredStoreInfo, err := convertToTieredstoreInfo(tieredStore)
		if err != nil {
			return err
		}
		info.tieredstoreInfo = tieredStoreInfo
		return nil
	}
}

func (info *RuntimeInfo) GetTieredStoreInfo() TieredStoreInfo {
	return info.tieredstoreInfo
}

// GetTieredstore gets TieredStore
//func (info *RuntimeInfo) GetTieredstore() datav1alpha1.TieredStore {
//	return info.tieredstore
//}

// GetName gets name
func (info *RuntimeInfo) GetName() string {
	return info.name
}

// GetNamespace gets namespace
func (info *RuntimeInfo) GetNamespace() string {
	return info.namespace
}

// GetOwnerDatasetUID retrieves the unique identifier (UID) of the owner Dataset.
func (info *RuntimeInfo) GetOwnerDatasetUID() string {
	return info.ownerDatasetUID
}

// GetRuntimeType gets runtime type
func (info *RuntimeInfo) GetRuntimeType() string {
	return info.runtimeType
}

// IsPlacementModeSet reports whether the placement mode has been explicitly configured
// on the RuntimeInfo. It returns true if the internal placementMode pointer is not nil,
// indicating that a placement mode (e.g., ExclusiveMode) has been set; otherwise, it returns false.
func (info *RuntimeInfo) IsPlacementModeSet() bool {
	return info.placementMode != nil
}

// GetPlacementModeWithDefault returns the configured placement mode, or the default value if it is not set.
func (info *RuntimeInfo) GetPlacementModeWithDefault(defaultMode datav1alpha1.PlacementMode) datav1alpha1.PlacementMode {
	if !info.IsPlacementModeSet() || info.placementMode == nil {
		return defaultMode
	}

	return *info.placementMode
}

// SetupWithDataset determines if need to setup with the info of dataset
func (info *RuntimeInfo) SetupWithDataset(dataset *datav1alpha1.Dataset) {
	var placementMode datav1alpha1.PlacementMode = dataset.Spec.PlacementMode
	if placementMode == datav1alpha1.DefaultMode {
		placementMode = datav1alpha1.ExclusiveMode
	}
	info.placementMode = &placementMode
}

// SetupWithDataset determines if need to setup with the info of dataset
func (info *RuntimeInfo) SetOwnerDatasetUID(datasetUID types.UID) {
	if datasetUID == "" {
		return
	}
	info.ownerDatasetUID = string(datasetUID)
}

// SetFuseNodeSelector setups the fuse deploy mode
func (info *RuntimeInfo) SetFuseNodeSelector(nodeSelector map[string]string) {
	info.fuse.NodeSelector = nodeSelector
}

// GetFuseNodeSelector gets the fuse deploy mode
func (info *RuntimeInfo) GetFuseName() string {
	return info.fuse.Name
}

// SetFuseNodeSelector setups the fuse deploy mode
func (info *RuntimeInfo) SetFuseName(fuseName string) {
	info.fuse.Name = fuseName
}

// GetFuseNodeSelector gets the fuse deploy mode
func (info *RuntimeInfo) GetFuseNodeSelector() (nodeSelector map[string]string) {
	nodeSelector = info.fuse.NodeSelector
	return
}

// SetupFuseCleanPolicy sets the clean policy for the fuse runtime.
// If the provided policy is NoneCleanPolicy, it defaults to OnRuntimeDeletedCleanPolicy.
// Otherwise, it assigns the given policy directly to the fuse runtime.
func (info *RuntimeInfo) SetupFuseCleanPolicy(policy datav1alpha1.FuseCleanPolicy) {
	if policy == datav1alpha1.NoneCleanPolicy {
		// Default to set the fuse clean policy to OnRuntimeDeleted
		info.fuse.CleanPolicy = datav1alpha1.OnRuntimeDeletedCleanPolicy
		return
	}
	info.fuse.CleanPolicy = policy
}

func (info *RuntimeInfo) GetWorkerPods(client client.Client) ([]corev1.Pod, error) {
	workers, err := kubeclient.GetStatefulSet(client, info.GetWorkerStatefulsetName(), info.GetNamespace())
	if err != nil {
		return nil, err
	}
	workerSelector, err := metav1.LabelSelectorAsSelector(workers.Spec.Selector)
	if err != nil {
		return nil, err
	}

	workerPods, err := kubeclient.GetPodsForStatefulSet(client, workers, workerSelector)

	return workerPods, err
}

func (info *RuntimeInfo) GetFuseCleanPolicy() datav1alpha1.FuseCleanPolicy {
	return info.fuse.CleanPolicy
}

// SetAPIReader sets the API reader for the runtime information.
func (info *RuntimeInfo) SetAPIReader(apiReader client.Reader) {
	info.apiReader = apiReader
}

func convertToTieredstoreInfo(tieredstore datav1alpha1.TieredStore) (TieredStoreInfo, error) {
	if len(tieredstore.Levels) == 0 {
		return TieredStoreInfo{}, nil
	}

	tieredstoreInfo := TieredStoreInfo{
		Levels: []Level{},
	}

	for _, level := range tieredstore.Levels {
		paths := strings.Split(level.Path, ",")
		numPaths := len(paths)

		var cachePaths []CachePath

		if len(level.QuotaList) == 0 {
			if level.Quota == nil {
				return TieredStoreInfo{}, fmt.Errorf("either quota or quotaList must be set")
			}
			// Only quota is set, divide quota equally to multiple paths
			avgQuota := resource.NewQuantity(level.Quota.Value()/int64(numPaths), resource.BinarySI)
			for _, path := range paths {
				pathQuota := avgQuota.DeepCopy()
				cachePaths = append(cachePaths, CachePath{
					Path:  strings.TrimRight(path, "/"),
					Quota: &pathQuota,
				})
			}
		} else {
			// quotaList will overwrite any value set in quota
			quotaStrs := strings.Split(level.QuotaList, ",")
			numQuotas := len(quotaStrs)
			if numQuotas != numPaths {
				return TieredStoreInfo{}, fmt.Errorf("length of quotaList must be consistent with length of paths")
			}

			for i, quotaStr := range quotaStrs {
				quotaQuantity, err := resource.ParseQuantity(quotaStr)
				if err != nil {
					return TieredStoreInfo{}, fmt.Errorf("can't correctly parse quota \"%s\" to a quantity type", quotaStr)
				}
				cachePaths = append(cachePaths, CachePath{
					Path:  strings.TrimRight(paths[i], "/"),
					Quota: &quotaQuantity,
				})
			}
		}

		tieredstoreInfo.Levels = append(tieredstoreInfo.Levels, Level{
			MediumType:   level.MediumType,
			VolumeType:   level.VolumeType,
			VolumeSource: level.VolumeSource,
			CachePaths:   cachePaths,
			High:         level.High,
			Low:          level.Low,
		})
	}
	return tieredstoreInfo, nil
}

// GetRuntimeInfo gets the RuntimeInfo according to name and namespace of it, must be called after dataset bound.
func GetRuntimeInfo(reader client.Reader, name, namespace string) (runtimeInfo RuntimeInfoInterface, err error) {
	dataset, err := utils.GetDataset(reader, name, namespace)
	if err != nil {
		return runtimeInfo, err
	}

	var runtimeType string
	if len(dataset.Status.Runtimes) != 0 {
		runtimeType = dataset.Status.Runtimes[0].Type
	}
	switch runtimeType {
	case common.AlluxioRuntime:
		alluxioRuntime, err := utils.GetAlluxioRuntime(reader, name, namespace)
		if err != nil {
			return runtimeInfo, err
		}
		opts := []RuntimeInfoOption{
			WithTieredStore(datav1alpha1.TieredStore{}),
			WithMetadataList(GetMetadataListFromAnnotation(alluxioRuntime)),
			WithAnnotations(alluxioRuntime.Annotations),
		}
		runtimeInfo, err = BuildRuntimeInfo(name, namespace, common.AlluxioRuntime, opts...)
		if err != nil {
			return runtimeInfo, err
		}
		runtimeInfo.SetFuseNodeSelector(alluxioRuntime.Spec.Fuse.NodeSelector)
		runtimeInfo.SetupFuseCleanPolicy(alluxioRuntime.Spec.Fuse.CleanPolicy)
	case common.JindoRuntime:
		jindoRuntime, err := utils.GetJindoRuntime(reader, name, namespace)
		if err != nil {
			return runtimeInfo, err
		}
		opts := []RuntimeInfoOption{
			WithTieredStore(datav1alpha1.TieredStore{}),
			WithMetadataList(GetMetadataListFromAnnotation(jindoRuntime)),
			WithClientMetrics(jindoRuntime.Spec.Fuse.Metrics),
			WithAnnotations(jindoRuntime.Annotations),
		}
		runtimeInfo, err = BuildRuntimeInfo(name, namespace, common.JindoRuntime, opts...)
		if err != nil {
			return runtimeInfo, err
		}
		runtimeInfo.SetFuseNodeSelector(jindoRuntime.Spec.Fuse.NodeSelector)
		runtimeInfo.SetupFuseCleanPolicy(jindoRuntime.Spec.Fuse.CleanPolicy)
	case common.JuiceFSRuntime:
		juicefsRuntime, err := utils.GetJuiceFSRuntime(reader, name, namespace)
		if err != nil {
			return runtimeInfo, err
		}
		opts := []RuntimeInfoOption{
			WithTieredStore(datav1alpha1.TieredStore{}),
			WithMetadataList(GetMetadataListFromAnnotation(juicefsRuntime)),
			WithAnnotations(juicefsRuntime.Annotations),
		}
		runtimeInfo, err = BuildRuntimeInfo(name, namespace, common.JuiceFSRuntime, opts...)
		if err != nil {
			return runtimeInfo, err
		}
		runtimeInfo.SetFuseNodeSelector(juicefsRuntime.Spec.Fuse.NodeSelector)
		runtimeInfo.SetupFuseCleanPolicy(juicefsRuntime.Spec.Fuse.CleanPolicy)
	case common.ThinRuntime:
		thinRuntime, err := utils.GetThinRuntime(reader, name, namespace)
		if err != nil {
			return runtimeInfo, err
		}
		opts := []RuntimeInfoOption{
			WithTieredStore(datav1alpha1.TieredStore{}),
			WithMetadataList(GetMetadataListFromAnnotation(thinRuntime)),
			WithAnnotations(thinRuntime.Annotations),
		}
		runtimeInfo, err = BuildRuntimeInfo(name, namespace, common.ThinRuntime, opts...)
		if err != nil {
			return runtimeInfo, err
		}
		runtimeInfo.SetFuseNodeSelector(thinRuntime.Spec.Fuse.NodeSelector)
		runtimeInfo.SetupFuseCleanPolicy(thinRuntime.Spec.Fuse.CleanPolicy)
	case common.EFCRuntime:
		efcRuntime, err := utils.GetEFCRuntime(reader, name, namespace)
		if err != nil {
			return runtimeInfo, err
		}
		opts := []RuntimeInfoOption{
			WithTieredStore(datav1alpha1.TieredStore{}),
			WithMetadataList(GetMetadataListFromAnnotation(efcRuntime)),
			WithAnnotations(efcRuntime.Annotations),
		}
		runtimeInfo, err = BuildRuntimeInfo(name, namespace, common.EFCRuntime, opts...)
		if err != nil {
			return runtimeInfo, err
		}
		runtimeInfo.SetFuseNodeSelector(efcRuntime.Spec.Fuse.NodeSelector)
		runtimeInfo.SetupFuseCleanPolicy(efcRuntime.Spec.Fuse.CleanPolicy)
	case common.VineyardRuntime:
		vineyardRuntime, err := utils.GetVineyardRuntime(reader, name, namespace)
		if err != nil {
			return runtimeInfo, err
		}
		opts := []RuntimeInfoOption{
			WithTieredStore(datav1alpha1.TieredStore{}),
			WithMetadataList(GetMetadataListFromAnnotation(vineyardRuntime)),
			WithAnnotations(vineyardRuntime.Annotations),
		}
		runtimeInfo, err = BuildRuntimeInfo(name, namespace, common.VineyardRuntime, opts...)
		if err != nil {
			return runtimeInfo, err
		}
		runtimeInfo.SetFuseNodeSelector(common.VineyardFuseNodeSelector)
		runtimeInfo.SetupFuseCleanPolicy(vineyardRuntime.Spec.Fuse.CleanPolicy)
	case common.CacheRuntime:
		cacheRuntime, err := utils.GetCacheRuntime(reader, name, namespace)
		if err != nil {
			return runtimeInfo, err
		}
		opts := []RuntimeInfoOption{
			WithTieredStore(datav1alpha1.TieredStore{}),
			WithMetadataList(GetMetadataListFromAnnotation(cacheRuntime)),
			WithAnnotations(cacheRuntime.Annotations),
		}
		runtimeInfo, err = BuildRuntimeInfo(name, namespace, common.CacheRuntime, opts...)
		if err != nil {
			return runtimeInfo, err
		}
		runtimeInfo.SetFuseNodeSelector(cacheRuntime.Spec.Client.NodeSelector)
		runtimeInfo.SetupFuseCleanPolicy(cacheRuntime.Spec.Client.CleanPolicy)
		// TODO(cache runtime): is this common logic for all runtimes? If so, move to below 'SetOwnerDatasetUID' line.
		runtimeInfo.SetupWithDataset(dataset)
	default:
		err = fmt.Errorf("fail to get runtimeInfo for runtime type: %s", runtimeType)
		return
	}

	// set fuse name
	var fuseName string
	switch runtimeType {
	case common.JindoRuntime:
		fuseName = name + "-" + common.JindoChartName + "-fuse"
	case common.CacheRuntime:
		fuseName = common.GetCacheComponentName(name, common.ComponentTypeClient)
	default:
		fuseName = name + "-fuse"
	}

	if runtimeInfo != nil {
		runtimeInfo.SetAPIReader(reader)
		runtimeInfo.SetOwnerDatasetUID(dataset.UID)
		runtimeInfo.SetFuseName(fuseName)
	}
	return runtimeInfo, err
}

// RuntimeStatusAccessor provides a unified interface to access common status fields across different runtime types
type RuntimeStatusAccessor interface {
	// GetCacheAffinity returns the cache affinity from the runtime status
	GetCacheAffinity() (*corev1.NodeAffinity, error)
}

// GetRuntimeStatusAccessor returns a unified status accessor for the given runtime
func GetRuntimeStatusAccessor(client client.Client, runtimeType, name, namespace string) (RuntimeStatusAccessor, error) {
	switch runtimeType {
	case common.AlluxioRuntime, common.JindoRuntime,
		common.JuiceFSRuntime, common.EFCRuntime, common.ThinRuntime, common.VineyardRuntime:
		status, err := GetDDCRuntimeStatus(client, runtimeType, name, namespace)
		if err != nil {
			return nil, err
		}
		return &DDCRuntimeStatusAccessor{status: status}, nil
	case common.CacheRuntime:
		runtime, err := utils.GetCacheRuntime(client, name, namespace)
		if err != nil {
			return nil, err
		}
		return &CacheRuntimeStatusAccessor{status: &runtime.Status}, nil
	default:
		return nil, fmt.Errorf("fail to get runtime status accessor for runtime type: %s", runtimeType)
	}
}

// GetDDCRuntimeStatus retrieves the runtime object based on runtime type for DDC-based runtimes
func GetDDCRuntimeStatus(client client.Client, runtimeType, name, namespace string) (*datav1alpha1.RuntimeStatus, error) {
	switch runtimeType {
	case common.AlluxioRuntime:
		runtime, err := utils.GetAlluxioRuntime(client, name, namespace)
		if err != nil {
			return nil, err
		}
		return &runtime.Status, nil
	case common.JindoRuntime:
		runtime, err := utils.GetJindoRuntime(client, name, namespace)
		if err != nil {
			return nil, err
		}
		return &runtime.Status, nil
	case common.JuiceFSRuntime:
		runtime, err := utils.GetJuiceFSRuntime(client, name, namespace)
		if err != nil {
			return nil, err
		}
		return &runtime.Status, nil
	case common.EFCRuntime:
		runtime, err := utils.GetEFCRuntime(client, name, namespace)
		if err != nil {
			return nil, err
		}
		return &runtime.Status, nil
	case common.ThinRuntime:
		runtime, err := utils.GetThinRuntime(client, name, namespace)
		if err != nil {
			return nil, err
		}
		return &runtime.Status, nil
	case common.VineyardRuntime:
		runtime, err := utils.GetVineyardRuntime(client, name, namespace)
		if err != nil {
			return nil, err
		}
		return &runtime.Status, nil
	default:
		return nil, fmt.Errorf("unsupported DDC runtime type: %s", runtimeType)
	}
}

// DDCRuntimeStatusAccessor implements RuntimeStatusAccessor for DDC-based runtimes (Alluxio, Jindo, GooseFS, etc.)
type DDCRuntimeStatusAccessor struct {
	status *datav1alpha1.RuntimeStatus
}

func (d *DDCRuntimeStatusAccessor) GetCacheAffinity() (*corev1.NodeAffinity, error) {
	return d.status.CacheAffinity, nil
}

// CacheRuntimeStatusAccessor implements RuntimeStatusAccessor for CacheRuntime
type CacheRuntimeStatusAccessor struct {
	status *datav1alpha1.CacheRuntimeStatus
}

func (c *CacheRuntimeStatusAccessor) GetCacheAffinity() (*corev1.NodeAffinity, error) {
	return c.status.CacheAffinity, nil
}
