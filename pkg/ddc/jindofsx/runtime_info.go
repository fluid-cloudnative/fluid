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

package jindofsx

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
)

// getRuntimeInfo gets runtime info
func (e *JindoFSxEngine) getRuntimeInfo() (base.RuntimeInfoInterface, error) {
	if e.runtimeInfo == nil {
		runtime, err := e.getRuntime()
		if err != nil {
			return e.runtimeInfo, err
		}
		opts := []base.RuntimeInfoOption{
			base.WithTieredStore(runtime.Spec.TieredStore),
			base.WithMetadataList(base.GetMetadataListFromAnnotation(runtime)),
			base.WithAnnotations(runtime.Annotations),
		}
		// TODO: For now hack runtimeType with engineImpl for backward compatibility. Fix this
		// when refactoring runtimeInfo.
		e.runtimeInfo, err = base.BuildRuntimeInfo(e.name, e.namespace, e.engineImpl, opts...)
		if err != nil {
			return e.runtimeInfo, err
		}

		// Setup Fuse Deploy Mode
		e.runtimeInfo.SetFuseNodeSelector(runtime.Spec.Fuse.NodeSelector)

		// Setup with Dataset Info
		dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
		if err != nil {
			if len(runtime.GetOwnerReferences()) > 0 {
				e.runtimeInfo.SetOwnerDatasetUID(runtime.GetOwnerReferences()[0].UID)
			}
			if utils.IgnoreNotFound(err) == nil {
				e.Log.Info("Dataset is notfound", "name", e.name, "namespace", e.namespace)
				return e.runtimeInfo, nil
			}

			e.Log.Info("Failed to get dataset when getruntimeInfo")
			return e.runtimeInfo, err
		}

		e.runtimeInfo.SetupWithDataset(dataset)
		e.Log.Info("Setup with dataset done", "exclusive", e.runtimeInfo.IsExclusive())

		e.runtimeInfo.SetOwnerDatasetUID(dataset.GetUID())

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
	}

	return e.runtimeInfo, nil
}
