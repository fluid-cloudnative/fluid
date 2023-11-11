/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package juicefs

import (
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

func (j JuiceFSEngine) UsedStorageBytes() (int64, error) {
	return j.usedSpaceInternal()
}

func (j JuiceFSEngine) FreeStorageBytes() (int64, error) {
	return 0, nil
}

func (j JuiceFSEngine) TotalStorageBytes() (int64, error) {
	return j.totalStorageBytesInternal()
}

func (j JuiceFSEngine) TotalFileNums() (int64, error) {
	return j.totalFileNumsInternal()
}

func (j JuiceFSEngine) ShouldCheckUFS() (should bool, err error) {
	return false, nil
}

func (j JuiceFSEngine) PrepareUFS() (err error) {
	return
}

// ShouldUpdateUFS JuiceFSEngine hasn't support UpdateOnUFSChange
func (j JuiceFSEngine) ShouldUpdateUFS() (ufsToUpdate *utils.UFSToUpdate) {
	return nil
}

func (j JuiceFSEngine) UpdateOnUFSChange(ufsToUpdate *utils.UFSToUpdate) (ready bool, err error) {
	return true, nil
}
