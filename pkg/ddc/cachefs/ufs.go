/*
Copyright 2021 The Fluid Authors.

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

package cachefs

import (
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

func (j *CacheFSEngine) UsedStorageBytes() (int64, error) {
	return j.usedSpaceInternal()
}

func (j *CacheFSEngine) FreeStorageBytes() (int64, error) {
	return 0, nil
}

func (j *CacheFSEngine) TotalStorageBytes() (int64, error) {
	return j.totalStorageBytesInternal()
}

func (j *CacheFSEngine) TotalFileNums() (int64, error) {
	return j.totalFileNumsInternal()
}

func (j *CacheFSEngine) ShouldCheckUFS() (should bool, err error) {
	return false, nil
}

func (j *CacheFSEngine) PrepareUFS() (err error) {
	return
}

// ShouldUpdateUFS JuiceFSEngine hasn't support UpdateOnUFSChange
func (j *CacheFSEngine) ShouldUpdateUFS() (ufsToUpdate *utils.UFSToUpdate) {
	return nil
}

func (j *CacheFSEngine) UpdateOnUFSChange(ufsToUpdate *utils.UFSToUpdate) (ready bool, err error) {
	return true, nil
}
