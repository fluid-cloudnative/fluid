/*
Copyright 2023 The Fluid Author.

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
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var valuesConfigMapData = `
fullnameOverride: hbase
image: ccr.ccs.tencentyun.com/goosefs/goosefs
imageTag: v1.0.1
imagePullPolicy: IfNotPresent
user: 0
group: 0
fsGroup: 0
properties:
  goosefs.fuse.cached.paths.max: "1000000"
  goosefs.fuse.debug.enabled: "false"
  goosefs.fuse.jnifuse.enabled: "true"
  goosefs.fuse.logging.threshold: 1000ms
  goosefs.fuse.user.group.translation.enabled: "true"
  goosefs.job.master.finished.job.retention.time: 30sec
  goosefs.job.master.rpc.port: "20004"
  goosefs.job.master.web.port: "20005"
  goosefs.job.worker.data.port: "20008"
  goosefs.job.worker.rpc.port: "20006"
  goosefs.job.worker.threadpool.size: "164"
  goosefs.job.worker.web.port: "20007"
  goosefs.master.journal.folder: /journal
  goosefs.master.journal.log.size.bytes.max: 500MB
  goosefs.master.journal.type: UFS
  goosefs.master.metadata.sync.concurrency.level: "128"
  goosefs.master.metadata.sync.executor.pool.size: "128"
  goosefs.master.metadata.sync.ufs.prefetch.pool.size: "128"
  goosefs.master.metastore: ROCKS
  goosefs.master.metastore.inode.cache.max.size: "10000000"
  goosefs.master.mount.table.root.ufs: /underFSStorage
  goosefs.master.rpc.executor.core.pool.size: "128"
  goosefs.master.rpc.executor.max.pool.size: "1024"
  goosefs.master.rpc.port: "20000"
  goosefs.master.security.impersonation.root.groups: '*'
  goosefs.master.security.impersonation.root.users: '*'
  goosefs.master.web.port: "20001"
  goosefs.security.authorization.permission.enabled: "false"
  goosefs.security.stale.channel.purge.interval: 365d
  goosefs.underfs.object.store.breadcrumbs.enabled: "false"
  goosefs.user.block.avoid.eviction.policy.reserved.size.bytes: 2GB
  goosefs.user.block.master.client.pool.gc.threshold: 2day
  goosefs.user.block.master.client.threads: "1024"
  goosefs.user.block.size.bytes.default: 256MB
  goosefs.user.block.worker.client.pool.min: "512"
  goosefs.user.block.write.location.policy.class: com.qcloud.cos.goosefs.client.block.policy.LocalFirstAvoidEvictionPolicy
  goosefs.user.client.cache.enabled: "false"
  goosefs.user.direct.memory.io.enabled: "true"
  goosefs.user.file.create.ttl.action: FREE
  goosefs.user.file.master.client.threads: "1024"
  goosefs.user.file.passive.cache.enabled: "false"
  goosefs.user.file.readtype.default: CACHE
  goosefs.user.file.replication.max: "1"
  goosefs.user.file.writetype.default: MUST_CACHE
  goosefs.user.local.reader.chunk.size.bytes: 256MB
  goosefs.user.logging.threshold: 1000ms
  goosefs.user.metadata.cache.enabled: "true"
  goosefs.user.metadata.cache.expiration.time: 2day
  goosefs.user.metadata.cache.max.size: "6000000"
  goosefs.user.metrics.collection.enabled: "true"
  goosefs.user.streaming.data.timeout: 300sec
  goosefs.user.streaming.reader.chunk.size.bytes: 256MB
  goosefs.user.ufs.block.read.location.policy: com.qcloud.cos.goosefs.client.block.policy.LocalFirstPolicy
  goosefs.user.update.file.accesstime.disabled: "true"
  goosefs.user.worker.list.refresh.interval: 2min
  goosefs.web.ui.enabled: "false"
  goosefs.worker.allocator.class: goosefs.worker.block.allocator.MaxFreeAllocator
  goosefs.worker.block.master.client.pool.size: "1024"
  goosefs.worker.network.reader.buffer.size: 256MB
  goosefs.worker.rpc.port: "20002"
  goosefs.worker.web.port: "20003"
fuse:
  image: ccr.ccs.tencentyun.com/goosefs/goosefs-fuse
  nodeSelector:
    fluid.io/s-default-hbase: "true"
  imageTag: v1.0.1
  imagePullPolicy: IfNotPresent
  env:
    MOUNT_POINT: /runtime-mnt/goosefs/default/hbase/goosefs-fuse
  jvmOptions:
  - -Xmx16G
  - -Xms16G
  - -XX:+UseG1GC
  - -XX:MaxDirectMemorySize=32g
  - -XX:+UnlockExperimentalVMOptions
  mountPath: /runtime-mnt/goosefs/default/hbase/goosefs-fuse
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
monitoring: goosefs_runtime_metrics
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

func TestGetReservedPorts(t *testing.T) {
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase-goosefs-values",
			Namespace: "fluid",
		},
		Data: map[string]string{
			"data": valuesConfigMapData,
		},
	}
	dataSets := []*v1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Status: v1alpha1.DatasetStatus{
				Runtimes: []v1alpha1.Runtime{
					{
						Name:      "hbase",
						Namespace: "fluid",
						Type:      "goosefs",
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "no-runtime",
				Namespace: "fluid",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "runtime-type",
				Namespace: "fluid",
			},
			Status: v1alpha1.DatasetStatus{
				Runtimes: []v1alpha1.Runtime{
					{
						Type: "not-goosefs",
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "no-map",
				Namespace: "fluid",
			},
			Status: v1alpha1.DatasetStatus{
				Runtimes: []v1alpha1.Runtime{
					{
						Type: "goosefs",
					},
				},
			},
		},
	}
	runtimeObjs := []runtime.Object{}
	runtimeObjs = append(runtimeObjs, configMap)
	for _, dataSet := range dataSets {
		runtimeObjs = append(runtimeObjs, dataSet.DeepCopy())
	}
	fakeClient := fake.NewFakeClientWithScheme(testScheme, runtimeObjs...)
	wantPorts := []int{20000, 20001, 20002, 20003, 20004, 20005, 20006, 20007, 20008}
	ports, err := GetReservedPorts(fakeClient)
	if err != nil {
		t.Errorf("GetReservedPorts failed.")
	}
	if !reflect.DeepEqual(ports, wantPorts) {
		t.Errorf("gotPorts = %v, want %v", ports, wantPorts)
	}

}
