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

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/net"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

var (
	LocalMount = datav1alpha1.Mount{
		MountPoint: "local:///mnt/test",
		Name:       "local",
	}
)

func TestTransformFuse(t *testing.T) {

	var x int64 = 1000
	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	var tests = []struct {
		runtime *datav1alpha1.AlluxioRuntime
		dataset *datav1alpha1.Dataset
		value   *Alluxio
		expect  []string
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Fuse: datav1alpha1.AlluxioFuseSpec{},
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					LocalMount,
				},
				Owner: &datav1alpha1.User{
					UID: &x,
					GID: &x,
				},
			},
		}, &Alluxio{}, []string{"fuse", "--fuse-opts=kernel_cache,rw,uid=1000,gid=1000,allow_other"}},
	}
	for _, test := range tests {
		runtimeInfo, err := base.BuildRuntimeInfo("hbase", "fluid", "alluxio")
		if err != nil {
			t.Errorf("fail to create the runtimeInfo with error %v", err)
		}
		engine := &AlluxioEngine{
			runtimeInfo: runtimeInfo,
			Client:      fake.NewFakeClientWithScheme(testScheme),
		}
		engine.Log = ctrl.Log
		err = engine.transformFuse(test.runtime, test.dataset, test.value)
		if err != nil {
			t.Errorf("error %v", err)
		}
		if test.value.Fuse.Args[1] != test.expect[1] {
			t.Errorf("expected %v, got %v", test.expect, test.value.Fuse.Args)
		}
	}
}

func TestTransformMaster(t *testing.T) {
	testCases := map[string]struct {
		runtime   *datav1alpha1.AlluxioRuntime
		wantValue *Alluxio
	}{
		"test network mode case 1": {
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Master: datav1alpha1.AlluxioCompTemplateSpec{
						NetworkMode: datav1alpha1.ContainerNetworkMode,
					},
				},
			},
			wantValue: &Alluxio{
				Master: Master{
					HostNetwork: false,
				},
			},
		},
		"test network mode case 2": {
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Master: datav1alpha1.AlluxioCompTemplateSpec{
						NetworkMode: datav1alpha1.HostNetworkMode,
					},
				},
			},
			wantValue: &Alluxio{
				Master: Master{
					HostNetwork: true,
				},
			},
		},
		"test network mode case 3": {
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Master: datav1alpha1.AlluxioCompTemplateSpec{
						NetworkMode: datav1alpha1.HostNetworkMode,
					},
				},
			},
			wantValue: &Alluxio{
				Master: Master{
					HostNetwork: true,
				},
			},
		},
		"test hierarchical imagePullSecrets case1": {
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Master: datav1alpha1.AlluxioCompTemplateSpec{
						ImagePullSecrets: []corev1.LocalObjectReference{{Name: "secret-master"}},
					},
				},
			},
			wantValue: &Alluxio{
				Master: Master{
					ImagePullSecrets: []corev1.LocalObjectReference{{Name: "secret-master"}},
					HostNetwork:      true,
				},
			},
		},
	}

	engine := &AlluxioEngine{Log: fake.NullLogger()}
	ds := &datav1alpha1.Dataset{}
	for k, v := range testCases {
		gotValue := &Alluxio{}
		if err := engine.transformMasters(v.runtime, ds, gotValue); err == nil {
			if gotValue.Master.HostNetwork != v.wantValue.Master.HostNetwork {
				t.Errorf("check %s failure, got:%t,want:%t",
					k,
					gotValue.Master.HostNetwork,
					v.wantValue.Master.HostNetwork,
				)
			}
		}
		if len(v.wantValue.Master.ImagePullSecrets) > 0 {
			if !reflect.DeepEqual(gotValue.Master.ImagePullSecrets, v.wantValue.Master.ImagePullSecrets) {
				t.Errorf("check %s failure, got:%v,want:%v",
					k,
					gotValue.Master.ImagePullSecrets,
					v.wantValue.Master.ImagePullSecrets,
				)
			}
		}
	}
}

