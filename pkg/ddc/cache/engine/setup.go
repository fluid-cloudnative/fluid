/*
  Copyright 2026 The Fluid Authors.

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

package engine

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/metrics"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (e *CacheEngine) Setup(ctx cruntime.ReconcileRequestContext) (ready bool, err error) {
	defer func() {
		if err != nil {
			metrics.GetOrCreateRuntimeMetrics(ctx.Runtime.GetObjectKind().GroupVersionKind().Kind, ctx.Namespace, ctx.Name).SetupErrorInc()
		}
	}()

	dataset := ctx.Dataset
	runtime := ctx.Runtime.(*datav1alpha1.CacheRuntime)
	runtimeClass, err := e.getRuntimeClass(runtime.Spec.RuntimeClassName)
	if err != nil {
		return false, errors.Wrapf(err, "failed to get CacheRuntimeClass %s", runtime.Spec.RuntimeClassName)
	}

	runtimeValue, err := e.transform(dataset, runtime, runtimeClass)
	if err != nil {
		return false, err
	}

	// create runtime value configmap for runtime mount
	err = e.createRuntimeConfigMaps(runtimeClass)
	if err != nil {
		return false, err
	}

	// Create Master/Worker/Client components
	e.Log.Info("Setup runtime", "runtime", ctx.Runtime)
	if runtimeValue.Master.Enabled {
		e.Log.Info("Setup master", "runtime", ctx.Runtime)
		ready, err = e.SetupMasterComponent(runtimeValue.Master)
		if !ready || err != nil {
			return
		}
	}

	if runtimeValue.Worker.Enabled {
		e.Log.Info("Setup worker", "runtime", ctx.Runtime)
		ready, err = e.SetupWorkerComponent(runtimeValue.Worker)
		if !ready || err != nil {
			return
		}
	}

	if runtimeValue.Client.Enabled {
		e.Log.Info("Setup client", "runtime", ctx.Runtime)
		ready, err = e.SetupClientComponent(runtimeValue.Client)
		if !ready || err != nil {
			return
		}
	}

	// dataset mount
	if runtimeValue.Master.Enabled {
		// Wait for the master pod to be running before executing the mount command.
		// The pod may still be starting up after StatefulSet creation.
		masterPodRunning, err := e.isMasterPodRunning(runtimeValue)
		if err != nil {
			return false, err
		}
		if !masterPodRunning {
			e.Log.Info("Master pod is not yet running, will retry in next reconcile")
			return false, nil
		}

		err = e.PrepareUFS(runtimeClass.Topology.Master.ExecutionEntries, runtimeValue)
		if err != nil {
			return false, err
		}
	}

	ready, err = e.CheckAndUpdateRuntimeStatus(runtimeValue)
	if err != nil {
		_ = utils.LoggingErrorExceptConflict(e.Log, err, "Failed to check if the runtime is ready", types.NamespacedName{Namespace: e.namespace, Name: e.name})
		return
	}
	if !ready {
		return
	}

	if err = e.BindToDataset(); err != nil {
		_ = utils.LoggingErrorExceptConflict(e.Log, err, "Failed to bind the dataset", types.NamespacedName{Namespace: e.namespace, Name: e.name})
		return false, err
	}

	return true, nil
}
