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
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestOptimizeDefaultProperties(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
		key          string
		expect       string
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Properties: map[string]string{},
			},
		}, &Alluxio{}, "alluxio.fuse.jnifuse.enabled", "true"},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{}
		engine.optimizeDefaultProperties(test.runtime, test.alluxioValue)
		if test.alluxioValue.Properties[test.key] != test.expect {
			t.Errorf("expected %s, got %v for key %s", test.expect, test.alluxioValue.Properties[test.key], test.key)
		}
	}
}

func TestOptimizeDefaultPropertiesWithSet(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
		key          string
		expect       string
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Properties: map[string]string{
					"alluxio.fuse.jnifuse.enabled": "false",
				},
			},
		}, &Alluxio{}, "alluxio.fuse.jnifuse.enabled", "false"},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{}
		engine.optimizeDefaultProperties(test.runtime, test.alluxioValue)
		if test.alluxioValue.Properties[test.key] != test.expect {
			t.Errorf("expected %s, got %v for key %s", test.expect, test.alluxioValue.Properties[test.key], test.key)
		}
	}
}

func TestSetDefaultPropertiesWithoutSet(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
		key          string
		value        string
		expect       string
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Properties: map[string]string{},
			},
		}, &Alluxio{
			Properties: map[string]string{},
		}, "alluxio.fuse.jnifuse.enabled", "true", "true"},
	}
	for _, test := range tests {
		setDefaultProperties(test.runtime, test.alluxioValue, test.key, test.value)
		if test.value != test.expect {
			t.Errorf("expected %v, got %v for key %s", test.expect, test.value, test.key)
		}
	}
}

func TestSetDefaultPropertiesWithSet(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
		key          string
		value        string
		expect       string
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Properties: map[string]string{
					"alluxio.fuse.jnifuse.enabled": "false",
				},
			},
		}, &Alluxio{
			Properties: map[string]string{},
		}, "alluxio.fuse.jnifuse.enabled", "true", "false"},
	}
	for _, test := range tests {
		setDefaultProperties(test.runtime, test.alluxioValue, test.key, test.value)
		if test.value == test.expect {
			t.Errorf("expected %v, got %v for key %s", test.expect, test.value, test.key)
		}
	}
}

func TestOptimizeDefaultForMasterNoValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
		expect       []string
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{},
		}, &Alluxio{
			Properties: map[string]string{},
		}, []string{"-Xmx6G",
			"-XX:+UnlockExperimentalVMOptions"}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{}
		engine.optimizeDefaultForMaster(test.runtime, test.alluxioValue)
		if test.alluxioValue.Master.JvmOptions[1] != test.expect[1] {
			t.Errorf("expected %v, got %v", test.expect, test.alluxioValue.JvmOptions)
		}
	}
}

func TestOptimizeDefaultForMasterWithValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
		expect       []string
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Master: datav1alpha1.AlluxioCompTemplateSpec{
					JvmOptions: []string{"-Xmx4G"},
				},
			},
		}, &Alluxio{
			Properties: map[string]string{},
			Master:     Master{},
		}, []string{"-Xmx4G"}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{}
		engine.optimizeDefaultForMaster(test.runtime, test.alluxioValue)
		if test.alluxioValue.Master.JvmOptions[0] != test.expect[0] {
			t.Errorf("expected %v, got %v", test.expect, test.alluxioValue.Master.JvmOptions)
		}
	}
}

func TestOptimizeDefaultForWorkerNoValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
		expect       []string
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{},
		}, &Alluxio{
			Properties: map[string]string{},
		}, []string{"-Xmx12G",
			"-XX:+UnlockExperimentalVMOptions",
			"-XX:MaxDirectMemorySize=32g"}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{}
		engine.optimizeDefaultForWorker(test.runtime, test.alluxioValue)
		if test.alluxioValue.Worker.JvmOptions[1] != test.expect[1] {
			t.Errorf("expected %v, got %v", test.expect, test.alluxioValue.Worker.JvmOptions)
		}
	}
}

func TestOptimizeDefaultForWorkerWithValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
		expect       []string
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Worker: datav1alpha1.AlluxioCompTemplateSpec{
					JvmOptions: []string{"-Xmx4G"},
				},
			},
		}, &Alluxio{
			Properties: map[string]string{},
		}, []string{"-Xmx4G"}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{}
		engine.optimizeDefaultForWorker(test.runtime, test.alluxioValue)
		if test.alluxioValue.Worker.JvmOptions[0] != test.expect[0] {
			t.Errorf("expected %v, got %v", test.expect, test.alluxioValue.Worker.JvmOptions)
		}
	}
}

func TestOptimizeDefaultForFuseNoValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
		expect       []string
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{},
		}, &Alluxio{
			Properties: map[string]string{},
		}, []string{"-Xmx16G",
			"-Xms16G",
			"-XX:+UseG1GC",
			"-XX:MaxDirectMemorySize=32g",
			"-XX:+UnlockExperimentalVMOptions"}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{}
		engine.optimizeDefaultFuse(test.runtime, test.alluxioValue)
		if test.alluxioValue.Fuse.JvmOptions[1] != test.expect[1] {
			t.Errorf("expected %v, got %v", test.expect, test.alluxioValue.Fuse.JvmOptions)
		}
	}
}

func TestOptimizeDefaultForFuseWithValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
		expect       []string
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Fuse: datav1alpha1.AlluxioFuseSpec{
					JvmOptions: []string{"-Xmx4G"},
				},
			},
		}, &Alluxio{
			Properties: map[string]string{},
		}, []string{"-Xmx4G"}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{}
		engine.optimizeDefaultFuse(test.runtime, test.alluxioValue)
		if test.alluxioValue.Fuse.JvmOptions[0] != test.expect[0] {
			t.Errorf("expected %v, got %v", test.expect, test.alluxioValue.Fuse.JvmOptions)
		}
	}
}

func TestAlluxioEngine_setPortProperties(t *testing.T) {
	type fields struct {
		runtime                *datav1alpha1.AlluxioRuntime
		name                   string
		namespace              string
		runtimeType            string
		Log                    logr.Logger
		Client                 client.Client
		gracefulShutdownLimits int32
		retryShutdown          int32
		initImage              string
		MetadataSyncDoneCh     chan MetadataSyncResult
	}
	type args struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
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
				runtime: &datav1alpha1.AlluxioRuntime{},
			},
			args: args{
				runtime: &datav1alpha1.AlluxioRuntime{},
				alluxioValue: &Alluxio{
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
					},
					Properties: map[string]string{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AlluxioEngine{
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
			e.setPortProperties(tt.args.runtime, tt.args.alluxioValue)
			key := tt.args.alluxioValue.Properties["alluxio.master.rpc.port"]
			if key != strconv.Itoa(port) {
				t.Errorf("expected %d, got %s", port, tt.args.alluxioValue.Properties["alluxio.master.rpc.port"])
			}
		})
	}
}
