/*
Copyright 2023 The Fluid Author.

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

package alluxio

import (
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
)

func IsMountWithConfigMap() bool {
	if envVal, exists := os.LookupEnv(MountConfigStorage); exists {
		return envVal == ConfigmapStorageName
	}
	// default value
	return true
}

// UsedStorageBytes returns used storage size of Alluxio in bytes
func (e *AlluxioEngine) UsedStorageBytes() (value int64, err error) {
	// return e.usedStorageBytesInternal()
	return e.usedStorageBytesInternal()
}

// FreeStorageBytes returns free storage size of Alluxio in bytes
func (e *AlluxioEngine) FreeStorageBytes() (value int64, err error) {
	// return e.freeStorageBytesInternal()
	return e.freeStorageBytesInternal()
}

// TotalStorageBytes returns total storage size of Alluxio in bytes
func (e *AlluxioEngine) TotalStorageBytes() (value int64, err error) {
	// return e.totalStorageBytesInternal()
	return e.totalStorageBytesInternal()
}

// TotalFileNums returns the total num of files in Alluxio
func (e *AlluxioEngine) TotalFileNums() (value int64, err error) {
	// return e.totalFileNumsInternal()
	return e.totalFileNumsInternal()
}

// ShouldCheckUFS checks if it requires checking UFS
func (e *AlluxioEngine) ShouldCheckUFS() (should bool, err error) {
	// For Alluxio Engine, always attempt to prepare UFS
	should = true
	return
}

// PrepareUFS does all the UFS preparations
func (e *AlluxioEngine) PrepareUFS() (err error) {
	// if using configmap to store mount info, no need to execute ufs mount in alluxio master pod in `Setup` phase.
	usingConfigMap := IsMountWithConfigMap()
	if !usingConfigMap {
		// 1. Mount UFS (Synchronous Operation)
		shouldMountUfs, err := e.shouldMountUFS()
		if err != nil {
			return err
		}
		e.Log.Info("shouldMountUFS", "should", shouldMountUfs)

		if shouldMountUfs {
			err = e.mountUFS()
			if err != nil {
				return err
			}
		}
		e.Log.Info("mountUFS")
	}
	err = e.SyncMetadata()
	if err != nil {
		// just report this error and ignore it because SyncMetadata isn't on the critical path of Setup
		e.Log.Error(err, "SyncMetadata")
		return nil
	}

	return
}

func (e *AlluxioEngine) ShouldUpdateUFS() (ufsToUpdate *utils.UFSToUpdate) {
	// 1. get the dataset
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		e.Log.Error(err, "Failed to get the dataset")
		return
	}

	// 2.get the ufs to update
	ufsToUpdate = utils.NewUFSToUpdate(dataset)
	ufsToUpdate.AnalyzePathsDelta()

	// 3. for hostpath ufs mount, check if all mountpoints have been mounted
	e.checkIfRemountRequired(ufsToUpdate)

	return
}

func (e *AlluxioEngine) UpdateOnUFSChange(ufsToUpdate *utils.UFSToUpdate) (updateReady bool, err error) {
	// 1. check if need to update ufs
	if !ufsToUpdate.ShouldUpdate() {
		e.Log.Info("no need to update ufs",
			"namespace", e.namespace,
			"name", e.name)
		return
	}

	// 2. set update status to updating
	err = utils.UpdateMountStatus(e.Client, e.name, e.namespace, datav1alpha1.UpdatingDatasetPhase)
	if err != nil {
		e.Log.Error(err, "Failed to update dataset status to updating")
		return
	}

	// 3. process added and removed
	updateReady, err = e.processUpdatingUFS(ufsToUpdate)
	if err != nil {
		e.Log.Error(err, "Failed to add or remove mount points")
		return
	}

	return
}

func (e *AlluxioEngine) checkIfRemountRequired(ufsToUpdate *utils.UFSToUpdate) {
	runtime, err := e.getRuntime()
	if err != nil {
		e.Log.Error(err, "checkIfRemountRequired", "runtime", e.name)
		return
	}

	masterPodName, masterContainerName := e.getMasterPodInfo()
	masterPod, err := e.getMasterPod(masterPodName, e.namespace)
	if err != nil {
		e.Log.Error(err, "checkIfRemountRequired", "master pod", e.name)
		return
	}

	var startedAt *v1.Time
	for _, containerStatus := range masterPod.Status.ContainerStatuses {
		if containerStatus.Name == masterContainerName {
			if containerStatus.State.Running == nil {
				e.Log.Error(fmt.Errorf("container is not running"), "checkIfRemountRequired", "master pod", masterPodName)
				return
			} else {
				startedAt = &containerStatus.State.Running.StartedAt
				break
			}
		}
	}

	// If mounttime is earlier than master container starttime, remount is necessary
	if startedAt != nil && runtime.Status.MountTime != nil && runtime.Status.MountTime.Before(startedAt) {
		e.Log.Info("remount on master restart", "alluxioruntime", e.name)

		unmountedPaths, err := e.FindUnmountedUFS()
		if err != nil {
			e.Log.Error(err, "Failed in finding unmounted ufs")
			return
		}

		if len(unmountedPaths) != 0 {
			ufsToUpdate.AddMountPaths(unmountedPaths)
		} else {
			// if no path can be mounted, set mountTime to be now
			e.updateMountTime()
		}
	}
}
