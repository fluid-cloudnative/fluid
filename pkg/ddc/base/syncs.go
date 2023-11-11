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
