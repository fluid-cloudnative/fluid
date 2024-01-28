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

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Runtime Information interface defines the interfaces that should be implemented
// by Alluxio Runtime or other implementation .
// Thread safety is required from implementations of this interface.
type RuntimeInfoInterface interface {
	GetTieredStoreInfo() TieredStoreInfo

	GetName() string

	GetNamespace() string

	GetRuntimeType() string

	// GetStoragetLabelname(read common.ReadType, storage common.StorageType) string

	GetLabelNameForMemory() string

	GetLabelNameForDisk() string

	GetLabelNameForTotal() string

	GetCommonLabelName() string

	GetFuseLabelName() string

	GetRuntimeLabelName() string

	GetDatasetNumLabelName() string

	GetPersistentVolumeName() string

	IsExclusive() bool

	SetupFuseDeployMode(global bool, nodeSelector map[string]string)

	SetupFuseCleanPolicy(policy datav1alpha1.FuseCleanPolicy)

	SetupWithDataset(dataset *datav1alpha1.Dataset)

	GetFuseDeployMode() (global bool, nodeSelector map[string]string)

	GetFuseCleanPolicy() datav1alpha1.FuseCleanPolicy

	SetDeprecatedNodeLabel(deprecated bool)

	IsDeprecatedNodeLabel() bool

	SetDeprecatedPVName(deprecated bool)

	IsDeprecatedPVName() bool

	GetFuseContainerTemplate() (template *common.FuseInjectionTemplate, err error)

	SetClient(client client.Client)

	GetMetadataList() []datav1alpha1.Metadata
}

// The real Runtime Info should implement
type RuntimeInfo struct {
	name        string
	namespace   string
	runtimeType string

	//tieredstore datav1alpha1.TieredStore
	tieredstoreInfo TieredStoreInfo

	exclusive bool
	// Check if the runtime info is already setup by the dataset
	//setup bool

	// Fuse configuration
	fuse Fuse

	// Check if the deprecated node label is used
	deprecatedNodeLabel bool

	// Check if the deprecated PV naming style is used
	deprecatedPVName bool

	client client.Client

	metadataList []datav1alpha1.Metadata
}