// TestTransformWorkers verifies that the transformWorkers function correctly transforms 
// the worker configuration of AlluxioRuntime into the expected Alluxio structure. 
// It tests different network modes, node selectors, and image pull secrets to ensure 
// correct transformation behavior.
func TestTransformWorkers(t *testing.T) {
	testCases := map[string]struct {
		runtime   *datav1alpha1.AlluxioRuntime
		wantValue *Alluxio
	}{
		"test network mode case 1": {
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Worker: datav1alpha1.AlluxioCompTemplateSpec{
						NetworkMode: datav1alpha1.ContainerNetworkMode,
					},
				},
			},
			wantValue: &Alluxio{
				Worker: Worker{
					HostNetwork: false,
				},
			},
		},
		"test network mode case 2": {
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Worker: datav1alpha1.AlluxioCompTemplateSpec{
						NetworkMode: datav1alpha1.HostNetworkMode,
					},
				},
			},
			wantValue: &Alluxio{
				Worker: Worker{
					HostNetwork: true,
				},
			},
		},
		"test network mode case 3": {
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Worker: datav1alpha1.AlluxioCompTemplateSpec{
						NetworkMode: datav1alpha1.HostNetworkMode,
					},
					TieredStore: datav1alpha1.TieredStore{},
				},
			},
			wantValue: &Alluxio{
				Worker: Worker{
					HostNetwork: true,
				},
			},
		},
		"test network mode case 4": {
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Worker: datav1alpha1.AlluxioCompTemplateSpec{
						NetworkMode: datav1alpha1.HostNetworkMode,
						NodeSelector: map[string]string{
							"workerSelector": "true",
						},
					},
					TieredStore: datav1alpha1.TieredStore{},
				},
			},
			wantValue: &Alluxio{
				Worker: Worker{
					HostNetwork: true,
					NodeSelector: map[string]string{
						"workerSelector": "true",
					},
				},
			},
		},
		"test hierarchical imagePullSecrets case1": {
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Worker: datav1alpha1.AlluxioCompTemplateSpec{
						ImagePullSecrets: []corev1.LocalObjectReference{{Name: "secret-worker"}},
					},
				},
			},
			wantValue: &Alluxio{
				Worker: Worker{
					ImagePullSecrets: []corev1.LocalObjectReference{{Name: "secret-worker"}},
					HostNetwork:      true,
				},
			},
		},
	}

	engine := &AlluxioEngine{Log: fake.NullLogger()}
	for k, v := range testCases {
		gotValue := &Alluxio{}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "alluxio", base.WithTieredStore(v.runtime.Spec.TieredStore))
		if err := engine.transformWorkers(v.runtime, gotValue); err == nil {
			if gotValue.Worker.HostNetwork != v.wantValue.Worker.HostNetwork {
				t.Errorf("check %s failure, got:%t,want:%t",
					k,
					gotValue.Worker.HostNetwork,
					v.wantValue.Worker.HostNetwork,
				)
			}
			if len(v.wantValue.Worker.NodeSelector) > 0 {
				if !reflect.DeepEqual(v.wantValue.Worker.NodeSelector, gotValue.Worker.NodeSelector) {
					t.Errorf("check %s failure, got:%v,want:%v",
						k,
						gotValue.Worker.NodeSelector,
						v.wantValue.Worker.NodeSelector,
					)
				}
			}
			if len(v.wantValue.Worker.ImagePullSecrets) > 0 {
				if !reflect.DeepEqual(v.wantValue.Worker.ImagePullSecrets, gotValue.Worker.ImagePullSecrets) {
					t.Errorf("check %s failure, got:%s,want:%s",
						k,
						gotValue.Worker.ImagePullSecrets,
						v.wantValue.Worker.ImagePullSecrets,
					)
				}
			}
		}
	}
}

func TestGenerateStaticPorts(t *testing.T) {
	engine := &AlluxioEngine{Log: fake.NullLogger(),
		runtime: &datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Master: datav1alpha1.AlluxioCompTemplateSpec{
					Replicas: 3,
				},
				APIGateway: datav1alpha1.AlluxioCompTemplateSpec{
					Enabled: true,
				},
			},
		}}
	gotValue := &Alluxio{}
	engine.generateStaticPorts(gotValue)
	expect := &Alluxio{
		Master: Master{
			Ports: Ports{
				Embedded: 19200,
				Rpc:      19998,
				Web:      19999,
			},
		}, JobMaster: JobMaster{
			Ports: Ports{
				Embedded: 20003,
				Rpc:      20001,
				Web:      20002,
			},
		}, APIGateway: APIGateway{
			Ports: Ports{
				Rest: 39999,
			},
		}, Worker: Worker{
			Ports: Ports{Rpc: 29999,
				Web: 30000},
		}, JobWorker: JobWorker{
			Ports: Ports{
				Rpc:  30001,
				Data: 30002,
				Web:  30003,
			},
		},
	}

	if !reflect.DeepEqual(expect, gotValue) {
		t.Errorf("Expect the value %v, but got %v", expect, gotValue)
	}
}

