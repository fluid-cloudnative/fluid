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

package jindocache

const (
	CSI_DRIVER = "fuse.csi.fluid.io"

	//fluid_PATH = "fluid_path"

	Mount_TYPE = "mount_type"

	SUMMARY_PREFIX_TOTAL_CAPACITY = "Total Disk Capacity: "

	SUMMARY_PREFIX_USED_CAPACITY = "Used Disk Capacity: "

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

	defaultJindofsxRuntimeImage = "registry.cn-shanghai.aliyuncs.com/jindofs/smartdata:5.0.0"

	engineType = "jindocache"

	FuseOnly = "fuseOnly"

	defaultMemLimit = 100

	defaultMetaSize = "30Gi"

	QueryUfsTotal = "QUERY_UFS_TOTAL"

	imageTagSupportAKFile = "4.6.8"

	// Write Policy
	WriteAround  = "WRITE_AROUND"
	WriteThrough = "WRITE_THROUGH"
	CacheOnly    = "CACHE_ONLY"
)
