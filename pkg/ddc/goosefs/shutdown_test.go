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

import (
	"reflect"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/net"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	testScheme        *runtime.Scheme
	mockConfigMapData = `----
fullnameOverride: mnist
image: ccr.ccs.tencentyun.com/qcloud/goosefs
imageTag: v1.2.0
imagePullPolicy: IfNotPresent
user: 0
group: 0
fsGroup: 0
properties:
  goosefs.fuse.cached.paths.max: "1000000"
  goosefs.fuse.debug.enabled: "true"
  goosefs.fuse.jnifuse.enabled: "true"
  goosefs.fuse.logging.threshold: 1000ms
  goosefs.fuse.user.group.translation.enabled: "true"
  goosefs.job.master.finished.job.retention.time: 30sec
  goosefs.job.master.rpc.port: "28362"
  goosefs.job.master.web.port: "31380"
  goosefs.job.worker.data.port: "30918"
  goosefs.job.worker.rpc.port: "29476"
  goosefs.job.worker.threadpool.size: "164"
  goosefs.job.worker.web.port: "27403"
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
  goosefs.master.rpc.port: "30399"
  goosefs.master.security.impersonation.root.groups: '*'
  goosefs.master.security.impersonation.root.users: '*'
  goosefs.master.web.port: "31203"
  goosefs.security.authorization.permission.enabled: "false"
  goosefs.security.stale.channel.purge.interval: 365d
  goosefs.underfs.object.store.breadcrumbs.enabled: "false"
  goosefs.user.block.avoid.eviction.policy.reserved.size.bytes: 2GB
  goosefs.user.block.master.client.pool.gc.threshold: 2day
  goosefs.user.block.master.client.threads: "1024"
  goosefs.user.block.size.bytes.default: 16MB
  goosefs.user.block.worker.client.pool.min: "512"
  goosefs.user.block.write.location.policy.class: com.qcloud.cos.goosefs.client.block.policy.LocalFirstAvoidEvictionPolicy
  goosefs.user.client.cache.enabled: "false"
  goosefs.user.file.create.ttl.action: FREE
  goosefs.user.file.master.client.threads: "1024"
  goosefs.user.file.passive.cache.enabled: "false"
  goosefs.user.file.readtype.default: CACHE
  goosefs.user.file.replication.max: "1"
  goosefs.user.file.writetype.default: CACHE_THROUGH
  goosefs.user.local.reader.chunk.size.bytes: 32MB
  goosefs.user.logging.threshold: 1000ms
  goosefs.user.metadata.cache.enabled: "true"
  goosefs.user.metadata.cache.expiration.time: 2day
  goosefs.user.metadata.cache.max.size: "6000000"
  goosefs.user.metrics.collection.enabled: "true"
  goosefs.user.streaming.reader.chunk.size.bytes: 32MB
  goosefs.user.ufs.block.read.location.policy: com.qcloud.cos.goosefs.client.block.policy.LocalFirstAvoidEvictionPolicy
  goosefs.user.update.file.accesstime.disabled: "true"
  goosefs.user.worker.list.refresh.interval: 2min
  goosefs.web.ui.enabled: "false"
  goosefs.worker.allocator.class: com.qcloud.cos.goosefs.worker.block.allocator.MaxFreeAllocator
  goosefs.worker.block.master.client.pool.size: "1024"
  goosefs.worker.file.buffer.size: 1MB
  goosefs.worker.network.reader.buffer.size: 32MB
  goosefs.worker.rpc.port: "31285"
  goosefs.worker.web.port: "31674"
  log4j.logger.alluxio.fuse: DEBUG
  log4j.logger.com.qcloud.cos.goosefs.fuse: DEBUG
master:
  jvmOptions:
  - -Xmx16G
  - -XX:+UnlockExperimentalVMOptions
  env:
    GOOSEFS_WORKER_TIEREDSTORE_LEVEL0_DIRS_PATH: /dev/shm/yijiupi/mnist
  affinity:
    nodeAffinity: null
  replicaCount: 1
  hostNetwork: true
  ports:
    rpc: 30399
    web: 31203
  backupPath: /tmp/goosefs-backup/yijiupi/mnist
jobMaster:
  ports:
    rpc: 28362
    web: 31380
worker:
  jvmOptions:
  - -Xmx12G
  - -XX:+UnlockExperimentalVMOptions
  - -XX:MaxDirectMemorySize=32g
  env:
    GOOSEFS_WORKER_TIEREDSTORE_LEVEL0_DIRS_PATH: /dev/shm/yijiupi/mnist
  hostNetwork: true
  ports:
    rpc: 31285
    web: 31674
jobWorker:
  ports:
    rpc: 29476
    web: 27403
    data: 30918
fuse:
  image: ccr.ccs.tencentyun.com/qcloud/goosefs-fuse
  nodeSelector:
    fluid.io/f-yijiupi-mnist: "true"
  imageTag: v1.2.0
  env:
    MOUNT_POINT: /runtime-mnt/goosefs/yijiupi/mnist/goosefs-fuse
  jvmOptions:
  - -Xmx16G
  - -Xms16G
  - -XX:+UseG1GC
  - -XX:MaxDirectMemorySize=32g
  - -XX:+UnlockExperimentalVMOptions
  mountPath: /runtime-mnt/goosefs/yijiupi/mnist/goosefs-fuse
  args:
  - fuse
  - --fuse-opts=rw,allow_other
  hostNetwork: true
  enabled: true
  criticalPod: true
tieredstore:
  levels:
  - alias: MEM
    level: 0
    mediumtype: MEM
    type: hostPath
    path: /dev/shm/yijiupi/mnist
    quota: 1953125KB
    high: "0.8"
    low: "0.7"
journal:
  volumeType: emptyDir
  size: 30Gi
shortCircuit:
  enable: true
  policy: local
  volumeType: emptyDir
initUsers:
  image: fluidcloudnative/init-users
  imageTag: v0.7.0-1cf2443
  imagePullPolicy: IfNotPresent
  envUsers: ""
  dir: ""
  envTieredPaths: ""
monitoring: goosefs_runtime_metrics
placement: Exclusive`
)

