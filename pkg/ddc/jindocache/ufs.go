/*
Copyright 2023 The Fluid Authors.

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

package jindocache

import (
	"context"
	"fmt"
	"reflect"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindocache/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

// ShouldCheckUFS checks if it requires checking UFS
func (e *JindoCacheEngine) ShouldCheckUFS() (should bool, err error) {
	should = true
	return
}

// PrepareUFS do all the UFS preparations
func (e *JindoCacheEngine) PrepareUFS() (err error) {

	if e.runtime.Spec.Master.Disabled {
		err = nil
		return
	}

	if !e.CheckRuntimeReady() {
		return fmt.Errorf("runtime engine is not ready")
	}

	// 1. Mount UFS (Synchronous Operation)
	shouldMountUfs, err := e.shouldMountUFS()
	if err != nil {
		return
	}
	e.Log.Info("ShouldMountUFS", "should", shouldMountUfs)

	if shouldMountUfs {
		err = e.mountUFS()
		if err != nil {
			return
		}
	}

	// 2. Setup cacheset
	shouldRefresh, err := e.ShouldRefreshCacheSet()
	if err != nil {
		return
	}
	e.Log.Info("ShouldRefresh", "should", shouldRefresh)

	if shouldRefresh {
		err = e.RefreshCacheSet()
		if err != nil {
			return
		}
	}

	// 3. SyncMetadata
	e.Log.Info("SyncMetadata")
	err = e.SyncMetadata()
	if err != nil {
		// just report this error and ignore it because SyncMetadata isn't on the critical path of Setup
		e.Log.Error(err, "SyncMetadata")
		return nil
	}

	return
}

// UsedStorageBytes returns used storage size of Jindo in bytes
func (e *JindoCacheEngine) UsedStorageBytes() (value int64, err error) {

	return
}

// FreeStorageBytes returns free storage size of Jindo in bytes
func (e *JindoCacheEngine) FreeStorageBytes() (value int64, err error) {
	return
}

// return total storage size of Jindo in bytes
func (e *JindoCacheEngine) TotalStorageBytes() (value int64, err error) {
	return
}

// return the total num of files in Jindo
func (e *JindoCacheEngine) TotalFileNums() (value int64, err error) {
	return
}

// report jindo summary
func (e *JindoCacheEngine) GetReportSummary() (summary string, err error) {
	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewJindoFileUtils(podName, containerName, e.namespace, e.Log)
	return fileUtils.ReportSummary()
}

// JindoCacheEngine hasn't support UpdateOnUFSChange
func (e *JindoCacheEngine) ShouldUpdateUFS() (ufsToUpdate *utils.UFSToUpdate) {
	return
}

func (e *JindoCacheEngine) UpdateOnUFSChange(*utils.UFSToUpdate) (updateReady bool, err error) {
	return
}

func (e *JindoCacheEngine) ShouldSyncDatasetMounts() (should bool, err error) {
	runtime, err := utils.GetJindoRuntime(e.Client, e.name, e.namespace)
	if err != nil {
		e.Log.Error(err, "failed to get runtime when checking ufs change")
		return false, errors.WithMessage(err, "failed to get runtime when checking if dataset mounts need to be synced")
	}

	if runtime.Spec.Master.Disabled {
		return false, nil
	}

	masterPodName, masterContainerName := e.getMasterPodInfo()
	masterPod, err := kubeclient.GetPodByName(e.Client, masterPodName, e.namespace)
	if err != nil || masterPod == nil {
		e.Log.Error(err, "failed to get master pod when checking ufs change")
		return false, errors.WithMessage(err, "failed to get master pod when checking if dataset mounts need to be synced")
	}

	var startedAt *metav1.Time
	for _, containerStatus := range masterPod.Status.ContainerStatuses {
		if containerStatus.Name == masterContainerName {
			if containerStatus.State.Running == nil {
				e.Log.Info("Jindocache master container is not running, recheck its status in next reconcilation loop")
				return false, nil
			} else {
				startedAt = &containerStatus.State.Running.StartedAt
				break
			}
		}
	}

	// either runtime.Status.MountTime is not set (for backward compatibility) or startedAt is earlier than runtime.Status.MountTime (i.e. jindocache master is restarted), we need to reprepare UFS
	needReprepareUFS := runtime.Status.MountTime == nil || (startedAt != nil && runtime.Status.MountTime.Before(startedAt))

	return needReprepareUFS, nil
}

func (e *JindoCacheEngine) SyncDatasetMounts() (err error) {
	// remount Dataset.spec.mounts and refresh cachesets
	if err = e.PrepareUFS(); err != nil {
		return err
	}

	// update runtime.Status.MountTime to indicate that the Dataset.spec.mounts has been synced
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := utils.GetJindoRuntime(e.Client, e.name, e.namespace)
		if err != nil {
			return err
		}

		runtimeToUpdate := runtime.DeepCopy()
		nowTime := metav1.Now()
		runtimeToUpdate.Status.MountTime = &nowTime

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			return e.Client.Status().Update(context.TODO(), runtimeToUpdate)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
