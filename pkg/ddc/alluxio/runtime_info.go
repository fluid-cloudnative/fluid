/*

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

package alluxio

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// getRuntimeInfo gets runtime info
func (e *AlluxioEngine) getRuntimeInfo() (base.RuntimeInfoInterface, error) {
	if e.runtimeInfo == nil {
		runtime, err := e.getRuntime()
		if err != nil {
			return e.runtimeInfo, err
		}

		e.runtimeInfo, err = base.BuildRuntimeInfo(e.name, e.namespace, e.runtimeType, runtime.Spec.Tieredstore)
		e.runtimeInfo.SetupFuseDeployMode(runtime.Spec.Fuse.Global, runtime.Spec.Fuse.NodeSelector)
		if err != nil {
			return e.runtimeInfo, err
		}
	}

	if !e.UnitTest {
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
	}

	return e.runtimeInfo, nil
}
