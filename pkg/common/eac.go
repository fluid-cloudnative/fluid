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

package common

// Runtime for EAC
const (
	EACRuntimeType                  = "eac"
	EACRuntimeResourceFinalizerName = "eac-runtime-controller-finalizer"
	EACRuntimeControllerName        = "EACRuntimeController"

	EACRuntime = EACRuntimeType

	EACChart = EACRuntime

	EACMountType = "alifuse.aliyun-alinas-eac"

	EACMasterImageEnv = "EAC_MASTER_IMAGE_ENV"

	EACFuseImageEnv = "EAC_FUSE_IMAGE_ENV"

	EACWorkerImageEnv = "EAC_WORKER_IMAGE_ENV"

	EACInitFuseImageEnv = "EAC_INIT_FUSE_IMAGE_ENV"

	DefaultEACMasterImage = "registry.cn-zhangjiakou.aliyuncs.com/nasteam/eac-fluid-img:update"

	DefaultEACFuseImage = "registry.cn-zhangjiakou.aliyuncs.com/nasteam/eac-fluid-img:update"

	DefaultEACWorkerImage = "registry.cn-zhangjiakou.aliyuncs.com/nasteam/eac-worker-img:update"

	DefaultEACInitFuseImage = "registry.cn-zhangjiakou.aliyuncs.com/nasteam/init-alifuse:update"
)

// Constants for EAC SessMgr
const (
	SessMgrNamespace     = "eac-system"
	SessMgrDaemonSetName = "eac-sessmgr"

	SessMgrNodeSelectorKey = "fluid.io/eac-sessmgr"

	EACSessMgrImageEnv = "EAC_SESSMGR_IMAGE_ENV"

	EACSessMgrUpdateStrategyEnv = "EAC_SESSMGR_UPDATE_STRATEGY_ENV"

	DefaultEACSessMgrImage = "registry.cn-zhangjiakou.aliyuncs.com/nascache/eac-fuse:v0.1.0-196d2b1"

	SessMgrSockFile = "sessmgrd.sock"

	VolumeAttrEACSessMgrWorkDir = "eac_sessmgr_workdir"
)