func TestTransformShortCircuit(t *testing.T) {
	engine := &AlluxioEngine{Log: fake.NullLogger()}

	type wantValue struct {
		wantShortCircuitPolicy string
		wantShortCircuit       ShortCircuit
		wantPropertyKey        string
		wantPropertyValue      string
	}

	type testCase struct {
		Name          string
		RuntimeInfo   base.RuntimeInfoInterface
		MockPatchFunc func(_ *base.RuntimeInfo) base.TieredStoreInfo
		Value         *Alluxio

		want wantValue
	}

	testsCases := []testCase{
		{
			Name:        "With emptyDir volume type tier",
			RuntimeInfo: &base.RuntimeInfo{},
			MockPatchFunc: func(_ *base.RuntimeInfo) base.TieredStoreInfo {
				return base.TieredStoreInfo{
					Levels: []base.Level{
						{
							MediumType: common.Memory,
							VolumeType: common.VolumeTypeEmptyDir,
						},
						{
							MediumType: common.SSD,
							VolumeType: common.VolumeTypeHostPath,
						},
					},
				}
			},
			Value: &Alluxio{
				Properties: map[string]string{},
			},
			want: wantValue{
				wantShortCircuitPolicy: "local",
				wantShortCircuit: ShortCircuit{
					Enable:     false,
					Policy:     "local",
					VolumeType: "emptyDir",
				},
				wantPropertyKey:   "alluxio.user.short.circuit.enabled",
				wantPropertyValue: "false",
			},
		},
		{
			Name:        "Without emptyDir volume type tier",
			RuntimeInfo: &base.RuntimeInfo{},
			MockPatchFunc: func(_ *base.RuntimeInfo) base.TieredStoreInfo {
				return base.TieredStoreInfo{
					Levels: []base.Level{
						{
							MediumType: common.Memory,
							VolumeType: common.VolumeTypeDefault,
						},
						{
							MediumType: common.SSD,
							VolumeType: common.VolumeTypeHostPath,
						},
					},
				}
			},
			Value: &Alluxio{
				Properties: map[string]string{},
			},
			want: wantValue{
				wantShortCircuitPolicy: "local",
				wantShortCircuit: ShortCircuit{
					Enable:     true,
					Policy:     "local",
					VolumeType: "emptyDir",
				},
			},
		},
	}

	for _, tt := range testsCases {
		patch := gomonkey.ApplyMethod(reflect.TypeOf(tt.RuntimeInfo), "GetTieredStoreInfo", tt.MockPatchFunc)
		engine.transformShortCircuit(tt.RuntimeInfo, tt.Value)

		if tt.Value.Fuse.ShortCircuitPolicy != "local" {
			t.Errorf("Expect Fuse.ShortCircuitPolicy=%s, got=%s", "local", tt.Value.Fuse.ShortCircuitPolicy)
		}

		if !reflect.DeepEqual(tt.Value.ShortCircuit, tt.want.wantShortCircuit) {
			t.Errorf("Expect ShortCircuit=%v, got=%v", tt.want.wantShortCircuit, tt.Value.ShortCircuit)
		}

		if len(tt.want.wantPropertyKey) > 0 {
			if val, ok := tt.Value.Properties[tt.want.wantPropertyKey]; !ok || val != tt.want.wantPropertyValue {
				t.Errorf("Expect Property has %s=%s, but got not expected", tt.want.wantPropertyKey, tt.want.wantPropertyValue)
			}
		}
		patch.Reset()
	}
}

func TestTransformPodMetadata(t *testing.T) {
	engine := &AlluxioEngine{Log: fake.NullLogger()}

	type testCase struct {
		Name    string
		Runtime *datav1alpha1.AlluxioRuntime
		Value   *Alluxio

		wantValue *Alluxio
	}

	testCases := []testCase{
		{
			Name: "set_common_labels_and_annotations",
			Runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					PodMetadata: datav1alpha1.PodMetadata{
						Labels:      map[string]string{"common-key": "common-value"},
						Annotations: map[string]string{"common-annotation": "val"},
					},
				},
			},
			Value: &Alluxio{},
			wantValue: &Alluxio{
				Master: Master{
					Labels:      map[string]string{"common-key": "common-value"},
					Annotations: map[string]string{"common-annotation": "val"},
				},
				Worker: Worker{
					Labels:      map[string]string{"common-key": "common-value"},
					Annotations: map[string]string{"common-annotation": "val"},
				},
				Fuse: Fuse{
					Labels:      map[string]string{"common-key": "common-value"},
					Annotations: map[string]string{"common-annotation": "val"},
				},
			},
		},
		{
			Name: "set_master_and_workers_labels_and_annotations",
			Runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					PodMetadata: datav1alpha1.PodMetadata{
						Labels:      map[string]string{"common-key": "common-value"},
						Annotations: map[string]string{"common-annotation": "val"},
					},
					Master: datav1alpha1.AlluxioCompTemplateSpec{
						PodMetadata: datav1alpha1.PodMetadata{
							Labels:      map[string]string{"common-key": "master-value"},
							Annotations: map[string]string{"common-annotation": "master-val"},
						},
					},
					Worker: datav1alpha1.AlluxioCompTemplateSpec{
						PodMetadata: datav1alpha1.PodMetadata{
							Labels:      map[string]string{"common-key": "worker-value"},
							Annotations: map[string]string{"common-annotation": "worker-val"},
						},
					},
				},
			},
			Value: &Alluxio{},
			wantValue: &Alluxio{
				Master: Master{
					Labels:      map[string]string{"common-key": "master-value"},
					Annotations: map[string]string{"common-annotation": "master-val"}},
				Worker: Worker{
					Labels:      map[string]string{"common-key": "worker-value"},
					Annotations: map[string]string{"common-annotation": "worker-val"},
				},
				Fuse: Fuse{
					Labels:      map[string]string{"common-key": "common-value"},
					Annotations: map[string]string{"common-annotation": "val"},
				},
			},
		},
	}

	for _, tt := range testCases {
		err := engine.transformPodMetadata(tt.Runtime, tt.Value)
		if err != nil {
			t.Fatalf("test name: %s. Expect err = nil, but got err = %v", tt.Name, err)
		}

		if !reflect.DeepEqual(tt.Value, tt.wantValue) {
			t.Fatalf("test name: %s. Expect value %v, but got %v", tt.Name, tt.wantValue, tt.Value)
		}
	}
}

