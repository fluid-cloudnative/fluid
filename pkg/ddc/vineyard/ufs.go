/*
Copyright 2024 The Fluid Authors.
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

package vineyard

import "github.com/fluid-cloudnative/fluid/pkg/utils"

// UsedStorageBytes returns used storage size of Vineyard in bytes
func (e *VineyardEngine) UsedStorageBytes() (value int64, err error) {
	// return e.usedStorageBytesInternal()
	return 0, nil
}

// FreeStorageBytes returns free storage size of Vineyard in bytes
func (e *VineyardEngine) FreeStorageBytes() (value int64, err error) {
	// return e.freeStorageBytesInternal()
	return 0, nil
}

// TotalStorageBytes returns total storage size of Vineyard in bytes
func (e *VineyardEngine) TotalStorageBytes() (value int64, err error) {
	// return e.totalStorageBytesInternal()
	return 0, nil
}

// TotalFileNums returns the total num of files in Vineyard
func (e *VineyardEngine) TotalFileNums() (value int64, err error) {
	// return e.totalFileNumsInternal()
	return 0, nil
}

// ShouldCheckUFS checks if it requires checking UFS
func (e *VineyardEngine) ShouldCheckUFS() (should bool, err error) {
	// For Vineyard Engine, always attempt to prepare UFS
	should = true
	return
}

func (e *VineyardEngine) ShouldUpdateUFS() (ufsToUpdate *utils.UFSToUpdate) {
	return
}

// PrepareUFS does all the UFS preparations
func (e *VineyardEngine) PrepareUFS() (err error) {

	return
}

func (e *VineyardEngine) UpdateOnUFSChange(ufsToUpdate *utils.UFSToUpdate) (updateReady bool, err error) {

	return
}
