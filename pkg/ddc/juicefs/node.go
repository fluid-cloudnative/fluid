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

package juicefs

import (
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	datasetschedule "github.com/fluid-cloudnative/fluid/pkg/utils/dataset/lifecycle"
)

func (j JuiceFSEngine) AssignNodesToCache(desiredNum int32) (currentScheduleNum int32, err error) {
	runtimeInfo, err := j.getRuntimeInfo()
	if err != nil {
		return currentScheduleNum, err
	}

	dataset, err := utils.GetDataset(j.Client, j.name, j.namespace)
	if err != nil {
		return
	}

	j.Log.Info("AssignNodesToCache", "dataset", dataset)
	return datasetschedule.AssignDatasetToNodes(runtimeInfo,
		dataset,
		j.Client,
		desiredNum)
}

func (j *JuiceFSEngine) SyncScheduleInfoToCacheNodes() (err error) {
	return
}