func TestGetMediumTypeFromVolumeSource(t *testing.T) {
	engine := &AlluxioEngine{Log: fake.NullLogger()}

	type testCase struct {
		name              string
		defaultMediumType string
		level             base.Level

		wantMediumType string
	}

	testCases := []testCase{
		{
			name:              "no_volume_resource",
			defaultMediumType: "MEM",
			level:             base.Level{VolumeType: common.VolumeTypeEmptyDir},
			wantMediumType:    "MEM",
		},
		{
			name:              "set_emptyDir_volume_resource",
			defaultMediumType: "SSD",
			level: base.Level{
				VolumeType: common.VolumeTypeEmptyDir,
				VolumeSource: datav1alpha1.VolumeSource{
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{Medium: "Memory"},
					},
				},
			},
			wantMediumType: "Memory",
		},
	}

	for _, tt := range testCases {
		got := engine.getMediumTypeFromVolumeSource(tt.defaultMediumType, tt.level)
		if got != tt.wantMediumType {
			t.Fatalf("test name: %s. Expected value=%s, but got value=%s", tt.name, tt.wantMediumType, got)
		}
	}
}

func TestAlluxioEngine_allocateSinglePort(t *testing.T) {
	type fields struct {
		runtime            *datav1alpha1.AlluxioRuntime
		name               string
		namespace          string
		runtimeType        string
		Log                logr.Logger
		Client             client.Client
		retryShutdown      int32
		initImage          string
		MetadataSyncDoneCh chan base.MetadataSyncResult
	}
	type args struct {
		allocatedPorts []int
		alluxioValue   *Alluxio
	}

	tests := []struct {
		name      string
		fields    fields
		args      args
		wantPorts []int
	}{
		{
			name: "test_set_Properties",
			fields: fields{
				runtime: &datav1alpha1.AlluxioRuntime{},
			},
			args: args{
				allocatedPorts: []int{20001, 20001, 20001, 20001},
				alluxioValue: &Alluxio{
					Properties: map[string]string{
						"alluxio.master.rpc.port": strconv.Itoa(30001),
						"alluxio.master.web.port": strconv.Itoa(30001),
						"alluxio.worker.rpc.port": strconv.Itoa(30001),
						"alluxio.worker.web.port": strconv.Itoa(30001),
					},
				},
			},
			wantPorts: []int{30001},
		},
		{
			name: "test_unset_Properties",
			fields: fields{
				runtime: &datav1alpha1.AlluxioRuntime{},
			},
			args: args{
				allocatedPorts: []int{20001, 20002, 20003, 20004},
				alluxioValue: &Alluxio{
					Properties: map[string]string{},
				},
			},
			wantPorts: []int{20001, 20002, 20003, 20004},
		},
		{
			name: "test_set_runtime",
			fields: fields{
				runtime: &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Master: datav1alpha1.AlluxioCompTemplateSpec{
							Ports: map[string]int{
								"rpc": 20005,
								"web": 20005,
							},
						},
						Worker: datav1alpha1.AlluxioCompTemplateSpec{
							Ports: map[string]int{
								"rpc": 20005,
								"web": 20005,
							},
						},
					},
				},
			},
			args: args{
				allocatedPorts: []int{20001, 20002, 20003, 20004},
				alluxioValue: &Alluxio{
					Properties: map[string]string{},
				},
			},
			wantPorts: []int{20005},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AlluxioEngine{
				runtime:            tt.fields.runtime,
				name:               tt.fields.name,
				namespace:          tt.fields.namespace,
				runtimeType:        tt.fields.runtimeType,
				Log:                tt.fields.Log,
				Client:             tt.fields.Client,
				retryShutdown:      tt.fields.retryShutdown,
				initImage:          tt.fields.initImage,
				MetadataSyncDoneCh: tt.fields.MetadataSyncDoneCh,
			}

			index := 0
			preIndex := 0

			tt.args.alluxioValue.Master.Ports.Rpc, index = e.allocateSinglePort(tt.args.alluxioValue, "alluxio.master.rpc.port", tt.args.allocatedPorts, index, tt.fields.runtime.Spec.Master.Ports, "rpc")
			if tt.args.alluxioValue.Master.Ports.Rpc != tt.wantPorts[preIndex] {
				t.Errorf("port %s expected %d, got %d", "alluxio.master.rpc.port", tt.wantPorts[preIndex], tt.args.alluxioValue.Master.Ports.Rpc)
			}
			preIndex = index
			tt.args.alluxioValue.Master.Ports.Web, index = e.allocateSinglePort(tt.args.alluxioValue, "alluxio.master.web.port", tt.args.allocatedPorts, index, tt.fields.runtime.Spec.Master.Ports, "web")
			if tt.args.alluxioValue.Master.Ports.Web != tt.wantPorts[preIndex] {
				t.Errorf("port %s expected %d, got %d", "alluxio.master.web.port", tt.wantPorts[preIndex], tt.args.alluxioValue.Master.Ports.Web)
			}
			preIndex = index
			tt.args.alluxioValue.Worker.Ports.Rpc, index = e.allocateSinglePort(tt.args.alluxioValue, "alluxio.worker.rpc.port", tt.args.allocatedPorts, index, tt.fields.runtime.Spec.Worker.Ports, "rpc")
			if tt.args.alluxioValue.Worker.Ports.Rpc != tt.wantPorts[preIndex] {
				t.Errorf("port %s expected %d, got %d", "alluxio.worker.rpc.port", tt.wantPorts[preIndex], tt.args.alluxioValue.Worker.Ports.Rpc)
			}
			preIndex = index
			tt.args.alluxioValue.Worker.Ports.Web, _ = e.allocateSinglePort(tt.args.alluxioValue, "alluxio.worker.web.port", tt.args.allocatedPorts, index, tt.fields.runtime.Spec.Worker.Ports, "web")
			if tt.args.alluxioValue.Worker.Ports.Web != tt.wantPorts[preIndex] {
				t.Errorf("port %s expected %d, got %d", "alluxio.worker.web.port", tt.wantPorts[preIndex], tt.args.alluxioValue.Worker.Ports.Web)
			}

		})
	}
}

