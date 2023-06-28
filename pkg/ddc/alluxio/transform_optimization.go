/*

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

import (
	"strconv"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	v1 "k8s.io/api/core/v1"
)

// transform dataset which has ufsPaths and ufsVolumes
func (e *AlluxioEngine) optimizeDefaultProperties(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) {
	if len(value.Properties) == 0 {
		if len(runtime.Spec.Properties) > 0 {
			value.Properties = runtime.Spec.Properties
		} else {
			value.Properties = map[string]string{}
		}
	}

	setDefaultProperties(runtime, value, "alluxio.fuse.jnifuse.enabled", "true")
	setDefaultProperties(runtime, value, "alluxio.master.metastore", "ROCKS")
	setDefaultProperties(runtime, value, "alluxio.web.ui.enabled", "false")
	setDefaultProperties(runtime, value, "alluxio.user.update.file.accesstime.disabled", "true")
	setDefaultProperties(runtime, value, "alluxio.user.client.cache.enabled", "false")
	setDefaultProperties(runtime, value, "alluxio.master.metastore.inode.cache.max.size", "10000000")
	setDefaultProperties(runtime, value, "alluxio.master.journal.log.size.bytes.max", "500MB")
	setDefaultProperties(runtime, value, "alluxio.master.metadata.sync.concurrency.level", "128")
	setDefaultProperties(runtime, value, "alluxio.master.metadata.sync.executor.pool.size", "128")
	setDefaultProperties(runtime, value, "alluxio.master.metadata.sync.ufs.prefetch.pool.size", "128")
	setDefaultProperties(runtime, value, "alluxio.user.block.worker.client.pool.min", "512")
	setDefaultProperties(runtime, value, "alluxio.fuse.debug.enabled", "false")
	setDefaultProperties(runtime, value, "alluxio.web.ui.enabled", "false")
	setDefaultProperties(runtime, value, "alluxio.user.file.writetype.default", "MUST_CACHE")
	setDefaultProperties(runtime, value, "alluxio.user.ufs.block.read.location.policy", "alluxio.client.block.policy.LocalFirstPolicy")
	setDefaultProperties(runtime, value, "alluxio.user.block.write.location.policy.class", "alluxio.client.block.policy.LocalFirstAvoidEvictionPolicy")
	setDefaultProperties(runtime, value, "alluxio.worker.allocator.class", "alluxio.worker.block.allocator.MaxFreeAllocator")
	setDefaultProperties(runtime, value, "alluxio.user.block.size.bytes.default", "16MB")
	setDefaultProperties(runtime, value, "alluxio.user.streaming.reader.chunk.size.bytes", "32MB")
	setDefaultProperties(runtime, value, "alluxio.user.local.reader.chunk.size.bytes", "32MB")
	setDefaultProperties(runtime, value, "alluxio.worker.network.reader.buffer.size", "32MB")
	// Enable metrics as default for better monitoring result, if you have performance concern, feel free to turn it off
	setDefaultProperties(runtime, value, "alluxio.user.metrics.collection.enabled", "true")
	setDefaultProperties(runtime, value, "alluxio.master.rpc.executor.max.pool.size", "1024")
	setDefaultProperties(runtime, value, "alluxio.master.rpc.executor.core.pool.size", "128")
	// setDefaultProperties(runtime, value, "alluxio.master.mount.table.root.readonly", "true")
	setDefaultProperties(runtime, value, "alluxio.user.update.file.accesstime.disabled", "true")
	setDefaultProperties(runtime, value, "alluxio.user.file.passive.cache.enabled", "false")
	setDefaultProperties(runtime, value, "alluxio.user.block.avoid.eviction.policy.reserved.size.bytes", "2GB")
	setDefaultProperties(runtime, value, "alluxio.master.journal.folder", "/journal")
	if value.Master.Replicas > 1 {
		setDefaultProperties(runtime, value, "alluxio.master.journal.type", "EMBEDDED")
	} else {
		setDefaultProperties(runtime, value, "alluxio.master.journal.type", "UFS")
	}
	setDefaultProperties(runtime, value, "alluxio.user.block.master.client.pool.gc.threshold", "10min")
	setDefaultProperties(runtime, value, "alluxio.user.file.master.client.threads", "1024")
	setDefaultProperties(runtime, value, "alluxio.user.block.master.client.threads", "1024")
	setDefaultProperties(runtime, value, "alluxio.user.file.create.ttl.action", "FREE")
	setDefaultProperties(runtime, value, "alluxio.user.file.readtype.default", "CACHE")
	setDefaultProperties(runtime, value, "alluxio.security.stale.channel.purge.interval", "365d")
	setDefaultProperties(runtime, value, "alluxio.user.metadata.cache.enabled", "true")
	setDefaultProperties(runtime, value, "alluxio.user.metadata.cache.expiration.time", "10min")
	// set the default max size of metadata cache
	setDefaultProperties(runtime, value, "alluxio.user.metadata.cache.max.size", "6000000")
	setDefaultProperties(runtime, value, "alluxio.fuse.cached.paths.max", "1000000")
	setDefaultProperties(runtime, value, "alluxio.job.worker.threadpool.size", "32")
	setDefaultProperties(runtime, value, "alluxio.user.worker.list.refresh.interval", "2min")
	setDefaultProperties(runtime, value, "alluxio.user.logging.threshold", "1000ms")
	setDefaultProperties(runtime, value, "alluxio.fuse.logging.threshold", "1000ms")
	setDefaultProperties(runtime, value, "alluxio.worker.block.master.client.pool.size", "1024")
	// Add the optimization of the same blocks concurrent access from https://github.com/Alluxio/alluxio/pull/12453 which is from Microsoft
	// Disable this optimization since it will cause availbilty issue. see https://github.com/Alluxio/alluxio/issues/14909
	// setDefaultProperties(runtime, value, "alluxio.fuse.shared.caching.reader.enabled", "true")
	// job retention time only kicks in when the JobMaster is at capacity in terms of number of jobs stored. see https://github.com/Alluxio/alluxio/pull/12468
	setDefaultProperties(runtime, value, "alluxio.job.master.finished.job.retention.time", "30sec")
	// fixed with https://github.com/Alluxio/alluxio/issues/11437
	setDefaultProperties(runtime, value, "alluxio.underfs.object.store.breadcrumbs.enabled", "false")

	// "alluxio.user.direct.memory.io.enabled" is only safe when the workload is read only and the
	// worker has only one tier and one storage directory in this tier.
	readOnly := false
	runtimeInfo := e.runtimeInfo
	if runtimeInfo != nil {
		accessModes, err := utils.GetAccessModesOfDataset(e.Client, runtimeInfo.GetName(), runtimeInfo.GetNamespace())
		if err != nil {
			e.Log.Info("Error:", "err", err)
		}

		if len(accessModes) > 0 {
			for _, mode := range accessModes {
				if mode == v1.ReadOnlyMany {
					readOnly = true
				}
			}
		}
		tieredstoreInfo := runtimeInfo.GetTieredStoreInfo()
		if readOnly && len(tieredstoreInfo.Levels) == 1 && len(tieredstoreInfo.Levels[0].CachePaths) == 1 {
			setDefaultProperties(runtime, value, "alluxio.user.direct.memory.io.enabled", "true")
		}
	}
}

// optimizeDefaultPropertiesAndFuseForHTTP sets the default value for properties and fuse when the mounts are all HTTP.
func (e *AlluxioEngine) optimizeDefaultPropertiesAndFuseForHTTP(runtime *datav1alpha1.AlluxioRuntime, dataset *datav1alpha1.Dataset, value *Alluxio) {
	var isHTTP = true
	for _, mount := range dataset.Spec.Mounts {
		// the mount is not http
		if !(strings.HasPrefix(mount.MountPoint, common.HttpScheme.String()) || strings.HasPrefix(mount.MountPoint, common.HttpsScheme.String())) {
			isHTTP = false
			break
		}
	}

	if isHTTP {
		setDefaultProperties(runtime, value, "alluxio.user.block.size.bytes.default", "256MB")
		setDefaultProperties(runtime, value, "alluxio.user.streaming.reader.chunk.size.bytes", "256MB")
		setDefaultProperties(runtime, value, "alluxio.user.local.reader.chunk.size.bytes", "256MB")
		setDefaultProperties(runtime, value, "alluxio.worker.network.reader.buffer.size", "256MB")
		setDefaultProperties(runtime, value, "alluxio.user.streaming.data.timeout", "300sec")
		if len(runtime.Spec.Fuse.Args) == 0 {
			value.Fuse.Args[1] = strings.Join([]string{value.Fuse.Args[1], "max_readahead=0"}, ",")
		}
	}
}

func setDefaultProperties(runtime *datav1alpha1.AlluxioRuntime, alluxioValue *Alluxio, key string, value string) {
	if _, found := runtime.Spec.Properties[key]; !found {
		alluxioValue.Properties[key] = value
	}
}

func (e *AlluxioEngine) setPortProperties(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) {
	setDefaultProperties(runtime, value, "alluxio.master.rpc.port", strconv.Itoa(value.Master.Ports.Rpc))
	setDefaultProperties(runtime, value, "alluxio.master.web.port", strconv.Itoa(value.Master.Ports.Web))
	setDefaultProperties(runtime, value, "alluxio.worker.rpc.port", strconv.Itoa(value.Worker.Ports.Rpc))
	setDefaultProperties(runtime, value, "alluxio.worker.web.port", strconv.Itoa(value.Worker.Ports.Web))
	setDefaultProperties(runtime, value, "alluxio.job.master.rpc.port", strconv.Itoa(value.JobMaster.Ports.Rpc))
	setDefaultProperties(runtime, value, "alluxio.job.master.web.port", strconv.Itoa(value.JobMaster.Ports.Web))
	setDefaultProperties(runtime, value, "alluxio.job.worker.rpc.port", strconv.Itoa(value.JobWorker.Ports.Rpc))
	setDefaultProperties(runtime, value, "alluxio.job.worker.web.port", strconv.Itoa(value.JobWorker.Ports.Web))
	setDefaultProperties(runtime, value, "alluxio.job.worker.data.port", strconv.Itoa(value.JobWorker.Ports.Data))
	if runtime.Spec.APIGateway.Enabled {
		setDefaultProperties(runtime, value, "alluxio.proxy.web.port", strconv.Itoa(value.APIGateway.Ports.Rest))
	}

	if value.Master.Ports.Embedded != 0 && value.JobMaster.Ports.Embedded != 0 {
		setDefaultProperties(runtime, value, "alluxio.master.embedded.journal.port", strconv.Itoa(value.Master.Ports.Embedded))
		setDefaultProperties(runtime, value, "alluxio.job.master.embedded.journal.port", strconv.Itoa(value.JobMaster.Ports.Embedded))
	}

	// If use EMBEDDED HA Mode, need set alluxio.master.embedded.journal.addresses
	if value.Master.Replicas > 1 {
		var journalAddresses string
		var journalAddress string
		var i int
		for i = 0; i < int(value.Master.Replicas); i++ {
			if i == int(value.Master.Replicas-1) {
				journalAddress = value.FullnameOverride + "-" + "master-" + strconv.Itoa(i) + ":" + strconv.Itoa(value.Master.Ports.Embedded)
			} else {
				journalAddress = value.FullnameOverride + "-" + "master-" + strconv.Itoa(i) + ":" + strconv.Itoa(value.Master.Ports.Embedded) + ","
			}

			journalAddresses += journalAddress
		}
		setDefaultProperties(runtime, value, "alluxio.master.embedded.journal.addresses", journalAddresses)
	}
}

func (e *AlluxioEngine) optimizeDefaultForMaster(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) {
	if len(runtime.Spec.Master.JvmOptions) > 0 {
		value.Master.JvmOptions = runtime.Spec.Master.JvmOptions
	}

	if len(value.Master.JvmOptions) == 0 {
		value.Master.JvmOptions = []string{
			"-Xmx16G",
			"-XX:+UnlockExperimentalVMOptions",
		}
	}
}

func (e *AlluxioEngine) optimizeDefaultForWorker(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) {
	if len(runtime.Spec.Worker.JvmOptions) > 0 {
		value.Worker.JvmOptions = runtime.Spec.Worker.JvmOptions
	}
	if len(value.Worker.JvmOptions) == 0 {
		value.Worker.JvmOptions = []string{
			"-Xmx12G",
			"-XX:+UnlockExperimentalVMOptions",
			"-XX:MaxDirectMemorySize=32g",
		}
	}
}

func (e *AlluxioEngine) optimizeDefaultFuse(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio, isNewFuseArgVersion bool) {

	if len(runtime.Spec.Fuse.JvmOptions) > 0 {
		value.Fuse.JvmOptions = runtime.Spec.Fuse.JvmOptions
	}

	if len(value.Fuse.JvmOptions) == 0 {
		value.Fuse.JvmOptions = []string{
			"-Xmx16G",
			"-Xms16G",
			"-XX:+UseG1GC",
			"-XX:MaxDirectMemorySize=32g",
			"-XX:+UnlockExperimentalVMOptions",
		}
	}

	readOnly := false
	runtimeInfo := e.runtimeInfo
	if runtimeInfo != nil {
		accessModes, err := utils.GetAccessModesOfDataset(e.Client, runtimeInfo.GetName(), runtimeInfo.GetNamespace())
		if err != nil {
			e.Log.Info("Error:", "err", err)
		}

		if len(accessModes) > 0 {
			for _, mode := range accessModes {
				if mode == v1.ReadOnlyMany {
					readOnly = true
				}
			}
		}
	}

	if len(runtime.Spec.Fuse.Args) > 0 {
		value.Fuse.Args = runtime.Spec.Fuse.Args
	} else {
		if readOnly {
			value.Fuse.Args = []string{"fuse", "--fuse-opts=kernel_cache,ro,max_read=131072,attr_timeout=7200,entry_timeout=7200,nonempty"}
		} else {
			value.Fuse.Args = []string{"fuse", "--fuse-opts=kernel_cache,rw,max_read=131072"}
		}
	}

	if isNewFuseArgVersion {
		value.Fuse.Args = append(value.Fuse.Args, value.Fuse.MountPath, "/")
	}

}
