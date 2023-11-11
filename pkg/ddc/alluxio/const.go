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

package alluxio

const (
	// NON_NATIVE_MOUNT_DATA_NAME also used in master 'statefulset.yaml' and config 'alluxio-mount.conf.yaml'
	NON_NATIVE_MOUNT_DATA_NAME = "mount.info"

	// alluxioHome string = "/opt/alluxio"

	// alluxioUser string = "fluid"
	metricsPrefixBytesReadLocal = "Cluster.BytesReadLocal "

	metricsPrefixBytesReadRemote = "Cluster.BytesReadRemote "

	metricsPrefixBytesReadUfsAll = "Cluster.BytesReadUfsAll "

	metricsPrefixBytesReadLocalThroughput = "Cluster.BytesReadLocalThroughput "

	metricsPrefixBytesReadRemoteThroughput = "Cluster.BytesReadRemoteThroughput "

	metricsPrefixBytesReadUfsThroughput = "Cluster.BytesReadUfsThroughput "

	summaryPrefixTotalCapacity = "Total Capacity: "

	summaryPrefixUsedCapacity = "Used Capacity: "

	metadataSyncNotDoneMsg = "[Calculating]"

	alluxioRuntimeMetricsLabel = "alluxio_runtime_metrics"

	checkMetadataSyncDoneTimeoutMillisec = 500

	portNum = 9

	cacheHitQueryIntervalMin = 1

	hadoopConfHdfsSiteFilename = "hdfs-site.xml"

	hadoopConfCoreSiteFilename = "core-site.xml"

	hadoopConfMountPath = "/hdfs-config"

	wokrerPodRole = "alluxio-worker"

	// defaultGracefulShutdownLimits is the limit for the system to forcibly clean up.
	defaultGracefulShutdownLimits       int32 = 3
	defaultCleanCacheGracePeriodSeconds int32 = 60

	MountConfigStorage   = "ALLUXIO_MOUNT_CONFIG_STORAGE"
	ConfigmapStorageName = "configmap"
)
