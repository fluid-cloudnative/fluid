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

package goosefs

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

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

func (e *GooseFSEngine) ShouldUpdateUFS() (ufsToUpdate *utils.UFSToUpdate) {
	// 1. get the dataset
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		e.Log.Error(err, "Failed to get the dataset")
		return
	}

	// 2.get the ufs to update
	ufsToUpdate = utils.NewUFSToUpdate(dataset)
	ufsToUpdate.AnalyzePathsDelta()

	return
}

func (e *GooseFSEngine) UpdateOnUFSChange(ufsToUpdate *utils.UFSToUpdate) (updateReady bool, err error) {
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
	err = e.processUpdatingUFS(ufsToUpdate)
	if err != nil {
		e.Log.Error(err, "Failed to add or remove mount points")
		return
	}
	updateReady = true
	return
}
