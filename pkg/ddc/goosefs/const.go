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
	MetricsPrefixBytesReadLocal = "Cluster.BytesReadLocal "

	MetricsPrefixBytesReadRemote = "Cluster.BytesReadRemote "

	MetricsPrefixBytesReadUfsAll = "Cluster.BytesReadUfsAll "

	MetricsPrefixBytesReadLocalThroughput = "Cluster.BytesReadLocalThroughput "

	MetricsPrefixBytesReadRemoteThroughput = "Cluster.BytesReadRemoteThroughput "

	MetricsPrefixBytesReadUfsThroughput = "Cluster.BytesReadUfsThroughput "

	MetadataSyncNotDoneMsg = "[Calculating]"

	GooseFSRuntimeMetricsLabel = "goosefs_runtime_metrics"

	CheckMetadataSyncDoneTimeoutMillisec = 500

	AUTO_SELECT_PORT_MIN = 20000
	AUTO_SELECT_PORT_MAX = 30000

	PortNum = 9

	CacheHitQueryIntervalMin = 1

	HadoopConfHdfsSiteFilename = "hdfs-site.xml"

	HadoopConfCoreSiteFilename = "core-site.xml"

	HadoopConfMountPath = "/hdfs-config"

	WokrerPodRole = "goosefs-worker"
)
