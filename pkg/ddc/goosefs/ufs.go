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

package goosefs

// UsedStorageBytes returns used storage size of GooseFS in bytes
func (e *GooseFSEngine) UsedStorageBytes() (value int64, err error) {
	// return e.usedStorageBytesInternal()
	return e.usedStorageBytesInternal()
}

// FreeStorageBytes returns free storage size of GooseFS in bytes
func (e *GooseFSEngine) FreeStorageBytes() (value int64, err error) {
	// return e.freeStorageBytesInternal()
	return e.freeStorageBytesInternal()
}

// TotalStorageBytes return total storage size of GooseFS in bytes
func (e *GooseFSEngine) TotalStorageBytes() (value int64, err error) {
	// return e.totalStorageBytesInternal()
	return e.totalStorageBytesInternal()
}

// TotalFileNums returns the total num of files in GooseFS
func (e *GooseFSEngine) TotalFileNums() (value int64, err error) {
	// return e.totalFileNumsInternal()
	return e.totalFileNumsInternal()
}

// ShouldCheckUFS checks if it requires checking UFS
func (e *GooseFSEngine) ShouldCheckUFS() (should bool, err error) {
	// For GooseFS Engine, always attempt to prepare UFS
	should = true
	return
}

// PrepareUFS does all the UFS preparations
func (e *GooseFSEngine) PrepareUFS() (err error) {
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

////du the ufs
//func (e *GooseFSEngine) du() (ufs int64, cached int64, cachedPercentage string, err error) {
//	podName, containerName := e.getMasterPodInfo()
//	fileUitls := operations.NewGooseFSFileUtils(podName, containerName, e.namespace, e.Log)
//	return fileUitls.Du("/")
//}
