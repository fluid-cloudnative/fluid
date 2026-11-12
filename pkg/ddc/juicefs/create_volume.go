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
	"context"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	volumehelper "github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
)

func (j *JuiceFSEngine) CreateVolume(ctx context.Context) (err error) {
	if j.runtime == nil {
		j.runtime, err = j.getRuntime()
		if err != nil {
			return
		}
	}

	err = j.createFusePersistentVolume(ctx)
	if err != nil {
		return err
	}

	err = j.createFusePersistentVolumeClaim(ctx)
	if err != nil {
		return err
	}
	return
}

// createFusePersistentVolume
func (j *JuiceFSEngine) createFusePersistentVolume(ctx context.Context) (err error) {
	runtimeInfo, err := j.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumehelper.CreatePersistentVolumeForRuntime(ctx, j.Client,
		runtimeInfo,
		j.getMountPoint(),
		common.JuiceFSMountType,
		j.Log)
}

// createFusePersistentVolume
func (j *JuiceFSEngine) createFusePersistentVolumeClaim(ctx context.Context) (err error) {
	runtimeInfo, err := j.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumehelper.CreatePersistentVolumeClaimForRuntime(ctx, j.Client, runtimeInfo, j.Log)
}
