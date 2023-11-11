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

package goosefs

const (

	// goosefsHome string = "/opt/goosefs"

	// goosefsUser string = "fluid"
	METRICS_PREFIX_BYTES_READ_LOCAL = "Cluster.BytesReadLocal "

	METRICS_PREFIX_BYTES_READ_REMOTE = "Cluster.BytesReadRemote "

	METRICS_PREFIX_BYTES_READ_UFS_ALL = "Cluster.BytesReadUfsAll "

	METRICS_PREFIX_BYTES_READ_LOCAL_THROUGHPUT = "Cluster.BytesReadLocalThroughput "

	METRICS_PREFIX_BYTES_READ_REMOTE_THROUGHPUT = "Cluster.BytesReadRemoteThroughput "

	METRICS_PREFIX_BYTES_READ_UFS_THROUGHPUT = "Cluster.BytesReadUfsThroughput "

	SUMMARY_PREFIX_TOTAL_CAPACITY = "Total Capacity: "

	SUMMARY_PREFIX_USED_CAPACITY = "Used Capacity: "

	METADATA_SYNC_NOT_DONE_MSG = "[Calculating]"

	GOOSEFS_RUNTIME_METRICS_LABEL = "goosefs_runtime_metrics"

	CHECK_METADATA_SYNC_DONE_TIMEOUT_MILLISEC = 500

	AUTO_SELECT_PORT_MIN = 20000
	AUTO_SELECT_PORT_MAX = 30000

	PORT_NUM = 9

	CACHE_HIT_QUERY_INTERVAL_MIN = 1

	HADOOP_CONF_HDFS_SITE_FILENAME = "hdfs-site.xml"

	HADOOP_CONF_CORE_SITE_FILENAME = "core-site.xml"

	HADOOP_CONF_MOUNT_PATH = "/hdfs-config"

	WOKRER_POD_ROLE = "goosefs-worker"
)
