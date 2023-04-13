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
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/tieredstore"
)

func (e *EACEngine) transformMasterTieredStore(runtime *datav1alpha1.EFCRuntime,
	value *EAC) error {
	// TODO: set master tiered store quota according to master properties
	// TODO: allow user to config tiered store type

	var levels []Level
	levels = append(levels, Level{
		Level:      0,
		Type:       string(common.VolumeTypeEmptyDir),
		Path:       "/dev/shm",
		MediumType: string(common.Memory),
		// Quota:
	})
	value.Master.TieredStore.Levels = levels

	return nil
}

func (e *EACEngine) transformFuseTieredStore(runtime *datav1alpha1.EFCRuntime,
	value *EAC) error {
	// TODO: set fuse tiered store according to fuse properties
	// TODO: allow user to config tiered store type

	var levels []Level
	levels = append(levels, Level{
		Level:      0,
		Type:       string(common.VolumeTypeEmptyDir),
		Path:       "/dev/shm",
		MediumType: string(common.Memory),
		// Quota:
	})
	value.Fuse.TieredStore.Levels = levels

	return nil
}

func (e *EACEngine) transformWorkerTieredStore(runtime *datav1alpha1.EFCRuntime,
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
			continue
		}
		if len(level.CachePaths) == 0 {
			return fmt.Errorf("eac worker cache path not specificfied")
		}
		level.CachePaths = level.CachePaths[0:1]

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
		levels = append(levels, e.getDefaultTiredStoreLevel0())
	}

	value.Worker.TieredStore.Levels = levels

	return nil
}
