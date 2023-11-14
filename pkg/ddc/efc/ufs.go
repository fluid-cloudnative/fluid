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
