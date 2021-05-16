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

package base

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Runtime Information interface defines the interfaces that should be implemented
// by Alluxio Runtime or other implementation .
// Thread safety is required from implementations of this interface.
type RuntimeInfoInterface interface {
	//GetTieredstore() datav1alpha1.Tieredstore

	GetTieredstoreInfo() TieredstoreInfo

	GetName() string

	GetNamespace() string

	GetRuntimeType() string

	// GetStoragetLabelname(read common.ReadType, storage common.StorageType) string

	GetLabelnameForMemory() string

	GetLabelnameForDisk() string

	GetLabelnameForTotal() string

	GetCommonLabelname() string

	GetRuntimeLabelname() string

	GetPersistentVolumeName() string

	IsExclusive() bool

	SetupFuseDeployMode(global bool, nodeSelector map[string]string)

	SetupWithDataset(dataset *datav1alpha1.Dataset)

	GetFuseDeployMode() (global bool, nodeSelector map[string]string)

	SetDeprecatedNodeLabel(deprecated bool)

	IsDeprecatedNodeLabel() bool

	SetDeprecatedPVName(deprecated bool)

	IsDeprecatedPVName() bool
}

// The real Runtime Info should implement
type RuntimeInfo struct {
	name        string
	namespace   string
	runtimeType string

	//tieredstore datav1alpha1.Tieredstore
	tieredstoreInfo TieredstoreInfo

	exclusive bool
	// Check if the runtime info is already setup by the dataset
	//setup bool

	// Fuse configuration
	fuse Fuse

	// Check if the deprecated node label is used
	deprecatedNodeLabel bool

	// Check if the deprecated PV naming style is used
	deprecatedPVName bool
}

type Fuse struct {
	// fuse is deployed in global mode
	Global bool

	NodeSelector map[string]string
}

type TieredstoreInfo struct {
	Levels []Level
}

type Level struct {
	MediumType common.MediumType

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
	tieredstore datav1alpha1.Tieredstore) (runtime RuntimeInfoInterface, err error) {

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
	return
}

func (info *RuntimeInfo) GetTieredstoreInfo() TieredstoreInfo {
	return info.tieredstoreInfo
}

// GetTieredstore gets Tieredstore
//func (info *RuntimeInfo) GetTieredstore() datav1alpha1.Tieredstore {
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
	info.fuse.Global = global
	info.fuse.NodeSelector = nodeSelector
}

// GetFuseDeployMode gets the fuse deploy mode
func (info *RuntimeInfo) GetFuseDeployMode() (global bool, nodeSelector map[string]string) {
	global = info.fuse.Global
	nodeSelector = info.fuse.NodeSelector
	return
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

func convertToTieredstoreInfo(tieredstore datav1alpha1.Tieredstore) (TieredstoreInfo, error) {
	tieredstoreInfo := TieredstoreInfo{
		Levels: []Level{},
	}

	for _, level := range tieredstore.Levels {
		paths := strings.Split(level.Path, ",")
		numPaths := len(paths)

		var cachePaths []CachePath

		if len(level.QuotaList) == 0 {
			if level.Quota == nil {
				return TieredstoreInfo{}, fmt.Errorf("either quota or quotaList must be set")
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
				return TieredstoreInfo{}, fmt.Errorf("length of quotaList must be consistent with length of paths")
			}

			for i, quotaStr := range quotaStrs {
				quotaQuantity, err := resource.ParseQuantity(quotaStr)
				if err != nil {
					return TieredstoreInfo{}, fmt.Errorf("can't correctly parse quota \"%s\" to a quantity type", quotaStr)
				}
				cachePaths = append(cachePaths, CachePath{
					Path:  strings.TrimRight(paths[i], "/"),
					Quota: &quotaQuantity,
				})
			}
		}

		tieredstoreInfo.Levels = append(tieredstoreInfo.Levels, Level{
			MediumType: level.MediumType,
			CachePaths: cachePaths,
			High:       level.High,
			Low:        level.Low,
		})
	}
	return tieredstoreInfo, nil
}

// GetRuntimeInfo gets the RuntimeInfo according to name and namespace of it
func GetRuntimeInfo(client client.Client, name, namespace string) (RuntimeInfoInterface, error) {
	dataset, err := utils.GetDataset(client, name, namespace)
	if err != nil {
		return &RuntimeInfo{}, err
	}

	var runtimeType string
	if len(dataset.Status.Runtimes) != 0 {
		runtimeType = dataset.Status.Runtimes[0].Type
	}
	switch runtimeType {
	case "":
		err = fmt.Errorf("fail to get runtime type")
		return &RuntimeInfo{}, err
	case "alluxio":
		runtimeInfo, err := BuildRuntimeInfo(name, namespace, "alluxio", datav1alpha1.Tieredstore{})
		if err != nil {
			return runtimeInfo, err
		}
		alluxioRuntime, err := utils.GetAlluxioRuntime(client, name, namespace)
		if err != nil {
			return runtimeInfo, err
		}
		runtimeInfo.SetupFuseDeployMode(alluxioRuntime.Spec.Fuse.Global, alluxioRuntime.Spec.Fuse.NodeSelector)
		return runtimeInfo, nil
	case "jindo":
		runtimeInfo, err := BuildRuntimeInfo(name, namespace, "jindo", datav1alpha1.Tieredstore{})
		if err != nil {
			return runtimeInfo, err
		}
		jindoRuntime, err := utils.GetJindoRuntime(client, name, namespace)
		if err != nil {
			return runtimeInfo, err
		}
		runtimeInfo.SetupFuseDeployMode(jindoRuntime.Spec.Fuse.Global, jindoRuntime.Spec.Fuse.NodeSelector)
		return runtimeInfo, nil
	default:
		runtimeInfo, err := BuildRuntimeInfo(name, namespace, runtimeType, datav1alpha1.Tieredstore{})
		return runtimeInfo, err
	}
}
