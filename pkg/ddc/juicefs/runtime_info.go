/*
Copyright 2021 The Fluid Authors.

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

		opts := []base.RuntimeInfoOption{
			base.WithTieredStore(runtime.Spec.TieredStore),
			base.WithMetadataList(base.GetMetadataListFromAnnotation(runtime)),
		}

		j.runtimeInfo, err = base.BuildRuntimeInfo(j.name, j.namespace, j.runtimeType, opts...)
		if err != nil {
			return j.runtimeInfo, err
		}

		// Setup Fuse Deploy Mode
		j.runtimeInfo.SetFuseNodeSelector(runtime.Spec.Fuse.NodeSelector)

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
