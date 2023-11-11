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

package thin

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
)

// getRuntimeInfo gets runtime info
func (t *ThinEngine) getRuntimeInfo() (base.RuntimeInfoInterface, error) {
	if t.runtimeInfo == nil {
		runtime, err := t.getRuntime()
		if err != nil {
			return t.runtimeInfo, err
		}

		t.runtimeInfo, err = base.BuildRuntimeInfo(t.name, t.namespace, t.runtimeType, runtime.Spec.TieredStore, base.WithMetadataList(base.GetMetadataListFromAnnotation(runtime)))
		if err != nil {
			return t.runtimeInfo, err
		}

		// Setup Fuse Deploy Mode
		t.runtimeInfo.SetupFuseDeployMode(true, runtime.Spec.Fuse.NodeSelector)

		if !t.UnitTest {
			// Check if the runtime is using deprecated naming style for PersistentVolumes
			isPVNameDeprecated, err := volume.HasDeprecatedPersistentVolumeName(t.Client, t.runtimeInfo, t.Log)
			if err != nil {
				return t.runtimeInfo, err
			}
			t.runtimeInfo.SetDeprecatedPVName(isPVNameDeprecated)

			t.Log.Info("Deprecation check finished", "isLabelDeprecated", t.runtimeInfo.IsDeprecatedNodeLabel(), "isPVNameDeprecated", t.runtimeInfo.IsDeprecatedPVName())

			// Setup with Dataset Info
			dataset, err := utils.GetDataset(t.Client, t.name, t.namespace)
			if err != nil {
				if utils.IgnoreNotFound(err) == nil {
					t.Log.Info("Dataset is notfound", "name", t.name, "namespace", t.namespace)
					return t.runtimeInfo, nil
				}

				t.Log.Info("Failed to get dataset when getruntimeInfo")
				return t.runtimeInfo, err
			}

			t.runtimeInfo.SetupWithDataset(dataset)

			t.Log.Info("Setup with dataset done", "exclusive", t.runtimeInfo.IsExclusive())
		}
	}

	return t.runtimeInfo, nil
}
