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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/juicefs/operations"
)

func (j *JuiceFSEngine) totalStorageBytesInternal() (total int64, err error) {
	stsName := j.getWorkerName()
	pods, err := j.GetRunningPodsOfStatefulSet(stsName, j.namespace)
	if err != nil || len(pods) == 0 {
		return
	}
	fileUtils := operations.NewJuiceFileUtils(pods[0].Name, common.JuiceFSWorkerContainer, j.namespace, j.Log)
	total, err = fileUtils.GetUsedSpace(j.getMountPoint())
	if err != nil {
		return
	}

	return
}

func (j *JuiceFSEngine) totalFileNumsInternal() (fileCount int64, err error) {
	stsName := j.getWorkerName()
	pods, err := j.GetRunningPodsOfStatefulSet(stsName, j.namespace)
	if err != nil || len(pods) == 0 {
		return
	}
	fileUtils := operations.NewJuiceFileUtils(pods[0].Name, common.JuiceFSWorkerContainer, j.namespace, j.Log)
	fileCount, err = fileUtils.GetFileCount(j.getMountPoint())
	if err != nil {
		return
	}

	return
}

func (j *JuiceFSEngine) usedSpaceInternal() (usedSpace int64, err error) {
	stsName := j.getWorkerName()
	pods, err := j.GetRunningPodsOfStatefulSet(stsName, j.namespace)
	if err != nil || len(pods) == 0 {
		return
	}
	fileUtils := operations.NewJuiceFileUtils(pods[0].Name, common.JuiceFSWorkerContainer, j.namespace, j.Log)
	usedSpace, err = fileUtils.GetUsedSpace(j.getMountPoint())
	if err != nil {
		return
	}

	return
}
