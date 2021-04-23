package alluxio

import (
	v1 "k8s.io/api/core/v1"
	"reflect"
	"testing"
)

var valuesConfigMapData = `
fullnameOverride: hbase
image: registry.aliyuncs.com/alluxio/alluxio
imageTag: release-2.5.0-2-SNAPSHOT-52ad95c
imagePullPolicy: IfNotPresent
user: 0
group: 0
fsGroup: 0
properties:
  alluxio.fuse.cached.paths.max: "1000000"
  alluxio.fuse.debug.enabled: "false"
  alluxio.fuse.jnifuse.enabled: "true"
  alluxio.fuse.logging.threshold: 1000ms
  alluxio.fuse.shared.caching.reader.enabled: "true"
  alluxio.fuse.user.group.translation.enabled: "true"
  alluxio.job.master.finished.job.retention.time: 30sec
  alluxio.job.master.rpc.port: "20004"
  alluxio.job.master.web.port: "20005"
  alluxio.job.worker.data.port: "20008"
  alluxio.job.worker.rpc.port: "20006"
  alluxio.job.worker.threadpool.size: "164"
  alluxio.job.worker.web.port: "20007"
  alluxio.master.journal.folder: /journal
  alluxio.master.journal.log.size.bytes.max: 500MB
  alluxio.master.journal.type: UFS
  alluxio.master.metadata.sync.concurrency.level: "128"
  alluxio.master.metadata.sync.executor.pool.size: "128"
  alluxio.master.metadata.sync.ufs.prefetch.pool.size: "128"
  alluxio.master.metastore: ROCKS
  alluxio.master.metastore.inode.cache.max.size: "10000000"
  alluxio.master.mount.table.root.ufs: /underFSStorage
  alluxio.master.rpc.executor.core.pool.size: "128"
  alluxio.master.rpc.executor.max.pool.size: "1024"
  alluxio.master.rpc.port: "20000"
  alluxio.master.security.impersonation.root.groups: '*'
  alluxio.master.security.impersonation.root.users: '*'
  alluxio.master.web.port: "20001"
  alluxio.security.authorization.permission.enabled: "false"
  alluxio.security.stale.channel.purge.interval: 365d
  alluxio.underfs.object.store.breadcrumbs.enabled: "false"
  alluxio.user.block.avoid.eviction.policy.reserved.size.bytes: 2GB
  alluxio.user.block.master.client.pool.gc.threshold: 2day
  alluxio.user.block.master.client.threads: "1024"
  alluxio.user.block.size.bytes.default: 256MB
  alluxio.user.block.worker.client.pool.min: "512"
  alluxio.user.block.write.location.policy.class: alluxio.client.block.policy.LocalFirstAvoidEvictionPolicy
  alluxio.user.client.cache.enabled: "false"
  alluxio.user.direct.memory.io.enabled: "true"
  alluxio.user.file.create.ttl.action: FREE
  alluxio.user.file.master.client.threads: "1024"
  alluxio.user.file.passive.cache.enabled: "false"
  alluxio.user.file.readtype.default: CACHE
  alluxio.user.file.replication.max: "1"
  alluxio.user.file.writetype.default: MUST_CACHE
  alluxio.user.local.reader.chunk.size.bytes: 256MB
  alluxio.user.logging.threshold: 1000ms
  alluxio.user.metadata.cache.enabled: "true"
  alluxio.user.metadata.cache.expiration.time: 2day
  alluxio.user.metadata.cache.max.size: "6000000"
  alluxio.user.metrics.collection.enabled: "true"
  alluxio.user.streaming.data.timeout: 300sec
  alluxio.user.streaming.reader.chunk.size.bytes: 256MB
  alluxio.user.ufs.block.read.location.policy: alluxio.client.block.policy.LocalFirstPolicy
  alluxio.user.update.file.accesstime.disabled: "true"
  alluxio.user.worker.list.refresh.interval: 2min
  alluxio.web.ui.enabled: "false"
  alluxio.worker.allocator.class: alluxio.worker.block.allocator.MaxFreeAllocator
  alluxio.worker.block.master.client.pool.size: "1024"
  alluxio.worker.network.reader.buffer.size: 256MB
  alluxio.worker.rpc.port: "20002"
  alluxio.worker.web.port: "20003"
fuse:
  image: registry.aliyuncs.com/alluxio/alluxio-fuse
  nodeSelector:
    fluid.io/s-default-hbase: "true"
  imageTag: release-2.5.0-2-SNAPSHOT-52ad95c
  imagePullPolicy: IfNotPresent
  env:
    MOUNT_POINT: /runtime-mnt/alluxio/default/hbase/alluxio-fuse
  jvmOptions:
  - -Xmx16G
  - -Xms16G
  - -XX:+UseG1GC
  - -XX:MaxDirectMemorySize=32g
  - -XX:+UnlockExperimentalVMOptions
  mountPath: /runtime-mnt/alluxio/default/hbase/alluxio-fuse
  args:
  - fuse
  - --fuse-opts=kernel_cache,ro,max_read=131072,attr_timeout=7200,entry_timeout=7200,nonempty,max_readahead=0,allow_other
  hostNetwork: true
  enabled: true
tieredstore:
  levels:
  - alias: MEM
    level: 0
    mediumtype: MEM
    type: hostPath
    path: /dev/shm/default/hbase
    quota: 2GB
    high: "0.95"
    low: "0.7"
journal:
  volumeType: emptyDir
  size: 30Gi
shortCircuit:
  enable: true
  policy: local
  volumeType: emptyDir
monitoring: alluxio_runtime_metrics
`

func Test_parsePortsFromConfigMap(t *testing.T) {
	type args struct {
		configMap *v1.ConfigMap
	}
	tests := []struct {
		name      string
		args      args
		wantPorts []int
		wantErr   bool
	}{
		{
			name: "parsePortsFromConfigMap",
			args: args{configMap: &v1.ConfigMap{
				Data: map[string]string{
					"data": valuesConfigMapData,
				},
			}},
			wantPorts: []int{20000, 20001, 20002, 20003, 20004, 20005, 20006, 20007, 20008},
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPorts, err := parsePortsFromConfigMap(tt.args.configMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePortsFromConfigMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPorts, tt.wantPorts) {
				t.Errorf("parsePortsFromConfigMap() gotPorts = %v, want %v", gotPorts, tt.wantPorts)
			}
		})
	}
}