// Test function to test the allocatePorts functionality in the AlluxioEngine
// It sets up the port range for allocation and initializes a port allocator
// The test cases are structured with different input values and expected results
// Fields include various configuration values for the Alluxio runtime and test setup
// Args contain the allocated ports and Alluxio instance to be used in the test cases
// Test case for setting properties in Alluxio runtime with given allocated ports
// The test checks that the allocated ports are properly mapped to the expected ports
// Initialize the AlluxioEngine instance with the test case fields.
// Call allocatePorts to allocate ports and check for errors.
// Check if the allocated APIGateway port matches the expected value.
func TestAlluxioEngine_allocatePorts(t *testing.T) {
	pr := net.ParsePortRangeOrDie("20000-21000")
	err := portallocator.SetupRuntimePortAllocator(nil, pr, "bitmap", dummy)
	if err != nil {
		t.Fatal(err.Error())
	}
	type fields struct {
		runtime            *datav1alpha1.AlluxioRuntime
		name               string
		namespace          string
		runtimeType        string
		Log                logr.Logger
		Client             client.Client
		retryShutdown      int32
		initImage          string
		MetadataSyncDoneCh chan base.MetadataSyncResult
	}
	type args struct {
		allocatedPorts []int
		alluxioValue   *Alluxio
	}

	tests := []struct {
		name      string
		fields    fields
		args      args
		wantPorts []int
	}{
		{
			name: "test_set_Properties",
			fields: fields{
				runtime: &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						APIGateway: datav1alpha1.AlluxioCompTemplateSpec{
							Enabled: true,
						},
					},
				},
			},
			args: args{
				allocatedPorts: []int{20001, 20001, 20001, 20001},
				alluxioValue: &Alluxio{
					Properties: map[string]string{
						"alluxio.master.rpc.port": strconv.Itoa(30001),
						"alluxio.master.web.port": strconv.Itoa(30001),
						"alluxio.worker.rpc.port": strconv.Itoa(30001),
						"alluxio.worker.web.port": strconv.Itoa(30001),
						"alluxio.proxy.web.port":  strconv.Itoa(30001),
					},
				},
			},
			wantPorts: []int{30001},
		},
		{
			name: "test_set_runtime_Properties",
			fields: fields{
				runtime: &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Master: datav1alpha1.AlluxioCompTemplateSpec{
							Ports: map[string]int{
								"rpc": 30002,
								"web": 30002,
							},
						},
						Worker: datav1alpha1.AlluxioCompTemplateSpec{
							Ports: map[string]int{
								"rpc": 30002,
								"web": 30002,
							},
						},
						APIGateway: datav1alpha1.AlluxioCompTemplateSpec{
							Enabled: true,
							Ports: map[string]int{
								"web": 30002,
							},
						},
					},
				},
			},
			args: args{
				allocatedPorts: []int{20001, 20001, 20001, 20001},
				alluxioValue: &Alluxio{
					Properties: map[string]string{
						"alluxio.master.rpc.port": strconv.Itoa(30001),
						"alluxio.master.web.port": strconv.Itoa(30001),
						"alluxio.worker.rpc.port": strconv.Itoa(30001),
						"alluxio.worker.web.port": strconv.Itoa(30001),
						"alluxio.proxy.web.port":  strconv.Itoa(30001),
					},
				},
			},
			wantPorts: []int{30001},
		},
		{
			name: "test_set_runtime_NoProperties",
			fields: fields{
				runtime: &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Master: datav1alpha1.AlluxioCompTemplateSpec{
							Ports: map[string]int{
								"rpc": 30002,
								"web": 30002,
							},
						},
						Worker: datav1alpha1.AlluxioCompTemplateSpec{
							Ports: map[string]int{
								"rpc": 30002,
								"web": 30002,
							},
						},
						APIGateway: datav1alpha1.AlluxioCompTemplateSpec{
							Enabled: true,
							Ports: map[string]int{
								"web": 30002,
							},
						},
					},
				},
			},
			args: args{
				allocatedPorts: []int{20001, 20001, 20001, 20001},
				alluxioValue: &Alluxio{
					Properties: map[string]string{},
				},
			},
			wantPorts: []int{30002},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AlluxioEngine{
				runtime:            tt.fields.runtime,
				name:               tt.fields.name,
				namespace:          tt.fields.namespace,
				runtimeType:        tt.fields.runtimeType,
				Log:                tt.fields.Log,
				Client:             tt.fields.Client,
				retryShutdown:      tt.fields.retryShutdown,
				initImage:          tt.fields.initImage,
				MetadataSyncDoneCh: tt.fields.MetadataSyncDoneCh,
			}

			err := e.allocatePorts(tt.args.alluxioValue, e.runtime)
			if err != nil {
				t.Errorf("allocatePorts err: %v", err)
			}
			if tt.args.alluxioValue.APIGateway.Ports.Rest != tt.wantPorts[0] {
				t.Errorf("expect %d got %d", tt.wantPorts[0], tt.args.alluxioValue.APIGateway.Ports.Rest)
			}

		})
	}
}

