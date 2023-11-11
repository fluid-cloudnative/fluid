/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package efc

import (
	"fmt"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/tieredstore"
)

func (e *EFCEngine) transformMasterTieredStore(runtime *datav1alpha1.EFCRuntime,
	value *EFC) error {
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

func (e *EFCEngine) transformFuseTieredStore(runtime *datav1alpha1.EFCRuntime,
	value *EFC) error {
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

func (e *EFCEngine) transformWorkerTieredStore(runtime *datav1alpha1.EFCRuntime,
	value *EFC) error {

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
			return fmt.Errorf("efc worker cache path not specificfied")
		}
		level.CachePaths = level.CachePaths[0:1]

		var paths []string
		var quotas []string
		for _, cachePath := range level.CachePaths {
			paths = append(paths, fmt.Sprintf("%s/%s/%s", cachePath.Path, runtime.Namespace, runtime.Name))
			quotas = append(quotas, utils.TransformQuantityToEFCUnit(cachePath.Quota))
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
