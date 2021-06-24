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
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// Setup the ddc engine
func (b *TemplateEngine) Setup(ctx cruntime.ReconcileRequestContext) (createready bool, updateready bool, err error) {
	var (
		shouldSetupMaster  bool
		masterReady        bool
		shouldSetupWorkers bool
		workersReady       bool
	)

	b.Log.Info("Setup the ddc engine", "runtime", ctx.Runtime)
	// check if the dataset is ready
	// if not ready, check if we should Setup the master, worker, ufs
	// if ready, check if we should update ufs
	if !utils.IsSetupDone(ctx.Dataset) {
		// 1.Check if we should Setup1 the master
		// shouldSetupMaster, err
		shouldSetupMaster, err = b.Implement.ShouldSetupMaster()
		if err != nil {
			return createready, updateready, err
		}
		if shouldSetupMaster {
			err = b.Implement.SetupMaster()
			if err != nil {
				b.Log.Error(err, "SetupMaster")
				return createready, updateready, err
			}
		}

		// 2.Check if the master is ready, then go forward to workers Setup1
		masterReady, err = b.Implement.CheckMasterReady()
		if err != nil {
			b.Log.Error(err, "Failed to check if it CheckMasterReady.")
			return createready, updateready, err
		}

		if !masterReady {
			return createready, updateready, err
		}

		shouldCheckUFS, err := b.Implement.ShouldCheckUFS()
		if err != nil {
			b.Log.Error(err, "Failed to check if it requires checking ufs.")
			return createready, updateready, err
		}

		if shouldCheckUFS {
			err = b.Implement.PrepareUFS()
			if err != nil {
				b.Log.Error(err, "Failed to prepare ufs.")
				return createready, updateready, err
			}
		}

		// 3.Check if we should Setup1 the workers
		shouldSetupWorkers, err = b.Implement.ShouldSetupWorkers()
		if err != nil {
			b.Log.Error(err, "Failed to check if it ShouldSetupWorkers.")
			return createready, updateready, err
		}

		if shouldSetupWorkers {
			err = b.Implement.SetupWorkers()
			if err != nil {
				b.Log.Error(err, "SetupWorker")
				return createready, updateready, err
			}
		}

		// 4.Check if the workers are ready
		workersReady, err = b.Implement.CheckWorkersReady()
		if err != nil {
			b.Log.Error(err, "Check if the workers are ready")
			return createready, updateready, err
		}

		if !workersReady {
			return createready, updateready, err
		}

		// 5.Check if the runtime is ready
		runtimeReady, err := b.Implement.CheckAndUpdateRuntimeStatus()
		if err != nil {
			b.Log.Error(err, "Check if the runtime is ready")
			return createready, updateready, err
		}

		if !runtimeReady {
			return createready, updateready, err
		}

		// 6.Update the dataset status from pending to bound
		err = b.Implement.BindToDataset()
		if err != nil {
			b.Log.Error(err, "Bind the dataset")
			return createready, updateready, err
		}
		createready = true
	} else {
		createready = true
		should, added, removed, err := b.Implement.ShouldUpdateUFS()
		if err != nil {
			//r.Recorder.Eventf(ctx.Runtime, corev1.EventTypeWarning, common.ErrorProcessRuntimeReason, "Failed to check if we need to update ufs due to error %v", err)
			b.Log.Error(err, "Failed to check if we need to update the ufs")
		}
		if should {
			b.Implement.UpdateUFS(added, removed)
			updateready = true
			err = b.Implement.UFSUpdated()
		}
	}

	return createready, updateready, err
}
