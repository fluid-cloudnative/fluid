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

package jindo

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindo/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// ShouldCheckUFS checks if it requires checking UFS
func (e *JindoEngine) ShouldCheckUFS() (should bool, err error) {
	should = true
	return
}

// PrepareUFS do all the UFS preparations
func (e *JindoEngine) PrepareUFS() (err error) {
	// For Jindo Engine, not need to prepare UFS
	return
}

// UsedStorageBytes returns used storage size of Jindo in bytes
func (e *JindoEngine) UsedStorageBytes() (value int64, err error) {

	return
}

// FreeStorageBytes returns free storage size of Jindo in bytes
func (e *JindoEngine) FreeStorageBytes() (value int64, err error) {
	return
}

// return total storage size of Jindo in bytes
func (e *JindoEngine) TotalStorageBytes() (value int64, err error) {
	return
}

// return the total num of files in Jindo
func (e *JindoEngine) TotalFileNums() (value int64, err error) {
	return
}

// report jindo summary
func (e *JindoEngine) GetReportSummary() (summary string, err error) {
	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewJindoFileUtils(podName, containerName, e.namespace, e.Log)
	return fileUtils.ReportSummary()
}

// JindoEngine hasn't support UpdateOnUFSChange
func (e *JindoEngine) ShouldUpdateUFS() (ufsToUpdate *utils.UFSToUpdate) {
	return
}

func (e *JindoEngine) UpdateOnUFSChange(*utils.UFSToUpdate) (updateReady bool, err error) {
	return
}
