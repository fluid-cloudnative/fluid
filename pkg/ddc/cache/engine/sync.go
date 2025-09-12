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

package engine

import (
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/pkg/errors"
	"os"
	"time"

	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func (e *CacheEngine) Sync(ctx cruntime.ReconcileRequestContext) (err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return err
	}

	runtime, err := e.getRuntime()
	if err != nil {
		return errors.Wrapf(err, "failed to get CacheRuntime %s", runtime.Name)
	}

	runtimeClass, err := utils.GetCacheRuntimeClass(ctx.Client, runtime.Spec.RuntimeClassName)
	if err != nil {
		return errors.Wrapf(err, "failed to get CacheRuntimeClass %s", runtime.Spec.RuntimeClassName)
	}

	runtimeValue, err := e.transform(dataset, runtime, runtimeClass)
	if err != nil {
		return err
	}

	return e.syncCacheRuntimeConfig(dataset, runtimeValue)
	//// permitSyncEngineStatus avoids frequent rpcs with engines with rate-limited retries
	//permitSyncEngineStatus := e.permitSync(types.NamespacedName{Name: ctx.Name, Namespace: ctx.Namespace})
	//if permitSyncEngineStatus {
	//	defer e.setTimeOfLastSync()
	//}
	//
	//defer utils.TimeTrack(time.Now(), "cache.Sync", "ctx", ctx)
	//
	//// 1. Sync replicas
	//if err = e.SyncReplicas(ctx); err != nil {
	//	return
	//}
	//
	//// 2. Sync Runtime Spec
	//updated, err := e.SyncRuntime(ctx)
	//if err != nil || updated {
	//	return
	//}
	//
	//if permitSyncEngineStatus {
	//	// 3. Update runtime status
	//	if _, err = e.CheckAndUpdateRuntimeStatus(nil); err != nil {
	//		return
	//	}
	//	// 4. Update runtime config
	//	if err = e.SyncRuntimeConfigFile(ctx); err != nil {
	//		return
	//	}
	//}
	//
	//return e.SyncScheduleInfoToCacheNodes(ctx)
}

func getSyncRetryDuration() (d *time.Duration, err error) {
	if value, existed := os.LookupEnv(syncRetryDurationEnv); existed {
		duration, err := time.ParseDuration(value)
		if err != nil {
			return d, err
		}
		d = &duration
	}
	return
}

func (e *CacheEngine) permitSync(key types.NamespacedName) (permit bool) {
	if time.Since(e.timeOfLastSync) < e.syncRetryDuration {
		info := fmt.Sprintf("Skipping engine.Sync(). Not permmitted until  %v (syncRetryDuration %v) since timeOfLastSync %v.",
			e.timeOfLastSync.Add(e.syncRetryDuration),
			e.syncRetryDuration,
			e.timeOfLastSync)
		e.Log.Info(info, "name", key.Name, "namespace", key.Namespace)
	} else {
		permit = true
		info := fmt.Sprintf("Processing engine.Sync(). permmitted  %v (syncRetryDuration %v) since timeOfLastSync %v.",
			e.timeOfLastSync.Add(e.syncRetryDuration),
			e.syncRetryDuration,
			e.timeOfLastSync)
		e.Log.V(1).Info(info, "name", key.Name, "namespace", key.Namespace)
	}

	return
}

func (e *CacheEngine) setTimeOfLastSync() {
	e.timeOfLastSync = time.Now()
	e.Log.V(1).Info("Set timeOfLastSync", "timeOfLastSync", e.timeOfLastSync)
}

func (e *CacheEngine) SyncReplicas(ctx context.Context) error {
	return nil
}

func (e *CacheEngine) SyncRuntime(ctx context.Context) (bool, error) {
	return false, nil
}

func (e *CacheEngine) SyncRuntimeConfigFile(ctx context.Context) error {
	return nil
}

func (e *CacheEngine) SyncScheduleInfoToCacheNodes(ctx context.Context) error {
	return nil
}
