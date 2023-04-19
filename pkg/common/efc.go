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

// Runtime for EFC
const (
	EFCRuntime = "efc"

	EFCChart = EFCRuntime

	EFCMountType = "alifuse.aliyun-alinas-eac"

	EFCRuntimeResourceFinalizerName = "efc-runtime-controller-finalizer"

	EFCRuntimeControllerName = "EFCRuntimeController"

	EFCMasterImageEnv = "EFC_MASTER_IMAGE_ENV"

	EFCFuseImageEnv = "EFC_FUSE_IMAGE_ENV"

	EFCWorkerImageEnv = "EFC_WORKER_IMAGE_ENV"

	EFCInitFuseImageEnv = "EFC_INIT_FUSE_IMAGE_ENV"

	DefaultEFCMasterImage = "registry.cn-zhangjiakou.aliyuncs.com/nascache/efc-master:latest"

	DefaultEFCFuseImage = "registry.cn-zhangjiakou.aliyuncs.com/nascache/efc-fuse:latest"

	DefaultEFCWorkerImage = "registry.cn-zhangjiakou.aliyuncs.com/nascache/efc-worker:latest"

	DefaultEFCInitFuseImage = "registry.cn-zhangjiakou.aliyuncs.com/nascache/init-alifuse:latest"
)

// Constants for EFC SessMgr
const (
	SessMgrNamespace = "efc-system"

	SessMgrDaemonSetName = "efc-sessmgr"

	SessMgrNodeSelectorKey = "fluid.io/efc-sessmgr"

	EFCSessMgrImageEnv = "EFC_SESSMGR_IMAGE_ENV"

	EFCSessMgrUpdateStrategyEnv = "EFC_SESSMGR_UPDATE_STRATEGY_ENV"

	DefaultEFCSessMgrImage = "registry.cn-zhangjiakou.aliyuncs.com/nascache/efc-fuse:latest"

	SessMgrSockFile = "sessmgrd.sock"

	VolumeAttrEFCSessMgrWorkDir = "efc_sessmgr_workdir"
)
