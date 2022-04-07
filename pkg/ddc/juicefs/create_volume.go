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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	volumehelper "github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
)

func (j *JuiceFSEngine) CreateVolume() (err error) {
	if j.runtime == nil {
		j.runtime, err = j.getRuntime()
		if err != nil {
			return
		}
	}

	err = j.createFusePersistentVolume()
	if err != nil {
		return err
	}

	err = j.createFusePersistentVolumeClaim()
	if err != nil {
		return err
	}
	return
}

// createFusePersistentVolume
func (j *JuiceFSEngine) createFusePersistentVolume() (err error) {
	runtimeInfo, err := j.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumehelper.CreatePersistentVolumeForRuntime(j.Client,
		runtimeInfo,
		j.getMountPoint(),
		common.JuiceFSMountType,
		j.Log)
}

// createFusePersistentVolume
func (j *JuiceFSEngine) createFusePersistentVolumeClaim() (err error) {
	runtimeInfo, err := j.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumehelper.CreatePersistentVolumeClaimForRuntime(j.Client, runtimeInfo, j.Log)
}