func init() {
	testScheme = runtime.NewScheme()
	_ = corev1.AddToScheme(testScheme)
	_ = datav1alpha1.AddToScheme(testScheme)
	_ = appsv1.AddToScheme(testScheme)
}

func TestDestroyWorker(t *testing.T) {
	// runtimeInfoSpark tests destroy Worker in exclusive mode.
	runtimeInfoSpark, err := base.BuildRuntimeInfo("spark", "fluid", "goosefs", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoSpark.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
	})

	// runtimeInfoSpark tests destroy Worker in shareMode mode.
	runtimeInfoHadoop, err := base.BuildRuntimeInfo("hadoop", "fluid", "goosefs", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoHadoop.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ShareMode},
	})
	nodeSelector := map[string]string{
		"node-select": "true",
	}
	runtimeInfoHadoop.SetupFuseDeployMode(true, nodeSelector)

	var nodeInputs = []*corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-spark",
				Labels: map[string]string{
					"fluid.io/dataset-num":               "1",
					"fluid.io/s-goosefs-fluid-spark":     "true",
					"fluid.io/s-fluid-spark":             "true",
					"fluid.io/s-h-goosefs-d-fluid-spark": "5B",
					"fluid.io/s-h-goosefs-m-fluid-spark": "1B",
					"fluid.io/s-h-goosefs-t-fluid-spark": "6B",
					"fluid_exclusive":                    "fluid_spark",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-share",
				Labels: map[string]string{
					"fluid.io/dataset-num":                "2",
					"fluid.io/s-goosefs-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":             "true",
					"fluid.io/s-h-goosefs-d-fluid-hadoop": "5B",
					"fluid.io/s-h-goosefs-m-fluid-hadoop": "1B",
					"fluid.io/s-h-goosefs-t-fluid-hadoop": "6B",
					"fluid.io/s-goosefs-fluid-hbase":      "true",
					"fluid.io/s-fluid-hbase":              "true",
					"fluid.io/s-h-goosefs-d-fluid-hbase":  "5B",
					"fluid.io/s-h-goosefs-m-fluid-hbase":  "1B",
					"fluid.io/s-h-goosefs-t-fluid-hbase":  "6B",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-hadoop",
				Labels: map[string]string{
					"fluid.io/dataset-num":                "1",
					"fluid.io/s-goosefs-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":             "true",
					"fluid.io/s-h-goosefs-d-fluid-hadoop": "5B",
					"fluid.io/s-h-goosefs-m-fluid-hadoop": "1B",
					"fluid.io/s-h-goosefs-t-fluid-hadoop": "6B",
					"node-select":                         "true",
				},
			},
		},
	}

	testNodes := []runtime.Object{}
	for _, nodeInput := range nodeInputs {
		testNodes = append(testNodes, nodeInput.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testNodes...)

	var testCase = []struct {
		expectedWorkers  int32
		runtimeInfo      base.RuntimeInfoInterface
		wantedNodeNumber int32
		wantedNodeLabels map[string]map[string]string
	}{
		{
			expectedWorkers:  -1,
			runtimeInfo:      runtimeInfoSpark,
			wantedNodeNumber: 0,
			wantedNodeLabels: map[string]map[string]string{
				"test-node-spark": {},
				"test-node-share": {
					"fluid.io/dataset-num":                "2",
					"fluid.io/s-goosefs-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":             "true",
					"fluid.io/s-h-goosefs-d-fluid-hadoop": "5B",
					"fluid.io/s-h-goosefs-m-fluid-hadoop": "1B",
					"fluid.io/s-h-goosefs-t-fluid-hadoop": "6B",
					"fluid.io/s-goosefs-fluid-hbase":      "true",
					"fluid.io/s-fluid-hbase":              "true",
					"fluid.io/s-h-goosefs-d-fluid-hbase":  "5B",
					"fluid.io/s-h-goosefs-m-fluid-hbase":  "1B",
					"fluid.io/s-h-goosefs-t-fluid-hbase":  "6B",
				},
				"test-node-hadoop": {
					"fluid.io/dataset-num":                "1",
					"fluid.io/s-goosefs-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":             "true",
					"fluid.io/s-h-goosefs-d-fluid-hadoop": "5B",
					"fluid.io/s-h-goosefs-m-fluid-hadoop": "1B",
					"fluid.io/s-h-goosefs-t-fluid-hadoop": "6B",
					"node-select":                         "true",
				},
			},
		},
		{
			expectedWorkers:  -1,
			runtimeInfo:      runtimeInfoHadoop,
			wantedNodeNumber: 0,
			wantedNodeLabels: map[string]map[string]string{
				"test-node-spark": {},
				"test-node-share": {
					"fluid.io/dataset-num":               "1",
					"fluid.io/s-goosefs-fluid-hbase":     "true",
					"fluid.io/s-fluid-hbase":             "true",
					"fluid.io/s-h-goosefs-d-fluid-hbase": "5B",
					"fluid.io/s-h-goosefs-m-fluid-hbase": "1B",
					"fluid.io/s-h-goosefs-t-fluid-hbase": "6B",
				},
				"test-node-hadoop": {
					"node-select": "true",
				},
			},
		},
	}
	for _, test := range testCase {
		engine := &GooseFSEngine{Log: fake.NullLogger(), runtimeInfo: test.runtimeInfo}
		engine.Client = client
		engine.name = test.runtimeInfo.GetName()
		engine.namespace = test.runtimeInfo.GetNamespace()
		if err != nil {
			t.Errorf("fail to exec the function with the error %v", err)
		}
		currentWorkers, err := engine.destroyWorkers(test.expectedWorkers)
		if err != nil {
			t.Errorf("fail to exec the function with the error %v", err)
		}
		if currentWorkers != test.wantedNodeNumber {
			t.Errorf("shutdown the worker with the wrong number of the workers")
		}
		for _, node := range nodeInputs {
			newNode, err := kubeclient.GetNode(client, node.Name)
			if err != nil {
				t.Errorf("fail to get the node with the error %v", err)
			}

			if len(newNode.Labels) != len(test.wantedNodeLabels[node.Name]) {
				t.Errorf("fail to decrease the labels")
			}
			if len(newNode.Labels) != 0 && !reflect.DeepEqual(newNode.Labels, test.wantedNodeLabels[node.Name]) {
				t.Errorf("fail to decrease the labels")
			}
		}

	}
}

