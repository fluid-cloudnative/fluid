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

package eac

import (
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/tieredstore"
	"strings"
)

//const (
//	EACDefaultPageCapacity    = 32000
//	EACDefaultJournalCapacity = 16000
//	EACDefaultFileCapacity    = 201649
//	EACDefaultVolumeCapacity  = 1
//
//	EACPageSize    = 0
//	EACJournalSize = 0
//	EACFileSize    = 0
//	EACVolumeSize  = 0
//)

func (e *EACEngine) transformMasterTieredStore(runtime *datav1alpha1.EACRuntime,
	value *EAC) error {
	// TODO: set master tiered store according to master properties
	//eacPageCapacity := EACDefaultPageCapacity
	//eacJournalCapacity := EACDefaultJournalCapacity
	//eacFileCapacity := EACDefaultFileCapacity
	//eacVolumeCapacity := EACDefaultVolumeCapacity
	//
	//if val, ok := runtime.Spec.Master.Properties["g_unas_ShmPageCapacity"]; ok {
	//	eacPageCapacity, _ = strconv.Atoi(val)
	//}
	//if val, ok := runtime.Spec.Master.Properties["g_unas_ShmJournalCapacity"]; ok {
	//	eacJournalCapacity, _ = strconv.Atoi(val)
	//}
	//if val, ok := runtime.Spec.Master.Properties["g_unas_ShmFileCapacity"]; ok {
	//	eacFileCapacity, _ = strconv.Atoi(val)
	//}
	//
	//eacShmSize := eacPageCapacity*EACPageSize + eacJournalCapacity*EACJournalSize + eacFileCapacity*EACFileSize + eacVolumeCapacity*EACVolumeSize

	var levels []Level
	levels = append(levels, Level{
		Level:      0,
		Type:       string(common.VolumeTypeEmptyDir),
		Path:       "/dev/shm",
		MediumType: string(common.Memory),
		// Quota:      utils.TransformQuantityToEACUnit(resource.NewQuantity(int64(eacShmSize), resource.DecimalSI)),
	})
	value.Master.TieredStore.Levels = levels

	return nil
}

func (e *EACEngine) transformFuseTieredStore(runtime *datav1alpha1.EACRuntime,
	value *EAC) error {
	// TODO: set fuse tiered store according to fuse properties
	//eacPageCapacity := EACDefaultPageCapacity
	//eacJournalCapacity := EACDefaultJournalCapacity
	//eacFileCapacity := EACDefaultFileCapacity
	//eacVolumeCapacity := EACDefaultVolumeCapacity
	//
	//if val, ok := runtime.Spec.Fuse.Properties["g_unas_ShmPageCapacity"]; ok {
	//	eacPageCapacity, _ = strconv.Atoi(val)
	//}
	//if val, ok := runtime.Spec.Fuse.Properties["g_unas_ShmJournalCapacity"]; ok {
	//	eacJournalCapacity, _ = strconv.Atoi(val)
	//}
	//if val, ok := runtime.Spec.Fuse.Properties["g_unas_ShmFileCapacity"]; ok {
	//	eacFileCapacity, _ = strconv.Atoi(val)
	//}
	//
	//eacShmSize := eacPageCapacity*EACPageSize + eacJournalCapacity*EACJournalSize + eacFileCapacity*EACFileSize + eacVolumeCapacity*EACVolumeSize

	var levels []Level
	levels = append(levels, Level{
		Level:      0,
		Type:       string(common.VolumeTypeEmptyDir),
		Path:       "/dev/shm",
		MediumType: string(common.Memory),
		// Quota:      utils.TransformQuantityToEACUnit(resource.NewQuantity(int64(eacShmSize), resource.DecimalSI)),
	})
	value.Fuse.TieredStore.Levels = levels

	return nil
}

func (e *EACEngine) transformWorkerTieredStore(runtime *datav1alpha1.EACRuntime,
	value *EAC) error {

	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	// Set tieredstore levels
	var levels []Level
	for _, level := range runtimeInfo.GetTieredStoreInfo().Levels {
		l := tieredstore.GetTieredLevel(runtimeInfo, level.MediumType)

		if l != 0 {
			return fmt.Errorf("eac worker only support one level of tiered store")
		}
		if len(level.CachePaths) != 1 {
			return fmt.Errorf("eac worker only support one cache path")
		}

		var paths []string
		var quotas []string
		for _, cachePath := range level.CachePaths {
			paths = append(paths, fmt.Sprintf("%s/%s/%s", cachePath.Path, runtime.Namespace, runtime.Name))
			quotas = append(quotas, utils.TransformQuantityToEACUnit(cachePath.Quota))
		}

		pathConfigStr := strings.Join(paths, ",")
		quotaConfigStr := strings.Join(quotas, ",")
		mediumTypeConfigStr := strings.Join(*utils.FillSliceWithString(string(level.MediumType), len(paths)), ",")

		levels = append(levels, Level{
			Alias:      string(level.MediumType),
			Level:      l,
			Type:       string(level.VolumeType),
			Path:       pathConfigStr,
			MediumType: mediumTypeConfigStr,
			Low:        level.Low,
			High:       level.High,
			Quota:      quotaConfigStr,
		})
	}

	// default worker tiered store
	if len(levels) == 0 {
		levels = append(levels, Level{
			Level:      0,
			Type:       string(common.VolumeTypeEmptyDir),
			Path:       "/cache_dir",
			MediumType: string(common.Memory),
			Quota:      "1GB",
		})
	}

	value.Worker.TieredStore.Levels = levels

	return nil
}
