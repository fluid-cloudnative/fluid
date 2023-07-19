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

package jindocache

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindocache/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
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

	// 2. Setup cacheset
	shouldRefresh, err := e.ShouldRefreshCacheSet()
	if err != nil {
		return
	}
	e.Log.Info("shouldMountUFS", "should", shouldMountUfs)
	if shouldRefresh {
		err = e.RefreshCacheSet()
		if err != nil {
			return
		}
	}

	// 3. SyncMetadata
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
