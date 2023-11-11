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

package efc

import (
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

func (e *EFCEngine) UsedStorageBytes() (int64, error) {
	return 0, nil
}

func (e *EFCEngine) FreeStorageBytes() (int64, error) {
	return 0, nil
}

func (e *EFCEngine) TotalStorageBytes() (int64, error) {
	//mountInfo, err := e.getMountInfo()
	//if err != nil {
	//	return 0, err
	//}
	//response, err := mountInfo.DescribeDirQuota()
	//if err != nil {
	//	return 0, err
	//}
	//if len(response.DirQuotaInfos) == 0 || len(response.DirQuotaInfos[0].UserQuotaInfos) == 0 {
	//	return 0, fmt.Errorf("invalid DescribeDirQuotasResponse size")
	//}
	//base := resource.MustParse("1Gi")
	//return response.DirQuotaInfos[0].UserQuotaInfos[0].SizeReal * base.Value(), nil
	return 0, nil
}

func (e *EFCEngine) TotalFileNums() (int64, error) {
	//mountInfo, err := e.getMountInfo()
	//if err != nil {
	//	return 0, err
	//}
	//response, err := mountInfo.DescribeDirQuota()
	//if err != nil {
	//	return 0, err
	//}
	//if len(response.DirQuotaInfos) == 0 || len(response.DirQuotaInfos[0].UserQuotaInfos) == 0 {
	//	return 0, fmt.Errorf("invalid DescribeDirQuotasResponse size")
	//}
	//return response.DirQuotaInfos[0].UserQuotaInfos[0].FileCountReal, nil
	return 0, nil
}

func (e *EFCEngine) ShouldCheckUFS() (should bool, err error) {
	//mountInfo, err := e.getMountInfo()
	//if err != nil {
	//	return false, err
	//}
	//if len(mountInfo.AccessKeyID) == 0 || len(mountInfo.AccessKeySecret) == 0 {
	//	return false, nil
	//}

	return false, nil
}

func (e *EFCEngine) PrepareUFS() (err error) {
	//mountInfo, err := e.getMountInfo()
	//if err != nil {
	//	return err
	//}
	//_, err = mountInfo.SetDirQuota()
	//if err != nil {
	//	e.Log.Error(err, "Failed to set dir quota")
	//	return err
	//}
	return nil
}

// ShouldUpdateUFS EFCEngine hasn't support UpdateOnUFSChange
func (e *EFCEngine) ShouldUpdateUFS() (ufsToUpdate *utils.UFSToUpdate) {
	return nil
}

func (e *EFCEngine) UpdateOnUFSChange(ufsToUpdate *utils.UFSToUpdate) (ready bool, err error) {
	return true, nil
}
