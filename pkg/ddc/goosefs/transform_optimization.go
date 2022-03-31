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

import (
	"strconv"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"

	v1 "k8s.io/api/core/v1"
)

// transform dataset which has ufsPaths and ufsVolumes
func (e *GooseFSEngine) optimizeDefaultProperties(runtime *datav1alpha1.GooseFSRuntime, value *GooseFS) {
	if len(value.Properties) == 0 {
		if len(runtime.Spec.Properties) > 0 {
			value.Properties = runtime.Spec.Properties
		} else {
			value.Properties = map[string]string{}
		}
	}
	setDefaultProperties(runtime, value, "goosefs.fuse.jnifuse.enabled", "true")
	setDefaultProperties(runtime, value, "goosefs.master.metastore", "ROCKS")
	setDefaultProperties(runtime, value, "goosefs.web.ui.enabled", "false")
	setDefaultProperties(runtime, value, "goosefs.user.update.file.accesstime.disabled", "true")
	setDefaultProperties(runtime, value, "goosefs.user.client.cache.enabled", "false")
	setDefaultProperties(runtime, value, "goosefs.master.metastore.inode.cache.max.size", "10000000")
	setDefaultProperties(runtime, value, "goosefs.master.journal.log.size.bytes.max", "500MB")
	setDefaultProperties(runtime, value, "goosefs.master.metadata.sync.concurrency.level", "128")
	setDefaultProperties(runtime, value, "goosefs.master.metadata.sync.executor.pool.size", "128")
	setDefaultProperties(runtime, value, "goosefs.master.metadata.sync.ufs.prefetch.pool.size", "128")
	setDefaultProperties(runtime, value, "goosefs.user.block.worker.client.pool.min", "512")
	setDefaultProperties(runtime, value, "goosefs.fuse.debug.enabled", "false")
	setDefaultProperties(runtime, value, "goosefs.web.ui.enabled", "false")
	setDefaultProperties(runtime, value, "goosefs.user.file.writetype.default", "MUST_CACHE")
	setDefaultProperties(runtime, value, "goosefs.user.ufs.block.read.location.policy", "com.qcloud.cos.goosefs.client.block.policy.LocalFirstAvoidEvictionPolicy")
	setDefaultProperties(runtime, value, "goosefs.user.block.write.location.policy.class", "com.qcloud.cos.goosefs.client.block.policy.LocalFirstAvoidEvictionPolicy")
	setDefaultProperties(runtime, value, "goosefs.worker.allocator.class", "com.qcloud.cos.goosefs.worker.block.allocator.MaxFreeAllocator")
	setDefaultProperties(runtime, value, "goosefs.user.block.size.bytes.default", "16MB")
	setDefaultProperties(runtime, value, "goosefs.user.streaming.reader.chunk.size.bytes", "32MB")
	setDefaultProperties(runtime, value, "goosefs.user.local.reader.chunk.size.bytes", "32MB")
	setDefaultProperties(runtime, value, "goosefs.worker.network.reader.buffer.size", "32MB")
	setDefaultProperties(runtime, value, "goosefs.worker.file.buffer.size", "1MB")
	// Enable metrics as default for better monitoring result, if you have performance concern, feel free to turn it off
	setDefaultProperties(runtime, value, "goosefs.user.metrics.collection.enabled", "true")
	setDefaultProperties(runtime, value, "goosefs.master.rpc.executor.max.pool.size", "1024")
	setDefaultProperties(runtime, value, "goosefs.master.rpc.executor.core.pool.size", "128")
	// setDefaultProperties(runtime, value, "goosefs.master.mount.table.root.readonly", "true")
	setDefaultProperties(runtime, value, "goosefs.user.update.file.accesstime.disabled", "true")
	setDefaultProperties(runtime, value, "goosefs.user.file.passive.cache.enabled", "false")
	setDefaultProperties(runtime, value, "goosefs.user.block.avoid.eviction.policy.reserved.size.bytes", "2GB")
	setDefaultProperties(runtime, value, "goosefs.master.journal.folder", "/journal")
	setDefaultProperties(runtime, value, "goosefs.user.block.master.client.pool.gc.threshold", "2day")
	setDefaultProperties(runtime, value, "goosefs.user.file.master.client.threads", "1024")
	setDefaultProperties(runtime, value, "goosefs.user.block.master.client.threads", "1024")
	setDefaultProperties(runtime, value, "goosefs.user.file.create.ttl.action", "FREE")
	setDefaultProperties(runtime, value, "goosefs.user.file.readtype.default", "CACHE")
	setDefaultProperties(runtime, value, "goosefs.security.stale.channel.purge.interval", "365d")
	setDefaultProperties(runtime, value, "goosefs.user.metadata.cache.enabled", "true")
	setDefaultProperties(runtime, value, "goosefs.user.metadata.cache.expiration.time", "2day")
	// set the default max size of metadata cache
	setDefaultProperties(runtime, value, "goosefs.user.metadata.cache.max.size", "6000000")
	setDefaultProperties(runtime, value, "goosefs.fuse.cached.paths.max", "1000000")
	setDefaultProperties(runtime, value, "goosefs.job.worker.threadpool.size", "164")
	setDefaultProperties(runtime, value, "goosefs.user.worker.list.refresh.interval", "2min")
	setDefaultProperties(runtime, value, "goosefs.user.logging.threshold", "1000ms")
	setDefaultProperties(runtime, value, "goosefs.fuse.logging.threshold", "1000ms")
	setDefaultProperties(runtime, value, "goosefs.worker.block.master.client.pool.size", "1024")
	// Disable this optimization since it will cause availbilty issue. see https://github.com/Alluxio/alluxio/issues/14909
	// setDefaultProperties(runtime, value, "goosefs.fuse.shared.caching.reader.enabled", "true")
	setDefaultProperties(runtime, value, "goosefs.job.master.finished.job.retention.time", "30sec")
	setDefaultProperties(runtime, value, "goosefs.underfs.object.store.breadcrumbs.enabled", "false")

	if value.Master.Replicas > 1 {
		setDefaultProperties(runtime, value, "goosefs.master.journal.type", "EMBEDDED")
	} else {
		setDefaultProperties(runtime, value, "goosefs.master.journal.type", "UFS")
	}

	// "goosefs.user.direct.memory.io.enabled" is only safe when the workload is read only and the
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
			setDefaultProperties(runtime, value, "goosefs.user.direct.memory.io.enabled", "true")
		}
	}
}