func TestGooseFSEngineCleanAll(t *testing.T) {
	type fields struct {
		name        string
		namespace   string
		cm          *corev1.ConfigMap
		runtimeType string
		log         logr.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "spark",
			fields: fields{
				name:        "spark",
				namespace:   "fluid",
				runtimeType: "goosefs",
				cm: &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-goosefs-values",
						Namespace: "fluid",
					},
					Data: map[string]string{"data": mockConfigMapData},
				},
				log: fake.NullLogger(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testObjs := []runtime.Object{}
			testObjs = append(testObjs, tt.fields.cm.DeepCopy())
			client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
			helper := &ctrl.Helper{}
			patch1 := ApplyMethod(reflect.TypeOf(helper), "CleanUpFuse", func(_ *ctrl.Helper) (int, error) {
				return 0, nil
			})
			defer patch1.Reset()
			e := &GooseFSEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    client,
				Log:       tt.fields.log,
			}
			if err := e.cleanAll(); (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.cleanAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGooseFSEngineReleasePorts(t *testing.T) {
	type fields struct {
		runtime     *datav1alpha1.GooseFSRuntime
		name        string
		namespace   string
		runtimeType string
		cm          *corev1.ConfigMap
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "spark",
			fields: fields{
				name:        "spark",
				namespace:   "fluid",
				runtimeType: "goosefs",
				cm: &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-goosefs-values",
						Namespace: "fluid",
					},
					Data: map[string]string{"data": mockConfigMapData},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			portRange := "26000-32000"
			pr, _ := net.ParsePortRange(portRange)
			testObjs := []runtime.Object{}
			testObjs = append(testObjs, tt.fields.cm.DeepCopy())
			client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

			e := &GooseFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    client,
				Log:       fake.NullLogger(),
			}

			err := portallocator.SetupRuntimePortAllocator(client, pr, "bitmap", GetReservedPorts)
			if err != nil {
				t.Fatalf("Failed to set up runtime port allocator due to %v", err)
			}
			allocator, _ := portallocator.GetRuntimePortAllocator()
			patch1 := ApplyMethod(reflect.TypeOf(allocator), "ReleaseReservedPorts",
				func(_ *portallocator.RuntimePortAllocator, ports []int) {
				})
			defer patch1.Reset()

			if err := e.releasePorts(); (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.releasePorts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGooseFSEngineCleanupCache(t *testing.T) {
	type fields struct {
		name      string
		namespace string
		Log       logr.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "spark",
			fields: fields{
				name:      "spark",
				namespace: "field",
				Log:       fake.NullLogger(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &GooseFSEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       tt.fields.Log,
			}

			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetReportSummary",
				func(_ *GooseFSEngine) (string, error) {
					summary := mockGooseFSReportSummary()
					return summary, nil
				})
			defer patch1.Reset()

			patch2 := ApplyFunc(utils.GetDataset,
				func(_ client.Client, _ string, _ string) (*datav1alpha1.Dataset, error) {
					d := &datav1alpha1.Dataset{
						Status: datav1alpha1.DatasetStatus{
							UfsTotal: "19.07MiB",
						},
					}
					return d, nil
				})
			defer patch2.Reset()

			patch3 := ApplyMethod(reflect.TypeOf(engine), "GetCacheHitStates",
				func(_ *GooseFSEngine) cacheHitStates {
					return cacheHitStates{
						bytesReadLocal:  20310917,
						bytesReadUfsAll: 32243712,
					}
				})
			defer patch3.Reset()

			if err := engine.cleanupCache(); (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.cleanupCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGooseFSEngineDestroyMaster(t *testing.T) {
	type fields struct {
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "spark",
			fields: fields{
				name:      "spark",
				namespace: "fluid",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
			}

			patch1 := ApplyFunc(helm.CheckRelease,
				func(_ string, _ string) (bool, error) {
					d := true
					return d, nil
				})
			defer patch1.Reset()

			patch2 := ApplyFunc(helm.DeleteRelease,
				func(_ string, _ string) error {
					return nil
				})
			defer patch2.Reset()

			if err := e.destroyMaster(); (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.destroyMaster() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
