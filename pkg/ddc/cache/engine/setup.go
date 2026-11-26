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
	"context"
	stderrors "errors"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/metrics"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (e *CacheEngine) Setup(ctx cruntime.ReconcileRequestContext) (ready bool, err error) {
	defer func() {
		if err != nil {
			metrics.GetOrCreateRuntimeMetrics(ctx.Runtime.GetObjectKind().GroupVersionKind().Kind, ctx.Namespace, ctx.Name).SetupErrorInc()
		}
	}()

	dataset := ctx.Dataset
	runtime := ctx.Runtime.(*datav1alpha1.CacheRuntime)
	// Bind the runtime to a snapshot of its CacheRuntimeClass before anything is created from
	// it: the components and configmaps are rendered from the snapshot, and so is every later
	// sync, so editing the class does not reach this runtime.
	runtimeClass, err := e.getBoundRuntimeClass(ctx, runtime)
	if err != nil {
		return false, errors.Wrapf(err, "failed to get CacheRuntimeClass %s", runtime.Spec.RuntimeClassName)
	}

	runtimeValue, transformErr := e.transform(dataset, runtime, runtimeClass)
	var conflict *componentResourcesConflictError
	if transformErr != nil && !stderrors.As(transformErr, &conflict) {
		return false, transformErr
	}

	current, statusErr := e.updateSetupResourceCondition(ctx.Context, runtime, conflict)
	if apierrors.IsConflict(statusErr) {
		if conflict == nil {
			return false, nil
		}
		// The informer may still be behind after the last rejected write. Read directly
		// from the API server before reporting the creation error on the Dataset.
		latest := &datav1alpha1.CacheRuntime{}
		if err := e.SubResource("status").Get(ctx.Context, runtime, latest); err != nil {
			return false, client.IgnoreNotFound(err)
		}
		current = isCurrentSetupGeneration(runtime, latest)
	} else if statusErr != nil {
		return false, stderrors.Join(transformErr, statusErr)
	}
	if !current {
		return false, nil
	}
	if transformErr != nil {
		return false, transformErr
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

// updateSetupResourceCondition records or clears a creation conflict for the generation
// that Setup transformed. A false result tells the caller to reconcile a fresh runtime.
func (e *CacheEngine) updateSetupResourceCondition(ctx context.Context, observed *datav1alpha1.CacheRuntime,
	conflict *componentResourcesConflictError) (bool, error) {
	var cond datav1alpha1.RuntimeCondition
	if conflict != nil {
		cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeResourcesSynced, datav1alpha1.RuntimeResourceCreationConflictReason,
			conflict.message, corev1.ConditionFalse)
	} else {
		if !hasSetupResourceConflict(observed) {
			return true, nil
		}
		cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeResourcesSynced, datav1alpha1.RuntimeWaitingForSyncReason,
			"Resource configuration is valid. Waiting for resource sync.", corev1.ConditionUnknown)
	}
	var current bool
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		latest := &datav1alpha1.CacheRuntime{}
		if err := e.Get(ctx, client.ObjectKeyFromObject(observed), latest); err != nil {
			return err
		}
		current = isCurrentSetupGeneration(observed, latest)
		if !current || (conflict == nil && !hasSetupResourceConflict(latest)) {
			return nil
		}
		if !applyRuntimeCondition(&latest.Status, cond) {
			return nil
		}
		return e.Client.Status().Update(ctx, latest)
	})
	return current, err
}

func isCurrentSetupGeneration(observed, latest *datav1alpha1.CacheRuntime) bool {
	return latest.UID == observed.UID && latest.Generation == observed.Generation && latest.DeletionTimestamp.IsZero()
}

func hasSetupResourceConflict(runtime *datav1alpha1.CacheRuntime) bool {
	_, cond := utils.GetRuntimeCondition(runtime.Status.Conditions, datav1alpha1.RuntimeResourcesSynced)
	if cond == nil || cond.Status != corev1.ConditionFalse {
		return false
	}
	if cond.Reason == datav1alpha1.RuntimeResourceCreationConflictReason {
		return true
	}
	if cond.Reason != common.RuntimeResourcesConflict {
		return false
	}
	// Older controllers used the same reason for Setup and Sync. Match only the
	// historical creation message when recovering one of those persisted conditions.
	for _, component := range []common.ComponentType{common.ComponentTypeMaster, common.ComponentTypeWorker, common.ComponentTypeClient} {
		if strings.HasPrefix(cond.Message, string(component)+" cannot be created: the CacheRuntime and its CacheRuntimeClass template merge into ") {
			return true
		}
	}
	return false
}
