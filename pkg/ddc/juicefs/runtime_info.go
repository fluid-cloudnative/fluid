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

package juicefs

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
)

// getRuntimeInfo gets runtime info
func (j *JuiceFSEngine) getRuntimeInfo() (base.RuntimeInfoInterface, error) {
	if j.runtimeInfo == nil {
		runtime, err := j.getRuntime()
		if err != nil {
			return j.runtimeInfo, err
		}

		j.runtimeInfo, err = base.BuildRuntimeInfo(j.name, j.namespace, j.runtimeType, runtime.Spec.TieredStore, base.WithMetadataList(base.GetMetadataListFromAnnotation(runtime)))
		if err != nil {
			return j.runtimeInfo, err
		}

		// Setup Fuse Deploy Mode
		j.runtimeInfo.SetupFuseDeployMode(runtime.Spec.Fuse.Global, runtime.Spec.Fuse.NodeSelector)

		if !j.UnitTest {
			// Check if the runtime is using deprecated labels
			isLabelDeprecated, err := j.HasDeprecatedCommonLabelName()
			if err != nil {
				return j.runtimeInfo, err
			}
			j.runtimeInfo.SetDeprecatedNodeLabel(isLabelDeprecated)

			// Check if the runtime is using deprecated naming style for PersistentVolumes
			isPVNameDeprecated, err := volume.HasDeprecatedPersistentVolumeName(j.Client, j.runtimeInfo, j.Log)
			if err != nil {
				return j.runtimeInfo, err
			}
			j.runtimeInfo.SetDeprecatedPVName(isPVNameDeprecated)

			j.Log.Info("Deprecation check finished", "isLabelDeprecated", j.runtimeInfo.IsDeprecatedNodeLabel(), "isPVNameDeprecated", j.runtimeInfo.IsDeprecatedPVName())

			// Setup with Dataset Info
			dataset, err := utils.GetDataset(j.Client, j.name, j.namespace)
			if err != nil {
				if utils.IgnoreNotFound(err) == nil {
					j.Log.Info("Dataset is notfound", "name", j.name, "namespace", j.namespace)
					return j.runtimeInfo, nil
				}

				j.Log.Info("Failed to get dataset when getruntimeInfo")
				return j.runtimeInfo, err
			}

			j.runtimeInfo.SetupWithDataset(dataset)

			j.Log.Info("Setup with dataset done", "exclusive", j.runtimeInfo.IsExclusive())
		}
	}

	return j.runtimeInfo, nil
}
