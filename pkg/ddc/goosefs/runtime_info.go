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

package goosefs

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
)

// getRuntimeInfo gets runtime info
func (e *GooseFSEngine) getRuntimeInfo() (base.RuntimeInfoInterface, error) {
	if e.runtimeInfo == nil {
		runtime, err := e.getRuntime()
		if err != nil {
			return e.runtimeInfo, err
		}

		e.runtimeInfo, err = base.BuildRuntimeInfo(e.name, e.namespace, e.runtimeType, runtime.Spec.TieredStore, base.WithMetadataList(base.GetMetadataListFromAnnotation(runtime)))
		if err != nil {
			return e.runtimeInfo, err
		}

		// Setup Fuse Deploy Mode
		e.runtimeInfo.SetupFuseDeployMode(runtime.Spec.Fuse.Global, runtime.Spec.Fuse.NodeSelector)

		if !e.UnitTest {
			// Check if the runtime is using deprecated labels
			isLabelDeprecated, err := e.HasDeprecatedCommonLabelname()
			if err != nil {
				return e.runtimeInfo, err
			}
			e.runtimeInfo.SetDeprecatedNodeLabel(isLabelDeprecated)

			// Check if the runtime is using deprecated naming style for PersistentVolumes
			isPVNameDeprecated, err := volume.HasDeprecatedPersistentVolumeName(e.Client, e.runtimeInfo, e.Log)
			if err != nil {
				return e.runtimeInfo, err
			}
			e.runtimeInfo.SetDeprecatedPVName(isPVNameDeprecated)

			e.Log.Info("Deprecation check finished", "isLabelDeprecated", e.runtimeInfo.IsDeprecatedNodeLabel(), "isPVNameDeprecated", e.runtimeInfo.IsDeprecatedPVName())

			// Setup with Dataset Info
			dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
			if err != nil {
				if utils.IgnoreNotFound(err) == nil {
					e.Log.Info("Dataset is notfound", "name", e.name, "namespace", e.namespace)
					return e.runtimeInfo, nil
				}

				e.Log.Info("Failed to get dataset when getruntimeInfo")
				return e.runtimeInfo, err
			}

			e.runtimeInfo.SetupWithDataset(dataset)

			e.Log.Info("Setup with dataset done", "exclusive", e.runtimeInfo.IsExclusive())
		}
	}

	return e.runtimeInfo, nil
}
