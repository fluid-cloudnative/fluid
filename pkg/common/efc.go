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

package common

// Runtime for EFC
const (
	EFCRuntime = "efc"

	EFCChart = EFCRuntime

	EFCMountType = "alifuse.aliyun-alinas-efc"

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
