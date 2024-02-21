/*
Copyright 2024 The Fluid Authors.
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

package vineyard

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
)

// getRuntimeInfo gets runtime info
func (e *VineyardEngine) getRuntimeInfo() (base.RuntimeInfoInterface, error) {
	if e.runtimeInfo == nil {
		runtime, err := e.getRuntime()
		if err != nil {
			return e.runtimeInfo, err
		}
		// Add the symlink method to the vineyard runtime metadata
		if runtime.ObjectMeta.Annotations == nil {
			runtime.ObjectMeta.Annotations = make(map[string]string)
		}
		runtime.ObjectMeta.Annotations["data.fluid.io/metadataList"] = `[{"Labels": {"fluid.io/node-puhlish-method": "symlink"}, "selector": { "kind": "PersistentVolume"}}]`
		tieredStore := runtime.Spec.TieredStore
		e.runtimeInfo, err = base.BuildRuntimeInfo(e.name, e.namespace, e.runtimeType, tieredStore, base.WithMetadataList(base.GetMetadataListFromAnnotation(runtime)))
		if err != nil {
			return e.runtimeInfo, err
		}

		// Setup Fuse Deploy Mode
		e.runtimeInfo.SetupFuseDeployMode(common.VineyardFuseIsGlobal, common.VineyardFuseNodeSelector)
	}

	return e.runtimeInfo, nil
}
