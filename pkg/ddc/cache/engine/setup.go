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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/metrics"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
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
	err = e.createRuntimeConfigMaps(ctx, runtimeClass)
	if err != nil {
		return false, err
	}

	// Create Master/Worker/Client components, won't be nil.
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

	// CheckAndUpdateRuntimeStatus after components are setup
	// Use lightweight getRuntimeStatusValue instead of full transform for status update
	statusValue, err := e.getRuntimeStatusValue(runtime, runtimeClass)
	if err != nil {
		return false, err
	}
	ready, err = e.CheckAndUpdateRuntimeStatus(statusValue)
	if err != nil {
		_ = utils.LoggingErrorExceptConflict(e.Log, err, "Failed to check if the runtime is ready", types.NamespacedName{Namespace: e.namespace, Name: e.name})
		return
	}
	if !ready {
		return
	}

	// Mount ufs if supported
	archApi := resolveArchitectureApi(e.name, e.namespace, runtime, runtimeClass)
	if archApi.IsMountUFSSupported() {
		// ignore the output for mount command, if executing succeed, all ufs mount will be ready.
		// Even if dataset changes the mount info concurrently, the `sync` phase will make it correct eventually.
		_, err = e.PrepareUFS(archApi)
		if err != nil {
			e.Recorder.Eventf(runtime, corev1.EventTypeWarning, common.RuntimeMountUfsFailed, "Failed to execute mount ufs command")
			return false, err
		}
	}
	if err = e.BindToDataset(runtime, runtimeClass); err != nil {
		_ = utils.LoggingErrorExceptConflict(e.Log, err, "Failed to bind the dataset", types.NamespacedName{Namespace: e.namespace, Name: e.name})
		return false, err
	}

	return true, nil
}
