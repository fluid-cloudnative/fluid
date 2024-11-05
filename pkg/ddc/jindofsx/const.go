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

package jindofsx

const (
	CSI_DRIVER = "fuse.csi.fluid.io"

	//fluid_PATH = "fluid_path"

	Mount_TYPE = "mount_type"

	SUMMARY_PREFIX_TOTAL_MEM_CAPACITY = "Total MEM Capacity: "

	SUMMARY_PREFIX_USED_MEM_CAPACITY = "Used MEM Capacity: "

	METADATA_SYNC_NOT_DONE_MSG = "[Calculating]"

	CHECK_METADATA_SYNC_DONE_TIMEOUT_MILLISEC = 500

	HADOOP_CONF_HDFS_SITE_FILENAME = "hdfs-site.xml"

	HADOOP_CONF_CORE_SITE_FILENAME = "core-site.xml"

	JINDO_MASTERNUM_DEFAULT = 1
	JINDO_HA_MASTERNUM      = 3

	DEFAULT_MASTER_RPC_PORT = 8101
	DEFAULT_WORKER_RPC_PORT = 6101
	DEFAULT_RAFT_RPC_PORT   = 8103

	workerPodRole = "jindo-worker"

	runtimeFSType = "jindofs"

	jindoFuseMountpath = "/jfs/jindofs-fuse"

	defaultJindofsxRuntimeImage = "registry.cn-shanghai.aliyuncs.com/jindofs/smartdata:4.6.8"

	FuseOnly = "fuseOnly"

	defaultMemLimit = 100

	defaultMetaSize = "30Gi"

	QueryUfsTotal = "QUERY_UFS_TOTAL"

	imageTagSupportAKFile = "4.6.8"
)
