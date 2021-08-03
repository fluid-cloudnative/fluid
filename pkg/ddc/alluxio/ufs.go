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

package alluxio

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

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
	// 1. Mount UFS (Synchronous Operation)
	shouldMountUfs, err := e.shouldMountUFS()
	if err != nil {
		return
	}
	e.Log.Info("shouldMountUFS", "should", shouldMountUfs)

	if shouldMountUfs {
		err = e.mountUFS()
		if err != nil {
			return
		}
	}
	e.Log.Info("mountUFS")

	err = e.SyncMetadata()
	if err != nil {
		// just report this error and ignore it because SyncMetadata isn't on the critical path of Setup
		e.Log.Error(err, "SyncMetadata")
		return nil
	}

	return
}

func (e *AlluxioEngine) UpdateUFS(updatedUFSMap map[string][]string) (err error) {
	//1. set update status to updating
	err = utils.UpdateMountStatus(e.Client, e.name, e.namespace, datav1alpha1.UpdatingDatasetPhase)
	if err != nil {
		e.Log.Error(err, "Failed to update dataset status to updating")
		return err
	}
	//2. process added and removed
	err = e.processUpdatingUFS(updatedUFSMap)
	if err != nil {
		e.Log.Error(err, "Failed to add or remove mount points")
		return err
	}
	return nil
}

func (e *AlluxioEngine) UpdateOnUFSChange() (updateReady bool, err error) {
	// 1. get the dataset
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		e.Log.Error(err, "Failed to get the dataset")
		return
	}

	// 2. if no need to add/remove mount points, return false immediately
	if len(dataset.Spec.Mounts) == len(dataset.Status.Mounts) {
		return
	}

	// 3.get the update ufs map
	updateUFSMap, err := utils.GetUpdateUFSMap(dataset)
	if err != nil {
		e.Log.Error(err, "Failed to check mount points changes")
		return
	}

	if len(updateUFSMap) > 0 {
		//4. update the ufs
		err = e.UpdateUFS(updateUFSMap)
		if err != nil {
			e.Log.Error(err, "Failed to add or remove mount points")
			return
		}
		updateReady = true
	}

	return
}
