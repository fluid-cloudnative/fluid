/*
  Copyright 2023 The Fluid Authors.

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

package efc

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/efc/operations"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
)

// CheckRuntimeReady checks if runtime is ready or not
func (e *EFCEngine) CheckRuntimeReady() (ready bool) {
	// 1. check master ready
	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewEFCFileUtils(podName, containerName, e.namespace, e.Log)
	ready = fileUtils.Ready()
	if !ready {
		e.Log.Info("runtime not ready", "runtime", ready)
		return false
	}

	// 2. check worker ready
	workerPods, err := e.getWorkerRunningPods()
	if err != nil {
		e.Log.Error(err, "Fail to get worker pods")
		return false
	}

	readyCount := 0
	for _, pod := range workerPods {
		if podutil.IsPodReady(&pod) {
			readyCount++
		}
	}
	return readyCount > 0
}