// TestTransformMasterProperties tests the transformMasters function of AlluxioEngine.
// It verifies whether the master properties are correctly transformed based on the given runtime configuration.
// The test cases ensure that the properties from the master template take precedence over the global properties.
//
// Test Cases:
// 1. "master properties is not null":
//    - Ensures that when master-specific properties exist, they override the global properties.
//
// 2. "properties is not null for master":
//    - Ensures that both master-specific and additional global properties are correctly handled.
//
// The function iterates over multiple test cases and checks if the transformed properties
// match the expected values. If the transformation does not produce the expected result, the test fails.
func TestTransformMasterProperties(t *testing.T) {
	engine := &AlluxioEngine{Log: fake.NullLogger()}

	type testCase struct {
		Name      string
		Runtime   *datav1alpha1.AlluxioRuntime
		Value     *Alluxio
		DataSet   *datav1alpha1.Dataset
		wantValue *Alluxio
	}

	testCases := []testCase{
		{
			Name: "master properties is not null",
			Runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Properties: map[string]string{
						"alluxio.master.rpc.executor.keepalive":     "45sec",
						"alluxio.master.rpc.executor.max.pool.size": "300",
					},
					Master: datav1alpha1.AlluxioCompTemplateSpec{
						Properties: map[string]string{
							"alluxio.master.rpc.executor.keepalive":     "30sec",
							"alluxio.master.rpc.executor.max.pool.size": "100",
						},
					},
				},
			},
			Value:   &Alluxio{},
			DataSet: &datav1alpha1.Dataset{},
			wantValue: &Alluxio{
				Master: Master{
					Properties: map[string]string{
						"alluxio.master.rpc.executor.keepalive":     "30sec",
						"alluxio.master.rpc.executor.max.pool.size": "100",
					},
				},
				Properties: map[string]string{
					"alluxio.master.rpc.executor.keepalive":     "30sec",
					"alluxio.master.rpc.executor.max.pool.size": "100",
				},
			},
		},
		{
			Name: "properties is not null for master",
			Runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Properties: map[string]string{
						"alluxio.worker.block.heartbeat.interval":   "300sec",
						"alluxio.master.rpc.executor.keepalive":     "45sec",
						"alluxio.master.rpc.executor.max.pool.size": "300",
					},
					Master: datav1alpha1.AlluxioCompTemplateSpec{
						Properties: map[string]string{
							"alluxio.master.rpc.executor.keepalive":     "30sec",
							"alluxio.master.rpc.executor.max.pool.size": "100",
						},
					},
				},
			},
			Value:   &Alluxio{},
			DataSet: &datav1alpha1.Dataset{},
			wantValue: &Alluxio{
				Master: Master{
					Properties: map[string]string{
						"alluxio.master.rpc.executor.keepalive":     "30sec",
						"alluxio.master.rpc.executor.max.pool.size": "100",
					},
				},
				Properties: map[string]string{
					"alluxio.master.rpc.executor.keepalive":     "30sec",
					"alluxio.master.rpc.executor.max.pool.size": "100",
					"alluxio.worker.block.heartbeat.interval":   "300sec",
				},
			},
		},
	}

	for _, tt := range testCases {
		err := engine.transformMasters(tt.Runtime, tt.DataSet, tt.Value)
		if err != nil {
			t.Fatalf("test name: %s. Expect err = nil, but got err = %v", tt.Name, err)
		}
		for k, v := range tt.Value.Properties {
			if data, ok := tt.wantValue.Properties[k]; ok {
				if data != v {
					t.Fatalf("test name: %s. expect %s got %s", tt.Name, v, data)
				}
			} else {
				t.Fatalf("test name: %s. expect %s in value,but not in ", tt.Name, k)
			}
		}
	}
}

