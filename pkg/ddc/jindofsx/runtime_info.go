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
	"github.com/fluid-cloudnative/fluid/pkg/utils/testutil"
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

	if testutil.IsUnitTest() {
		return e.runtimeInfo, nil
	}

	// Handling information of bound dataset. XXXEngine.getRuntimeInfo() might be called before the runtime is bound to a dataset,
	// so here we must lazily set dataset-related information once we found there's one bound dataset.
	if len(e.runtimeInfo.GetOwnerDatasetUID()) == 0 {
		runtime, err := e.getRuntime()
		if err != nil {
			return nil, err
		}

		uid, err := base.GetOwnerDatasetUIDFromRuntimeMeta(runtime.ObjectMeta)
		if err != nil {
			return nil, err
		}

		if len(uid) > 0 {
			e.runtimeInfo.SetOwnerDatasetUID(uid)
		}
	}

	exclusiveModePtr := e.runtimeInfo.IsExclusive()
	if exclusiveModePtr == nil {
		dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
		if utils.IgnoreNotFound(err) != nil {
			return nil, err
		}

		if dataset != nil {
			e.runtimeInfo.SetupWithDataset(dataset)
		}
	}

	return e.runtimeInfo, nil
}
