/*
Copyright 2024 The Fluid Authors.
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

// Runtime for Vineyard
const (
	VineyardRuntime = "vineyard"

	VineyardMountType = "vineyard-fuse"

	VineyardChart = VineyardRuntime

	VineyardMasterImageEnv = "VINEYARD_MASTER_IMAGE_ENV"

	VineyardWorkerImageEnv = "VINEYARD_WORKER_IMAGE_ENV"

	VineyardFuseImageEnv = "VINEYARD_FUSE_IMAGE_ENV"

	VineyardFuseIsGlobal = true

	DefaultVineyardMasterImage = "registry.aliyuncs.com/vineyard/vineyardd:v0.22.1"

	DefaultVineyardWorkerImage = "registry.aliyuncs.com/vineyard/vineyardd:v0.22.1"

	DefultVineyardFuseImage = "registry.aliyuncs.com/vineyard/vineyard-fluid-fuse:v0.22.1"

	VineyardEngineImpl = VineyardRuntime

	VineyardConfigmapSuffix = "-rpc-conf"

	VineyardConfigmapVolumeName = "vineyard-rpc-conf"

	VineyardRPCEndpoint = "VINEYARD_RPC_ENDPOINT"
)

var (
	VineyardFuseNodeSelector = map[string]string{}
)
