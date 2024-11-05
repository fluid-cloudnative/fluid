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

package referencedataset

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
)

func (e *ReferenceDatasetEngine) Sync(ctx cruntime.ReconcileRequestContext) (err error) {
	// Avoid the retires too frequently
	if !e.permitSync(types.NamespacedName{Name: ctx.Name, Namespace: ctx.Namespace}) {
		return
	}
	defer utils.TimeTrack(time.Now(), "base.Sync", "ctx", ctx)
	defer e.setTimeOfLastSync()

	virtualDataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return err
	}
	physicalRuntimeInfo, err := e.getPhysicalRuntimeInfo()
	if err != nil {
		return err
	}
	physicalDataset, err := utils.GetDataset(e.Client, physicalRuntimeInfo.GetName(), physicalRuntimeInfo.GetNamespace())
	if err != nil {
		return err
	}

	// 1. update dataset status

	// synchronize status field from physical dataset except DatasetRef and Runtimes field
	oldRuntimes := virtualDataset.Status.Runtimes
	virtualDatasetToUpdate := virtualDataset.DeepCopy()
	virtualDatasetToUpdate.Status = *physicalDataset.Status.DeepCopy()
	virtualDatasetToUpdate.Status.DatasetRef = nil
	virtualDatasetToUpdate.Status.Runtimes = oldRuntimes

	// set the Runtimes field
	if len(virtualDatasetToUpdate.Status.Runtimes) == 0 {
		virtualDatasetToUpdate.Status.Runtimes = []datav1alpha1.Runtime{}
	}
	newStatusRuntime := utils.AddRuntimesIfNotExist(virtualDatasetToUpdate.Status.Runtimes, utils.NewRuntime(e.name,
		e.namespace,
		common.AccelerateCategory,
		common.ThinRuntime,
		// TODO: should use physical dataset's runtime Spec.Master.Replicas？
		0))
	if len(newStatusRuntime) != len(virtualDatasetToUpdate.Status.Runtimes) {
		virtualDatasetToUpdate.Status.Runtimes = newStatusRuntime
		e.Log.Info("the dataset status", "runtime", virtualDatasetToUpdate.Status.Runtimes)
	}

	if !reflect.DeepEqual(virtualDataset.Status, virtualDatasetToUpdate.Status) {
		err = e.Client.Status().Update(context.TODO(), virtualDatasetToUpdate)
		if err != nil {
			return
		}
	} else {
		e.Log.Info("Do nothing because the dataset status is not changed.")
	}

	// 2. synchronize runtime status
	runtime, err := e.getRuntime()
	if err != nil {
		return err
	}
	runtimeToUpdate := runtime.DeepCopy()

	physicalRuntimeStatus, err := e.getPhysicalDatasetRuntimeStatus()
	if err != nil {
		return err
	}
	if physicalRuntimeStatus != nil {
		// status copy, include cacheStates, conditions, selector, valueFile, current*, desired*, fuse*, master*, worker* ...
		// TODO: Are there some fields should not copy?
		runtimeToUpdate.Status = *physicalRuntimeStatus.DeepCopy()
	}
	// update status.mounts to dataset mounts
	runtimeToUpdate.Status.Mounts = virtualDatasetToUpdate.Spec.Mounts

	if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
		err = e.Client.Status().Update(context.TODO(), runtimeToUpdate)
	} else {
		e.Log.Info("Do nothing because the runtime status is not changed.")
	}

	return
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

func (e *ReferenceDatasetEngine) permitSync(key types.NamespacedName) (permit bool) {
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

func (e *ReferenceDatasetEngine) setTimeOfLastSync() {
	e.timeOfLastSync = time.Now()
	e.Log.V(1).Info("Set timeOfLastSync", "timeOfLastSync", e.timeOfLastSync)
}
