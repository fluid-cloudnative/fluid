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

package base

import (
	"github.com/fluid-cloudnative/fluid/pkg/metrics"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

// Setup the ddc engine
func (b *TemplateEngine) Setup(ctx cruntime.ReconcileRequestContext) (ready bool, err error) {
	defer func() {
		if err != nil {
			metrics.GetRuntimeMetrics(ctx.Runtime.GetObjectKind().GroupVersionKind().Kind, ctx.Namespace, ctx.Name).SetupErrorInc()
		}
	}()

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
			// b.Log.Error(err, "SetupWorker")
			_ = b.loggingErrorExceptConflict(err, "Failed to setup worker")
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
		// b.Log.Error(err, "Check if the runtime is ready")
		_ = b.loggingErrorExceptConflict(err, "Failed to check if the runtime is ready")
		return runtimeReady, err
	}

	if !runtimeReady {
		return runtimeReady, err
	}

	// 6.Update the dataset status from pending to bound
	err = b.Implement.BindToDataset()
	if err != nil {
		// b.Log.Error(err, "Bind the dataset")
		_ = b.loggingErrorExceptConflict(err, "Failed to bind the dataset")
		return workersReady, err
	}

	ready = true

	return ready, err
}
