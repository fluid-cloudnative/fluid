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

package thin

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/juicefs/operations"
)

func (j *ThinEngine) totalStorageBytesInternal() (total int64, err error) {
	stsName := j.getFuseDaemonsetName()
	pods, err := j.GetRunningPodsOfDaemonset(stsName, j.namespace)
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

func (j *ThinEngine) totalFileNumsInternal() (fileCount int64, err error) {
	stsName := j.getFuseDaemonsetName()
	pods, err := j.GetRunningPodsOfDaemonset(stsName, j.namespace)
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

func (j *ThinEngine) usedSpaceInternal() (usedSpace int64, err error) {
	stsName := j.getFuseDaemonsetName()
	pods, err := j.GetRunningPodsOfDaemonset(stsName, j.namespace)
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
