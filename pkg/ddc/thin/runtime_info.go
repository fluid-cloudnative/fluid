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

package thin

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/testutil"
)

// CheckRuntimeReady checks if the ThinRuntime is ready to serve data operations.
// Unlike Alluxio/Jindo which probe a master pod, ThinRuntime has no master component,
// so readiness is determined by worker availability. Fuse components are intentionally
// excluded because fluid treats fuse as always-ready by design (see pkg/ctrl/fuse.go).
func (t *ThinEngine) CheckRuntimeReady() (ready bool) {
	workerReady, err := t.CheckWorkersReady()
	if err != nil {
		return false
	}
	return workerReady
}

// getRuntimeInfo gets runtime info
func (t *ThinEngine) getRuntimeInfo() (base.RuntimeInfoInterface, error) {
	if t.runtimeInfo == nil {
		runtime, err := t.getRuntime()
		if err != nil {
			return t.runtimeInfo, err
		}

		opts := []base.RuntimeInfoOption{
			base.WithTieredStore(runtime.Spec.TieredStore),
			base.WithMetadataList(base.GetMetadataListFromAnnotation(runtime)),
			base.WithAnnotations(runtime.Annotations),
		}

		t.runtimeInfo, err = base.BuildRuntimeInfo(t.name, t.namespace, t.runtimeType, opts...)
		if err != nil {
			return t.runtimeInfo, err
		}

		// Setup Fuse Deploy Mode
		t.runtimeInfo.SetFuseNodeSelector(runtime.Spec.Fuse.NodeSelector)
	}

	if testutil.IsUnitTest() {
		return t.runtimeInfo, nil
	}

	// Handling information of bound dataset. XXXEngine.getRuntimeInfo() might be called before the runtime is bound to a dataset,
	// so here we must lazily set dataset-related information once we found there's one bound dataset.
	if len(t.runtimeInfo.GetOwnerDatasetUID()) == 0 {
		runtime, err := t.getRuntime()
		if err != nil {
			return nil, err
		}

		uid, err := base.GetOwnerDatasetUIDFromRuntimeMeta(runtime.ObjectMeta)
		if err != nil {
			return nil, err
		}

		if len(uid) > 0 {
			t.runtimeInfo.SetOwnerDatasetUID(uid)
		}
	}

	if !t.runtimeInfo.IsPlacementModeSet() {
		dataset, err := utils.GetDataset(t.Client, t.name, t.namespace)
		if utils.IgnoreNotFound(err) != nil {
			return nil, err
		}

		if dataset != nil {
			t.runtimeInfo.SetupWithDataset(dataset)
		}
	}

	return t.runtimeInfo, nil
}