// optimizeDefaultPropertiesAndFuseForHTTP sets the default value for properties and fuse when the mounts are all HTTP.
func (e *GooseFSEngine) optimizeDefaultPropertiesAndFuseForHTTP(runtime *datav1alpha1.GooseFSRuntime, dataset *datav1alpha1.Dataset, value *GooseFS) {
	var isHTTP = true
	for _, mount := range dataset.Spec.Mounts {
		// the mount is not http
		if !(strings.HasPrefix(mount.MountPoint, common.HttpScheme.String()) || strings.HasPrefix(mount.MountPoint, common.HttpsScheme.String())) {
			isHTTP = false
			break
		}
	}

	if isHTTP {
		setDefaultProperties(runtime, value, "goosefs.user.block.size.bytes.default", "256MB")
		setDefaultProperties(runtime, value, "goosefs.user.streaming.reader.chunk.size.bytes", "256MB")
		setDefaultProperties(runtime, value, "goosefs.user.local.reader.chunk.size.bytes", "256MB")
		setDefaultProperties(runtime, value, "goosefs.worker.network.reader.buffer.size", "256MB")
		setDefaultProperties(runtime, value, "goosefs.user.streaming.data.timeout", "300sec")
		if len(runtime.Spec.Fuse.Args) == 0 {
			value.Fuse.Args[1] = strings.Join([]string{value.Fuse.Args[1], "max_readahead=0"}, ",")
		}
	}
}

func setDefaultProperties(runtime *datav1alpha1.GooseFSRuntime, goosefsValue *GooseFS, key string, value string) {
	if _, found := runtime.Spec.Properties[key]; !found {
		goosefsValue.Properties[key] = value
	}
}

func (e *GooseFSEngine) setPortProperties(runtime *datav1alpha1.GooseFSRuntime, value *GooseFS) {
	setDefaultProperties(runtime, value, "goosefs.master.rpc.port", strconv.Itoa(value.Master.Ports.Rpc))
	setDefaultProperties(runtime, value, "goosefs.master.web.port", strconv.Itoa(value.Master.Ports.Web))
	setDefaultProperties(runtime, value, "goosefs.worker.rpc.port", strconv.Itoa(value.Worker.Ports.Rpc))
	setDefaultProperties(runtime, value, "goosefs.worker.web.port", strconv.Itoa(value.Worker.Ports.Web))
	setDefaultProperties(runtime, value, "goosefs.job.master.rpc.port", strconv.Itoa(value.JobMaster.Ports.Rpc))
	setDefaultProperties(runtime, value, "goosefs.job.master.web.port", strconv.Itoa(value.JobMaster.Ports.Web))
	setDefaultProperties(runtime, value, "goosefs.job.worker.rpc.port", strconv.Itoa(value.JobWorker.Ports.Rpc))
	setDefaultProperties(runtime, value, "goosefs.job.worker.web.port", strconv.Itoa(value.JobWorker.Ports.Web))
	setDefaultProperties(runtime, value, "goosefs.job.worker.data.port", strconv.Itoa(value.JobWorker.Ports.Data))
	if runtime.Spec.APIGateway.Enabled {
		setDefaultProperties(runtime, value, "goosefs.proxy.web.port", strconv.Itoa(value.APIGateway.Ports.Rest))
	}

	if value.Master.Ports.Embedded != 0 && value.JobMaster.Ports.Embedded != 0 {
		setDefaultProperties(runtime, value, "goosefs.master.embedded.journal.port", strconv.Itoa(value.Master.Ports.Embedded))
		setDefaultProperties(runtime, value, "goosefs.job.master.embedded.journal.port", strconv.Itoa(value.JobMaster.Ports.Embedded))
	}

	// If use EMBEDDED HA Mode, need set goosefs.master.embedded.journal.addresses
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
		setDefaultProperties(runtime, value, "goosefs.master.embedded.journal.addresses", journalAddresses)
	}
}

func (e *GooseFSEngine) optimizeDefaultForMaster(runtime *datav1alpha1.GooseFSRuntime, value *GooseFS) {
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

func (e *GooseFSEngine) optimizeDefaultForWorker(runtime *datav1alpha1.GooseFSRuntime, value *GooseFS) {
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

func (e *GooseFSEngine) optimizeDefaultFuse(runtime *datav1alpha1.GooseFSRuntime, value *GooseFS) {

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
			// value.Fuse.Args = []string{"fuse", "--fuse-opts=kernel_cache,ro,max_read=131072,attr_timeout=7200,entry_timeout=7200,nonempty"}
			value.Fuse.Args = []string{"fuse", "--fuse-opts=ro,direct_io"}
		} else {
			// value.Fuse.Args = []string{"fuse", "--fuse-opts=kernel_cache,rw,max_read=131072,attr_timeout=7200,entry_timeout=7200,nonempty"}
			value.Fuse.Args = []string{"fuse", "--fuse-opts=rw,direct_io"}
		}

	}

}