func TestTransformWorkerProperties(t *testing.T) {
	engine := &AlluxioEngine{Log: fake.NullLogger()}

	type testCase struct {
		Name      string
		Runtime   *datav1alpha1.AlluxioRuntime
		Value     *Alluxio
		wantValue *Alluxio
	}

	testCases := []testCase{
		{
			Name: "worker properties is not null",
			Runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Properties: map[string]string{
						"alluxio.worker.block.heartbeat.interval":      "300sec",
						"alluxio.worker.block.heartbeat.timeout=1hour": "5hour",
					},
					Worker: datav1alpha1.AlluxioCompTemplateSpec{
						Properties: map[string]string{
							"alluxio.worker.block.heartbeat.interval":      "30sec",
							"alluxio.worker.block.heartbeat.timeout=1hour": "1hour",
						},
					},
				},
			},
			Value: &Alluxio{},
			wantValue: &Alluxio{
				Worker: Worker{
					Properties: map[string]string{
						"alluxio.worker.block.heartbeat.interval":      "30sec",
						"alluxio.worker.block.heartbeat.timeout=1hour": "1hour",
					},
				},
				Properties: map[string]string{
					"alluxio.worker.block.heartbeat.interval":      "30sec",
					"alluxio.worker.block.heartbeat.timeout=1hour": "1hour",
				},
			},
		},
	}

	for _, tt := range testCases {
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "alluxio", base.WithTieredStore(tt.Runtime.Spec.TieredStore))
		err := engine.transformWorkers(tt.Runtime, tt.Value)
		if err != nil {
			t.Fatalf("test name: %s. Expect err = nil, but got err = %v", tt.Name, err)
		}
		for k, v := range tt.Value.Properties {
			if data, ok := tt.wantValue.Properties[k]; ok {
				if data != v {
					t.Fatalf("test name: %s. expect %s got %s", tt.Name, v, data)
				}
			} else {
				t.Fatalf("test name: %s. expect %s in value,but not in ", tt.Name, k)
			}
		}
	}
}

func TestTransformFuseProperties(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	engine := &AlluxioEngine{
		Log:         fake.NullLogger(),
		Client:      fake.NewFakeClientWithScheme(testScheme),
		runtimeInfo: runtimeInfo,
	}
	var x int64 = 1000
	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	type testCase struct {
		Name      string
		Runtime   *datav1alpha1.AlluxioRuntime
		Value     *Alluxio
		DataSet   *datav1alpha1.Dataset
		wantValue *Alluxio
	}

	testCases := []testCase{
		{
			Name: "fuse properties is not null",
			Runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Properties: map[string]string{
						"alluxio.fuse.cached.paths.max": "100000",
						"alluxio.fuse.maxcache.bytes":   "1MB",
					},
					Fuse: datav1alpha1.AlluxioFuseSpec{
						Properties: map[string]string{
							"alluxio.fuse.cached.paths.max": "1000",
							"alluxio.fuse.maxcache.bytes":   "2MB",
						},
					},
				},
			},
			Value: &Alluxio{},
			DataSet: &datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						LocalMount,
					},
					Owner: &datav1alpha1.User{
						UID: &x,
						GID: &x,
					},
				},
			},
			wantValue: &Alluxio{
				Fuse: Fuse{
					Properties: map[string]string{
						"alluxio.fuse.cached.paths.max": "1000",
						"alluxio.fuse.maxcache.bytes":   "2MB",
					},
				},
				Properties: map[string]string{
					"alluxio.fuse.cached.paths.max": "1000",
					"alluxio.fuse.maxcache.bytes":   "2MB",
				},
			},
		},
	}

	for _, tt := range testCases {
		engine.Log = ctrl.Log
		err := engine.transformFuse(tt.Runtime, tt.DataSet, tt.Value)
		if err != nil {
			t.Fatalf("test name: %s. Expect err = nil, but got err = %v", tt.Name, err)
		}
		for k, v := range tt.Value.Properties {
			if data, ok := tt.wantValue.Properties[k]; ok {
				if data != v {
					t.Fatalf("test name: %s. expect %s got %s", tt.Name, v, data)
				}
			} else {
				t.Fatalf("test name: %s. expect %s in value,but not in ", tt.Name, k)
			}
		}
	}
}

