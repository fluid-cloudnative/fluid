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

package base

import (
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/metrics"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
)

// SyncReplicas syncs the replicas
func (t *TemplateEngine) Sync(ctx cruntime.ReconcileRequestContext) (err error) {
	// permitSyncEngineStatus avoids frequent rpcs with engines with rate limited retries
	permitSyncEngineStatus := t.permitSync(types.NamespacedName{Name: ctx.Name, Namespace: t.Context.Namespace})
	if permitSyncEngineStatus {
		defer t.setTimeOfLastSync()
	}

	defer utils.TimeTrack(time.Now(), "base.Sync", "ctx", ctx)

	if permitSyncEngineStatus {
		err = t.Implement.SyncMetadata()
		if err != nil {
			return
		}
	}

	// 1. Sync replicas
	err = t.Implement.SyncReplicas(ctx)
	if err != nil {
		return
	}

	// 2. Sync Runtime Spec
	var updated bool
	updated, err = t.Implement.SyncRuntime(ctx)
	if err != nil {
		return
	}
	if updated {
		return
	}

	// 3. Check healthy
	err = t.Implement.CheckRuntimeHealthy()
	if err != nil {
		metrics.GetRuntimeMetrics(ctx.Runtime.GetObjectKind().GroupVersionKind().Kind, ctx.Namespace, ctx.Name).HealthCheckErrorInc()
		return
	}

	// 4. Update runtime status
	if permitSyncEngineStatus {
		_, err = t.Implement.CheckAndUpdateRuntimeStatus()
		if err != nil {
			return
		}
	}

	// 5. Update the cached of dataset
	err = t.Implement.UpdateCacheOfDataset()
	if err != nil {
		return
	}

	// 6. Update dataset mount point
	if permitSyncEngineStatus {
		ufsToUpdate := t.Implement.ShouldUpdateUFS()
		if ufsToUpdate != nil {
			if ufsToUpdate.ShouldUpdate() {
				var updateReady bool
				updateReady, err = t.Implement.UpdateOnUFSChange(ufsToUpdate)
				if err != nil {
					return
				}
				if updateReady {
					err = utils.UpdateMountStatus(t.Client, t.Context.Name, t.Context.Namespace, datav1alpha1.BoundDatasetPhase)
					if err != nil {
						return
					}
				}
			}
		}
	}

	return t.Implement.SyncScheduleInfoToCacheNodes()
}

func (t *TemplateEngine) setTimeOfLastSync() {
	t.timeOfLastSync = time.Now()
	t.Log.V(1).Info("Set timeOfLastSync", "timeOfLastSync", t.timeOfLastSync)
}
