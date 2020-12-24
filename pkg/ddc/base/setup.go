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

package base

import (
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

// Setup the ddc engine
func (b *TemplateEngine) Setup(ctx cruntime.ReconcileRequestContext) (ready bool, err error) {
	var (
		shouldSetupMaster  bool
		masterReady        bool
		shouldSetupWorkers bool
		workersReady       bool
	)

	b.Log.Info("Setup the ddc engine", "runtime", ctx.Runtime)
	// 1.Check if we should setup the master
	// shouldSetupMaster, err
	shouldSetupMaster, err = b.Implement.ShouldSetupMaster()
	if err != nil {
		return ready, err
	}
	if shouldSetupMaster {
		err = b.Implement.SetupMaster()
		if err != nil {
			b.Log.Error(err, "SetupMaster")
			return ready, err
		}
	}

	// 2.Check if the master is ready, then go forward to workers setup
	masterReady, err = b.Implement.CheckMasterReady()
	if err != nil {
		b.Log.Error(err, "Failed to check if it CheckMasterReady.")
		return ready, err
	}

	if !masterReady {
		return masterReady, err
	}

	shouldCheckUFS, err := b.Implement.ShouldCheckUFS()
	if err != nil {
		b.Log.Error(err, "Failed to check if it requires checking ufs.")
		return ready, err
	}

	if shouldCheckUFS {
		err = b.Implement.PrepareUFS()
		if err != nil {
			b.Log.Error(err, "Failed to prepare ufs.")
			return ready, err
		}
	}

	// 3.Check if we should setup the workers
	shouldSetupWorkers, err = b.Implement.ShouldSetupWorkers()
	if err != nil {
		b.Log.Error(err, "Failed to check if it ShouldSetupWorkers.")
		return ready, err
	}

	if shouldSetupWorkers {
		err = b.Implement.SetupWorkers()
		if err != nil {
			b.Log.Error(err, "SetupWorker")
			return ready, err
		}
	}

	// 4.Check if the workers are ready
	workersReady, err = b.Implement.CheckWorkersReady()
	if err != nil {
		b.Log.Error(err, "Check if the workers are ready")
		return workersReady, err
	}

	if !workersReady {
		return workersReady, err
	}

	// 5.Check if the runtime is ready
	runtimeReady, err := b.Implement.CheckAndUpdateRuntimeStatus()
	if err != nil {
		b.Log.Error(err, "Check if the runtime is ready")
		return runtimeReady, err
	}

	if !runtimeReady {
		return runtimeReady, err
	}

	// 6.Update the dataset status from pending to bound
	err = b.Implement.BindToDataset()
	if err != nil {
		b.Log.Error(err, "Bind the dataset")
		return workersReady, err
	}

	ready = true

	return ready, err
}
