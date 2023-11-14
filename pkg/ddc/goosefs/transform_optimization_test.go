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
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestOptimizeDefaultProperties(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
		key          string
		expect       string
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Properties: map[string]string{},
			},
		}, &GooseFS{}, "goosefs.master.journal.type", "UFS"},
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Properties: map[string]string{},
			},
		}, &GooseFS{Master: Master{
			Replicas: 3,
		}}, "goosefs.master.journal.type", "EMBEDDED"},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{}
		engine.optimizeDefaultProperties(test.runtime, test.goosefsValue)
		if test.goosefsValue.Properties[test.key] != test.expect {
			t.Errorf("expected %s, got %v for key %s", test.expect, test.goosefsValue.Properties[test.key], test.key)
		}
	}
}

func TestOptimizeDefaultPropertiesAndFuseForHTTP(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
		dataset      *datav1alpha1.Dataset
		key          string
		expect       string
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Properties: map[string]string{},
			},
		}, &GooseFS{
			Properties: map[string]string{},
			Fuse: Fuse{
				Args: []string{"fuse", "--fuse-opts=kernel_cache,rw,max_read=131072,attr_timeout=7200,entry_timeout=7200,nonempty"},
			},
		},
			&datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{MountPoint: "https://mirrors.bit.edu.cn/apache/zookeeper/zookeeper-3.6.2/"},
						//{MountPoint: "local:///root/test-data"},
					},
				},
			},
			"goosefs.user.block.size.bytes.default", "256MB"},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{}
		engine.optimizeDefaultProperties(test.runtime, test.goosefsValue)
		engine.optimizeDefaultPropertiesAndFuseForHTTP(test.runtime, test.dataset, test.goosefsValue)
		if test.goosefsValue.Properties[test.key] != test.expect {
			t.Errorf("expected %s, got %v for key %s", test.expect, test.goosefsValue.Properties[test.key], test.key)
		}
	}
}

func TestOptimizeDefaultPropertiesWithSet(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
		key          string
		expect       string
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Properties: map[string]string{
					"goosefs.fuse.jnifuse.enabled": "false",
				},
			},
		}, &GooseFS{}, "goosefs.fuse.jnifuse.enabled", "false"},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{}
		engine.optimizeDefaultProperties(test.runtime, test.goosefsValue)
		if test.goosefsValue.Properties[test.key] != test.expect {
			t.Errorf("expected %s, got %v for key %s", test.expect, test.goosefsValue.Properties[test.key], test.key)
		}
	}
}

func TestSetDefaultPropertiesWithoutSet(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
		key          string
		value        string
		expect       string
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Properties: map[string]string{},
			},
		}, &GooseFS{
			Properties: map[string]string{},
		}, "goosefs.fuse.jnifuse.enabled", "true", "true"},
	}
	for _, test := range tests {
		setDefaultProperties(test.runtime, test.goosefsValue, test.key, test.value)
		if test.value != test.expect {
			t.Errorf("expected %v, got %v for key %s", test.expect, test.value, test.key)
		}
	}
}

func TestSetDefaultPropertiesWithSet(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
		key          string
		value        string
		expect       string
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Properties: map[string]string{
					"goosefs.fuse.jnifuse.enabled": "false",
				},
			},
		}, &GooseFS{
			Properties: map[string]string{},
		}, "goosefs.fuse.jnifuse.enabled", "true", "false"},
	}
	for _, test := range tests {
		setDefaultProperties(test.runtime, test.goosefsValue, test.key, test.value)
		if test.value == test.expect {
			t.Errorf("expected %v, got %v for key %s", test.expect, test.value, test.key)
		}
	}
}

func TestOptimizeDefaultForMasterNoValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
		expect       []string
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{},
		}, &GooseFS{
			Properties: map[string]string{},
		}, []string{"-Xmx6G",
			"-XX:+UnlockExperimentalVMOptions"}},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{}
		engine.optimizeDefaultForMaster(test.runtime, test.goosefsValue)
		if test.goosefsValue.Master.JvmOptions[1] != test.expect[1] {
			t.Errorf("expected %v, got %v", test.expect, test.goosefsValue.JvmOptions)
		}
	}
}

func TestOptimizeDefaultForMasterWithValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
		expect       []string
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Master: datav1alpha1.GooseFSCompTemplateSpec{
					JvmOptions: []string{"-Xmx4G"},
				},
			},
		}, &GooseFS{
			Properties: map[string]string{},
			Master:     Master{},
		}, []string{"-Xmx4G"}},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{}
		engine.optimizeDefaultForMaster(test.runtime, test.goosefsValue)
		if test.goosefsValue.Master.JvmOptions[0] != test.expect[0] {
			t.Errorf("expected %v, got %v", test.expect, test.goosefsValue.Master.JvmOptions)
		}
	}
}

func TestOptimizeDefaultForWorkerNoValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
		expect       []string
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{},
		}, &GooseFS{
			Properties: map[string]string{},
		}, []string{"-Xmx12G",
			"-XX:+UnlockExperimentalVMOptions",
			"-XX:MaxDirectMemorySize=32g"}},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{}
		engine.optimizeDefaultForWorker(test.runtime, test.goosefsValue)
		if test.goosefsValue.Worker.JvmOptions[1] != test.expect[1] {
			t.Errorf("expected %v, got %v", test.expect, test.goosefsValue.Worker.JvmOptions)
		}
	}
}

func TestOptimizeDefaultForWorkerWithValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
		expect       []string
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Worker: datav1alpha1.GooseFSCompTemplateSpec{
					JvmOptions: []string{"-Xmx4G"},
				},
			},
		}, &GooseFS{
			Properties: map[string]string{},
		}, []string{"-Xmx4G"}},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{}
		engine.optimizeDefaultForWorker(test.runtime, test.goosefsValue)
		if test.goosefsValue.Worker.JvmOptions[0] != test.expect[0] {
			t.Errorf("expected %v, got %v", test.expect, test.goosefsValue.Worker.JvmOptions)
		}
	}
}

func TestOptimizeDefaultForFuseNoValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
		expect       []string
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{},
		}, &GooseFS{
			Properties: map[string]string{},
		}, []string{"-Xmx16G",
			"-Xms16G",
			"-XX:+UseG1GC",
			"-XX:MaxDirectMemorySize=32g",
			"-XX:+UnlockExperimentalVMOptions"}},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{}
		engine.optimizeDefaultFuse(test.runtime, test.goosefsValue)
		if test.goosefsValue.Fuse.JvmOptions[1] != test.expect[1] {
			t.Errorf("expected %v, got %v", test.expect, test.goosefsValue.Fuse.JvmOptions)
		}
	}
}

func TestOptimizeDefaultForFuseWithValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
		expect       []string
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Fuse: datav1alpha1.GooseFSFuseSpec{
					JvmOptions: []string{"-Xmx4G"},
				},
			},
		}, &GooseFS{
			Properties: map[string]string{},
		}, []string{"-Xmx4G"}},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{}
		engine.optimizeDefaultFuse(test.runtime, test.goosefsValue)
		if test.goosefsValue.Fuse.JvmOptions[0] != test.expect[0] {
			t.Errorf("expected %v, got %v", test.expect, test.goosefsValue.Fuse.JvmOptions)
		}
	}
}

func TestGooseFSEngine_setPortProperties(t *testing.T) {
	type fields struct {
		runtime                *datav1alpha1.GooseFSRuntime
		name                   string
		namespace              string
		runtimeType            string
		Log                    logr.Logger
		Client                 client.Client
		gracefulShutdownLimits int32
		retryShutdown          int32
		initImage              string
		MetadataSyncDoneCh     chan base.MetadataSyncResult
	}
	type args struct {
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
	}

	var port int = 20000
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test",
			fields: fields{
				runtime: &datav1alpha1.GooseFSRuntime{},
			},
			args: args{
				runtime: &datav1alpha1.GooseFSRuntime{},
				goosefsValue: &GooseFS{
					Master: Master{
						Ports: Ports{
							Rpc:      port,
							Web:      port,
							Embedded: 0,
						},
					},
					Worker: Worker{
						Ports: Ports{
							Rpc: port,
							Web: port,
						},
						Resources: common.Resources{
							Requests: common.ResourceList{
								corev1.ResourceCPU:    "100m",
								corev1.ResourceMemory: "100Mi",
							},
						},
					},
					JobMaster: JobMaster{
						Ports: Ports{
							Rpc:      port,
							Web:      port,
							Embedded: 0,
						},
					},
					JobWorker: JobWorker{
						Ports: Ports{
							Rpc:  port,
							Web:  port,
							Data: port,
						},
						Resources: common.Resources{
							Requests: common.ResourceList{
								corev1.ResourceCPU:    "100m",
								corev1.ResourceMemory: "100Mi",
							},
						},
					},
					Properties: map[string]string{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{
				runtime:                tt.fields.runtime,
				name:                   tt.fields.name,
				namespace:              tt.fields.namespace,
				runtimeType:            tt.fields.runtimeType,
				Log:                    tt.fields.Log,
				Client:                 tt.fields.Client,
				gracefulShutdownLimits: tt.fields.gracefulShutdownLimits,
				retryShutdown:          tt.fields.retryShutdown,
				initImage:              tt.fields.initImage,
				MetadataSyncDoneCh:     tt.fields.MetadataSyncDoneCh,
			}
			e.setPortProperties(tt.args.runtime, tt.args.goosefsValue)
			key := tt.args.goosefsValue.Properties["goosefs.master.rpc.port"]
			if key != strconv.Itoa(port) {
				t.Errorf("expected %d, got %s", port, tt.args.goosefsValue.Properties["goosefs.master.rpc.port"])
			}
		})
	}
}
