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

package thin

import "github.com/fluid-cloudnative/fluid/pkg/utils"

func (t *ThinEngine) UsedStorageBytes() (int64, error) {
	return t.usedSpaceInternal()
}

func (t *ThinEngine) FreeStorageBytes() (int64, error) {
	return 0, nil
}

func (t *ThinEngine) TotalStorageBytes() (int64, error) {
	return t.totalStorageBytesInternal()
}

func (t *ThinEngine) TotalFileNums() (int64, error) {
	return t.totalFileNumsInternal()
}

func (t ThinEngine) ShouldCheckUFS() (should bool, err error) {
	return false, nil
}

func (t ThinEngine) PrepareUFS() (err error) {
	return
}

func (t ThinEngine) ShouldUpdateUFS() (ufsToUpdate *utils.UFSToUpdate) {
	return nil
}

func (t ThinEngine) UpdateOnUFSChange(ufsToUpdate *utils.UFSToUpdate) (ready bool, err error) {
	return true, nil
}