// TestGenerateNonNativeMountsInfo validates the generation of non-local storage mount configurations for Alluxio engine.
//
// Functionality:
// - Tests the logic that converts Dataset storage definitions into Alluxio mount commands
// - Filters out local storage types (e.g. pvc://)
// - Handles security credential injection from Kubernetes Secrets
// - Verifies command argument formatting for different storage protocols
//
// Parameters:
//   - t *（testing.T） : Go test framework handler
//
// Returns:
//   - No direct return value
//   - Fails test through t.Fatalf if:
//     1. Unexpected error occurs during generation (err != nil)
//     2. Generated commands mismatch wantValue expectations
//     3. Protocol handling violates defined rules
func TestGenerateNonNativeMountsInfo(t *testing.T) {
	const (
		SecretName = "ds-secret"
		SecretKey  = "secret-key"
	)
	engine := &AlluxioEngine{Log: fake.NullLogger()}
	var x int64 = 1000
	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	type testCase struct {
		Name      string
		DataSet   *datav1alpha1.Dataset
		wantValue []string
	}

	testCases := []testCase{
		{
			Name: "generate non native mount infos",
			DataSet: &datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: "https://mirrors.bit.edu.cn/apache/hbase",
						Name:       "hbase",
						Shared:     true,
						ReadOnly:   true,
					}, {
						MountPoint: "oss://oss.com/test",
						Name:       "oss",
						EncryptOptions: []datav1alpha1.EncryptOption{
							{
								Name: "secret",
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: SecretName,
										Key:  SecretKey,
									},
								},
							},
						},
					}, LocalMount, {
						MountPoint: "pvc:///mnt/test",
						Name:       "pvc",
					}},
					Owner: &datav1alpha1.User{
						UID: &x,
						GID: &x,
					},
				},
			},
			wantValue: []string{
				"/hbase https://mirrors.bit.edu.cn/apache/hbase --readonly --shared",
				fmt.Sprintf("/oss oss://oss.com/test --option secret=/etc/fluid/secrets/%s/%s", SecretName, SecretKey),
			},
		},
	}

	for _, tt := range testCases {
		engine.Log = ctrl.Log
		mounts, err := engine.generateNonNativeMountsInfo(tt.DataSet)
		if err != nil {
			t.Fatalf("test name: %s. Expect err = nil, but got err = %v", tt.Name, err)
		}
		if !reflect.DeepEqual(mounts, tt.wantValue) {
			t.Fatalf("test name: %s. expect %v, get %v", tt.Name, tt.wantValue, mounts)
		}
	}
}

func TestTransformMasterMountConfigMap(t *testing.T) {
	const (
		SecretName = "alluxio-secret"
		SecretKey  = "secret-key"
	)

	engine := &AlluxioEngine{Log: fake.NullLogger()}

	type testCase struct {
		Name      string
		Runtime   *datav1alpha1.AlluxioRuntime
		Value     *Alluxio
		DataSet   *datav1alpha1.Dataset
		wantValue *Alluxio
	}

	testCases := []testCase{
		{
			Name: "use mount config map",
			Runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{},
			},
			Value: &Alluxio{},
			DataSet: &datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: "https://mirrors.bit.edu.cn/apache/hbase",
							Name:       "hbase",
							EncryptOptions: []datav1alpha1.EncryptOption{
								{
									Name: "secret",
									ValueFrom: datav1alpha1.EncryptOptionSource{
										SecretKeyRef: datav1alpha1.SecretKeySelector{
											Name: SecretName,
											Key:  SecretKey,
										},
									},
								},
							},
						},
					},
				},
			},
			wantValue: &Alluxio{
				Master: Master{
					MountConfigStorage: ConfigmapStorageName,
					NonNativeMounts: []string{
						fmt.Sprintf("/hbase https://mirrors.bit.edu.cn/apache/hbase --option secret=/etc/fluid/secrets/%s/%s", SecretName, SecretKey),
					},
					Volumes: []corev1.Volume{
						{
							Name: fmt.Sprintf("alluxio-mount-secret-%s", SecretName),
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: SecretName,
								},
							},
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      fmt.Sprintf("alluxio-mount-secret-%s", SecretName),
							ReadOnly:  true,
							MountPath: fmt.Sprintf("/etc/fluid/secrets/%s", SecretName),
						},
					},
				},
			},
		},
	}

	for _, tt := range testCases {
		err := engine.transformMasters(tt.Runtime, tt.DataSet, tt.Value)
		if err != nil {
			t.Fatalf("test name: %s. Expect err = nil, but got err = %v", tt.Name, err)
		}
		if tt.Value.Master.MountConfigStorage != tt.wantValue.Master.MountConfigStorage {
			t.Fatalf("test name: %s. expect %s got %s", tt.Name,
				tt.wantValue.Master.MountConfigStorage, tt.Value.Master.MountConfigStorage)
		}
		if !reflect.DeepEqual(tt.Value.Master.NonNativeMounts, tt.wantValue.Master.NonNativeMounts) {
			t.Fatalf("test name: %s. expect %s got %s", tt.Name,
				tt.wantValue.Master.NonNativeMounts, tt.Value.Master.NonNativeMounts)
		}
		if !reflect.DeepEqual(tt.Value.Master.VolumeMounts, tt.wantValue.Master.VolumeMounts) {
			t.Fatalf("test name: %s. expect %v got %v", tt.Name,
				tt.wantValue.Master.VolumeMounts, tt.Value.Master.VolumeMounts)
		}
		if !reflect.DeepEqual(tt.Value.Master.Volumes, tt.wantValue.Master.Volumes) {
			t.Fatalf("test name: %s. expect %v got %v", tt.Name,
				tt.wantValue.Master.Volumes, tt.Value.Master.Volumes)
		}
	}
}