type Fuse struct {
	// fuse is deployed in global mode
	Global bool

	NodeSelector map[string]string

	// CleanPolicy decides when to clean fuse pods.
	CleanPolicy datav1alpha1.FuseCleanPolicy
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

func BuildRuntimeInfo(name string,
	namespace string,
	runtimeType string,
	tieredstore datav1alpha1.TieredStore,
	opts ...RuntimeInfoOption) (runtime RuntimeInfoInterface, err error) {

	tieredstoreInfo, err := convertToTieredstoreInfo(tieredstore)
	if err != nil {
		return nil, err
	}

	runtime = &RuntimeInfo{
		name:            name,
		namespace:       namespace,
		runtimeType:     runtimeType,
		tieredstoreInfo: tieredstoreInfo,
	}
	for _, fn := range opts {
		fn(runtime.(*RuntimeInfo))
	}
	return
}

type RuntimeInfoOption func(info *RuntimeInfo)

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

func WithMetadataList(metadataList []datav1alpha1.Metadata) RuntimeInfoOption {
	return func(info *RuntimeInfo) {
		info.metadataList = metadataList
	}
}

func (info *RuntimeInfo) GetMetadataList() []datav1alpha1.Metadata {
	return info.metadataList
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

// GetRuntimeType gets runtime type
func (info *RuntimeInfo) GetRuntimeType() string {
	return info.runtimeType
}

// IsExclusive determines if the runtime is exlusive
func (info *RuntimeInfo) IsExclusive() bool {
	return info.exclusive
}

// SetupWithDataset determines if need to setup with the info of dataset
func (info *RuntimeInfo) SetupWithDataset(dataset *datav1alpha1.Dataset) {
	info.exclusive = dataset.IsExclusiveMode()
}

// SetupFuseDeployMode setups the fuse deploy mode
func (info *RuntimeInfo) SetupFuseDeployMode(global bool, nodeSelector map[string]string) {
	// Since Fluid v0.7.0, global is deprecated.
	info.fuse.Global = true
	info.fuse.NodeSelector = nodeSelector
}

// GetFuseDeployMode gets the fuse deploy mode
func (info *RuntimeInfo) GetFuseDeployMode() (global bool, nodeSelector map[string]string) {
	global = info.fuse.Global
	nodeSelector = info.fuse.NodeSelector
	return
}

func (info *RuntimeInfo) SetupFuseCleanPolicy(policy datav1alpha1.FuseCleanPolicy) {
	if policy == datav1alpha1.NoneCleanPolicy {
		// Default to set the fuse clean policy to OnRuntimeDeleted
		info.fuse.CleanPolicy = datav1alpha1.OnRuntimeDeletedCleanPolicy
		return
	}
	info.fuse.CleanPolicy = policy
}

func (info *RuntimeInfo) GetFuseCleanPolicy() datav1alpha1.FuseCleanPolicy {
	return info.fuse.CleanPolicy
}

// SetDeprecatedNodeLabel set the DeprecatedNodeLabel
func (info *RuntimeInfo) SetDeprecatedNodeLabel(deprecated bool) {
	info.deprecatedNodeLabel = deprecated
}

// IsDeprecatedNodeLabel checks if using deprecated node label
func (info *RuntimeInfo) IsDeprecatedNodeLabel() bool {
	return info.deprecatedNodeLabel
}

func (info *RuntimeInfo) SetDeprecatedPVName(deprecated bool) {
	info.deprecatedPVName = deprecated
}

func (info *RuntimeInfo) IsDeprecatedPVName() bool {
	return info.deprecatedPVName
}

func (info *RuntimeInfo) SetClient(client client.Client) {
	info.client = client
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

// GetRuntimeInfo gets the RuntimeInfo according to name and namespace of it
func GetRuntimeInfo(client client.Client, name, namespace string) (runtimeInfo RuntimeInfoInterface, err error) {
	dataset, err := utils.GetDataset(client, name, namespace)
	if err != nil {
		return runtimeInfo, err
	}

	var runtimeType string
	if len(dataset.Status.Runtimes) != 0 {
		runtimeType = dataset.Status.Runtimes[0].Type
	}
	switch runtimeType {
	case common.AlluxioRuntime:
		alluxioRuntime, err := utils.GetAlluxioRuntime(client, name, namespace)
		if err != nil {
			return runtimeInfo, err
		}
		runtimeInfo, err = BuildRuntimeInfo(name, namespace, common.AlluxioRuntime, datav1alpha1.TieredStore{}, WithMetadataList(GetMetadataListFromAnnotation(alluxioRuntime)))
		if err != nil {
			return runtimeInfo, err
		}
		runtimeInfo.SetupFuseDeployMode(alluxioRuntime.Spec.Fuse.Global, alluxioRuntime.Spec.Fuse.NodeSelector)
		runtimeInfo.SetupFuseCleanPolicy(alluxioRuntime.Spec.Fuse.CleanPolicy)
	case common.JindoRuntime:
		jindoRuntime, err := utils.GetJindoRuntime(client, name, namespace)
		if err != nil {
			return runtimeInfo, err
		}
		runtimeInfo, err = BuildRuntimeInfo(name, namespace, common.JindoRuntime, datav1alpha1.TieredStore{}, WithMetadataList(GetMetadataListFromAnnotation(jindoRuntime)))
		if err != nil {
			return runtimeInfo, err
		}
		runtimeInfo.SetupFuseDeployMode(jindoRuntime.Spec.Fuse.Global, jindoRuntime.Spec.Fuse.NodeSelector)
		runtimeInfo.SetupFuseCleanPolicy(jindoRuntime.Spec.Fuse.CleanPolicy)
	case common.GooseFSRuntime:
		goosefsRuntime, err := utils.GetGooseFSRuntime(client, name, namespace)
		if err != nil {
			return runtimeInfo, err
		}
		runtimeInfo, err = BuildRuntimeInfo(name, namespace, common.GooseFSRuntime, datav1alpha1.TieredStore{}, WithMetadataList(GetMetadataListFromAnnotation(goosefsRuntime)))
		if err != nil {
			return runtimeInfo, err
		}
		runtimeInfo.SetupFuseDeployMode(goosefsRuntime.Spec.Fuse.Global, goosefsRuntime.Spec.Fuse.NodeSelector)
		runtimeInfo.SetupFuseCleanPolicy(goosefsRuntime.Spec.Fuse.CleanPolicy)
	case common.JuiceFSRuntime:
		juicefsRuntime, err := utils.GetJuiceFSRuntime(client, name, namespace)
		if err != nil {
			return runtimeInfo, err
		}
		runtimeInfo, err = BuildRuntimeInfo(name, namespace, common.JuiceFSRuntime, datav1alpha1.TieredStore{}, WithMetadataList(GetMetadataListFromAnnotation(juicefsRuntime)))
		if err != nil {
			return runtimeInfo, err
		}
		runtimeInfo.SetupFuseDeployMode(juicefsRuntime.Spec.Fuse.Global, juicefsRuntime.Spec.Fuse.NodeSelector)
		runtimeInfo.SetupFuseCleanPolicy(juicefsRuntime.Spec.Fuse.CleanPolicy)
	case common.ThinRuntime:
		thinRuntime, err := utils.GetThinRuntime(client, name, namespace)
		if err != nil {
			return runtimeInfo, err
		}
		runtimeInfo, err = BuildRuntimeInfo(name, namespace, common.ThinRuntime, datav1alpha1.TieredStore{}, WithMetadataList(GetMetadataListFromAnnotation(thinRuntime)))
		if err != nil {
			return runtimeInfo, err
		}
		// Fuse global is always set to true
		runtimeInfo.SetupFuseDeployMode(true, thinRuntime.Spec.Fuse.NodeSelector)
		runtimeInfo.SetupFuseCleanPolicy(thinRuntime.Spec.Fuse.CleanPolicy)
	case common.EFCRuntime:
		efcRuntime, err := utils.GetEFCRuntime(client, name, namespace)
		if err != nil {
			return runtimeInfo, err
		}
		runtimeInfo, err = BuildRuntimeInfo(name, namespace, common.EFCRuntime, datav1alpha1.TieredStore{}, WithMetadataList(GetMetadataListFromAnnotation(efcRuntime)))
		if err != nil {
			return runtimeInfo, err
		}
		runtimeInfo.SetupFuseDeployMode(true, efcRuntime.Spec.Fuse.NodeSelector)
		runtimeInfo.SetupFuseCleanPolicy(efcRuntime.Spec.Fuse.CleanPolicy)
	case common.VineyardRuntime:
		vineyardRuntime, err := utils.GetVineyardRuntime(client, name, namespace)
		if err != nil {
			fmt.Println("Get vineyard runtime error")
			return runtimeInfo, err
		}
		runtimeInfo, err = BuildRuntimeInfo(name, namespace, common.VineyardRuntime, datav1alpha1.TieredStore{}, WithMetadataList(GetMetadataListFromAnnotation(vineyardRuntime)))
		if err != nil {
			fmt.Println("Build vineyard runtime error")
			return runtimeInfo, err
		}
		runtimeInfo.SetupFuseDeployMode(common.VineyardFuseIsGlobal, common.VineyardFuseNodeSelector)
		runtimeInfo.SetupFuseCleanPolicy(vineyardRuntime.Spec.Fuse.CleanPolicy)
	default:
		err = fmt.Errorf("fail to get runtimeInfo for runtime type: %s", runtimeType)
		return
	}

	if runtimeInfo != nil {
		runtimeInfo.SetClient(client)
	}
	return runtimeInfo, err
}

func GetRuntimeStatus(client client.Client, runtimeType, name, namespace string) (status *datav1alpha1.RuntimeStatus, err error) {
	switch runtimeType {
	case common.AlluxioRuntime:
		runtime, err := utils.GetAlluxioRuntime(client, name, namespace)
		if err != nil {
			return status, err
		}
		return &runtime.Status, nil
	case common.JindoRuntime:
		runtime, err := utils.GetJindoRuntime(client, name, namespace)
		if err != nil {
			return status, err
		}
		return &runtime.Status, nil
	case common.GooseFSRuntime:
		runtime, err := utils.GetGooseFSRuntime(client, name, namespace)
		if err != nil {
			return status, err
		}
		return &runtime.Status, nil
	case common.JuiceFSRuntime:
		runtime, err := utils.GetJuiceFSRuntime(client, name, namespace)
		if err != nil {
			return status, err
		}
		return &runtime.Status, nil
	case common.EFCRuntime:
		runtime, err := utils.GetEFCRuntime(client, name, namespace)
		if err != nil {
			return status, err
		}
		return &runtime.Status, nil
	case common.ThinRuntime:
		runtime, err := utils.GetThinRuntime(client, name, namespace)
		if err != nil {
			return status, err
		}
		return &runtime.Status, nil
	case common.VineyardRuntime:
		runtime, err := utils.GetVineyardRuntime(client, name, namespace)
		if err != nil {
			return status, err
		}
		return &runtime.Status, nil
	default:
		err = fmt.Errorf("fail to get runtimeInfo for runtime type: %s", runtimeType)
		return nil, err
	}
}
