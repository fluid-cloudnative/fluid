/*
Copyright 2023 The Fluid Authors.

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
