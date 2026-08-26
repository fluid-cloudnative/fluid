/*
Copyright 2020 The Fluid Authors.

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

import "time"

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

	// defaultWorkerWebPort is the fallback Alluxio worker web port used only
	// when a worker pod's own container spec doesn't expose one (see
	// workerWebPortFromPod). decommissionWorker addresses this port, not the
	// RPC port, since it monitors the worker's workload as exposed on the web
	// server.
	defaultWorkerWebPort = 30000

	// alluxioWorkerContainerName is the name charts/alluxio gives the
	// Alluxio worker container, see
	// charts/alluxio/templates/worker/statefulset.yaml. Used to pick the
	// worker's own "web" container port apart from the alluxio-job-worker
	// container, which exposes a differently-numbered port of the same name.
	alluxioWorkerContainerName = "alluxio-worker"

	MountConfigStorage   = "ALLUXIO_MOUNT_CONFIG_STORAGE"
	ConfigmapStorageName = "configmap"

	// defaultWorkerDecommissionDeadline bounds how long the engine keeps
	// retrying a stuck worker drain (e.g. an unhealthy master, unreplicable
	// blocks) before forcing the scale-down to proceed anyway, rather than
	// stalling on every reconcile indefinitely.
	defaultWorkerDecommissionDeadline = 10 * time.Minute
)
