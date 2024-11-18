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

package jindo

const (
	MetadataSyncNotDoneMsg = "[Calculating]"

	CheckMetadataSyncDoneTimeoutMillisec = 500

	HadoopConfHdfsSiteFilename = "hdfs-site.xml"

	HadoopConfCoreSiteFilename = "core-site.xml"

	JindoMasterNumDefault = 1
	JindoHAMasterNum      = 3

	DefaultMasterRpcPort = 8101
	DefaultWorkerRpcPort = 6101
	DefaultRaftRpcPort   = 8103

	WorkerPodRole = "jindo-worker"

	RuntimeFSType = "jindofs"

	JindoFuseMountPath = "/jfs/jindofs-fuse"

	DefaultJindoRuntimeImage = "registry.cn-shanghai.aliyuncs.com/jindofs/smartdata:3.8.0"
)
