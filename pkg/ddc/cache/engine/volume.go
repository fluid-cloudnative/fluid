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
	"context"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	volumeHelper "github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
)

func (e *CacheEngine) CreateVolume(ctx context.Context) (err error) {
	if err = e.createFusePersistentVolume(ctx); err != nil {
		return err
	}

	if err = e.createFusePersistentVolumeClaim(ctx); err != nil {
		return err
	}
	return nil
}

func (e *CacheEngine) DeleteVolume(ctx context.Context) (err error) {
	if err = e.deleteFusePersistentVolumeClaim(ctx); err != nil {
		return err
	}

	if err = e.deleteFusePersistentVolume(ctx); err != nil {
		return err
	}

	return nil
}

func (e *CacheEngine) createFusePersistentVolume(ctx context.Context) error {
	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumeHelper.CreatePersistentVolumeForRuntime(ctx,
		e.Client,
		runtimeInfo,
		e.getFuseMountPoint(),
		common.CacheRuntime,
		e.Log)
}

func (e *CacheEngine) createFusePersistentVolumeClaim(ctx context.Context) error {
	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumeHelper.CreatePersistentVolumeClaimForRuntime(ctx, e.Client, runtimeInfo, e.Log)
}

func (e *CacheEngine) deleteFusePersistentVolume(ctx context.Context) error {
	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumeHelper.DeleteFusePersistentVolume(ctx, e.Client, runtimeInfo, e.Log)
}

func (e *CacheEngine) deleteFusePersistentVolumeClaim(ctx context.Context) error {
	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumeHelper.DeleteFusePersistentVolumeClaim(ctx, e.Client, runtimeInfo, e.Log)
}
