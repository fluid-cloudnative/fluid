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
