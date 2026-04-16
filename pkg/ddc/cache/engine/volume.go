/*
  Copyright 2026 The Fluid Authors.

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

package engine

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	volumeHelper "github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
)

func (e *CacheEngine) CreateVolume() (err error) {
	if err = e.createFusePersistentVolume(); err != nil {
		return err
	}

	if err = e.createFusePersistentVolumeClaim(); err != nil {
		return err
	}
	return nil
}

func (e *CacheEngine) DeleteVolume() (err error) {
	if err = e.deleteFusePersistentVolumeClaim(); err != nil {
		return err
	}

	if err = e.deleteFusePersistentVolume(); err != nil {
		return err
	}

	return nil
}

func (e *CacheEngine) createFusePersistentVolume() error {
	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumeHelper.CreatePersistentVolumeForRuntime(e.Client,
		runtimeInfo,
		e.getFuseMountPoint(),
		common.CacheRuntime,
		e.Log)
}

func (e *CacheEngine) createFusePersistentVolumeClaim() error {
	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumeHelper.CreatePersistentVolumeClaimForRuntime(e.Client, runtimeInfo, e.Log)
}

func (e *CacheEngine) deleteFusePersistentVolume() error {
	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumeHelper.DeleteFusePersistentVolume(e.Client, runtimeInfo, e.Log)
}

func (e *CacheEngine) deleteFusePersistentVolumeClaim() error {
	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumeHelper.DeleteFusePersistentVolumeClaim(e.Client, runtimeInfo, e.Log)
}
