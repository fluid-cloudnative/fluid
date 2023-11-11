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

package jindofsx

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindofsx/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// ShouldCheckUFS checks if it requires checking UFS
func (e *JindoFSxEngine) ShouldCheckUFS() (should bool, err error) {
	should = true
	return
}

// PrepareUFS do all the UFS preparations
func (e *JindoFSxEngine) PrepareUFS() (err error) {

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

	err = e.SyncMetadata()
	if err != nil {
		// just report this error and ignore it because SyncMetadata isn't on the critical path of Setup
		e.Log.Error(err, "SyncMetadata")
		return nil
	}

	return
}

// UsedStorageBytes returns used storage size of Jindo in bytes
func (e *JindoFSxEngine) UsedStorageBytes() (value int64, err error) {

	return
}

// FreeStorageBytes returns free storage size of Jindo in bytes
func (e *JindoFSxEngine) FreeStorageBytes() (value int64, err error) {
	return
}

// return total storage size of Jindo in bytes
func (e *JindoFSxEngine) TotalStorageBytes() (value int64, err error) {
	return
}

// return the total num of files in Jindo
func (e *JindoFSxEngine) TotalFileNums() (value int64, err error) {
	return
}

// report jindo summary
func (e *JindoFSxEngine) GetReportSummary() (summary string, err error) {
	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewJindoFileUtils(podName, containerName, e.namespace, e.Log)
	return fileUtils.ReportSummary()
}

// JindoFSxEngine hasn't support UpdateOnUFSChange
func (e *JindoFSxEngine) ShouldUpdateUFS() (ufsToUpdate *utils.UFSToUpdate) {
	return
}

func (e *JindoFSxEngine) UpdateOnUFSChange(*utils.UFSToUpdate) (updateReady bool, err error) {
	return
}
