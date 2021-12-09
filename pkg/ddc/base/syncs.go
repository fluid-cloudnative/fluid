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
	"fmt"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// SyncReplicas syncs the replicas
func (t *TemplateEngine) Sync(ctx cruntime.ReconcileRequestContext) (err error) {
	defer utils.TimeTrack(time.Now(), "base.Sync", "ctx", ctx)
	// Avoid the retires too frequently
	if time.Since(t.timeOfLastSync) < t.syncRetryDuration {
		info := fmt.Sprintf("Skipping engine.Sync(). Not permmitted until  %v (syncRetryDuration %v).",
			t.timeOfLastSync.Add(t.syncRetryDuration),
			t.syncRetryDuration)
		log.Log.Info(info, "name", ctx.Name, "namespace", ctx.Namespace)
		return
	}

	err = t.Implement.SyncMetadata()
	if err != nil {
		return
	}

	_, err = t.Implement.CheckAndUpdateRuntimeStatus()
	if err != nil {
		return
	}

	err = t.Implement.UpdateCacheOfDataset()
	if err != nil {
		return
	}

	// 1. Check healthy
	err = t.Implement.CheckRuntimeHealthy()
	if err != nil {
		return
	}

	// 2. Sync replicas
	err = t.Implement.SyncReplicas(ctx)
	if err != nil {
		return
	}

	// 3. Update runtime
	_, err = t.Implement.CheckAndUpdateRuntimeStatus()
	if err != nil {
		return
	}

	// 4. Update dataset mount point
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

	return t.Implement.SyncScheduleInfoToCacheNodes()
}
