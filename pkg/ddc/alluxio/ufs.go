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

func (e *AlluxioEngine) GetUpdateUFSMap() (map[string][]string, error) {
	updateUFSMap := make(map[string][]string)
	// For Alluxio Engine, always attempt to prepare UFS
	resultInCtx, resultHaveMounted, err := e.getMounts()

	// 2. get mount point need to be added and removed
	//var added, removed []string
	added, removed := e.calculateMountPointsChanges(resultHaveMounted, resultInCtx)

	if len(added) > 0 {
		updateUFSMap["added"] = added
	}

	if len(removed) > 0 {
		updateUFSMap["removed"] = removed
	}

	return updateUFSMap, err
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
	errUpdating := e.SetUFSUpdating()
	if errUpdating != nil {
		e.Log.Error(err, "Failed to update dataset status to updating")
		return err
	}
	//2. process added and removed
	err = e.processUpdatingUFS(updatedUFSMap)
	if err != nil {
		e.Log.Error(err, "Failed to add or remove mount points")
		return err
	}

	//3. update dataset status to updated
	err = e.SetUFSUpdated()
	if err != nil {
		e.Log.Error(err, "Failed to update dataset status to updated")
		return err
	}

	return err
}

func (e *AlluxioEngine) UpdateOnUFSChange() (updateReady bool, err error) {
	// 1.get the updated ufs map
	// updatedUFSMap, err
	updatedUFSMap, err := e.GetUpdateUFSMap()

	if err != nil {
		e.Log.Error(err, "Failed to check mount points changes")
	}

	if len(updatedUFSMap) > 0 {
		//2. update the ufs
		err := e.UpdateUFS(updatedUFSMap)
		if err != nil {
			e.Log.Error(err, "Failed to add or remove mount points")
		}
	}

	updateReady = true

	return updateReady, err
}

////du the ufs
//func (e *AlluxioEngine) du() (ufs int64, cached int64, cachedPercentage string, err error) {
//	podName, containerName := e.getMasterPodInfo()
//	fileUitls := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)
//	return fileUitls.Du("/")
//}
