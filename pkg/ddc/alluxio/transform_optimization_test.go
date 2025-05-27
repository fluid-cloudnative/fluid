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
	"reflect"
	"strconv"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	corev1 "k8s.io/api/core/v1"

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

func TestOptimizeDefaultPropertiesAndFuseForHTTP(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
		dataset      *datav1alpha1.Dataset
		key          string
		expect       string
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Properties: map[string]string{},
			},
		}, &Alluxio{
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
			"alluxio.user.block.size.bytes.default", "256MB"},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{}
		engine.optimizeDefaultProperties(test.runtime, test.alluxioValue)
		engine.optimizeDefaultPropertiesAndFuseForHTTP(test.runtime, test.dataset, test.alluxioValue)
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
		runtime           *datav1alpha1.AlluxioRuntime
		alluxioValue      *Alluxio
		expect            []string
		foundMountPathEnv bool
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{},
		}, &Alluxio{
			Properties: map[string]string{},
		}, []string{"-Xmx6G",
			"-XX:+UnlockExperimentalVMOptions"}, true},

		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{},
		}, &Alluxio{
			Properties: map[string]string{},
		}, []string{"-Xmx6G",
			"-XX:+UnlockExperimentalVMOptions"}, false},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{}
		engine.optimizeDefaultForMaster(test.runtime, test.alluxioValue)
		if test.alluxioValue.Master.JvmOptions[1] != test.expect[1] {
			t.Errorf("expected %v, got %v", test.expect, test.alluxioValue.JvmOptions)
		}

	}
}

// TestOptimizeDefaultForMasterWithValue tests the optimizeDefaultForMaster method of AlluxioEngine
// It verifies that when AlluxioRuntime specifies JVM options for Master, these options are correctly
// applied to the Alluxio Master configuration. The test case checks if the JvmOptions from the runtime
// spec are properly transferred to the Alluxio Master's JvmOptions. It compares the expected JVM options
// with the actual options set in the Alluxio Master configuration to ensure they match.

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

// TestOptimizeDefaultForWorkerNoValue tests the default behavior of the optimizeDefaultForWorker function when no Worker JVM options are set.
// This test case verifies whether the optimizeDefaultForWorker function correctly sets the default JVM options when the Worker JVM options in AlluxioRuntime and Alluxio are empty.
//
// Parameters:
//   - t *testing.T: The testing framework used to report errors and log test results.
//
// Return:
//   - None. The function asserts the expected JVM options against the actual values and reports errors if they do not match.
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

// TestOptimizeDefaultForFuseNoValue tests the optimizeDefaultFuse function of the AlluxioEngine.
// It verifies that the JVM options and fuse arguments are correctly set based on the provided AlluxioRuntime and Alluxio configurations.
// The test cases cover scenarios with both new and old fuse argument versions, and check if the expected JVM options and fuse arguments are applied correctly.
//
// Parameters:
//   - t *testing.T: The testing framework used to report errors and log test results.
//
// Returns:
//   - None. The function asserts the expected JVM options and fuse arguments against the actual values and reports errors if they do not match.
func TestOptimizeDefaultForFuseNoValue(t *testing.T) {
	var tests = []struct {
		runtime             *datav1alpha1.AlluxioRuntime
		alluxioValue        *Alluxio
		isNewFuseArgVersion bool
		expect              []string
		expectArgs          []string
		foundMountPathEnv   bool
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{},
		}, &Alluxio{
			Properties: map[string]string{},
			Fuse: Fuse{
				MountPath: "/mnt/runtime",
			},
		}, true, []string{"-Xmx16G",
			"-Xms16G",
			"-XX:+UseG1GC",
			"-XX:MaxDirectMemorySize=32g",
			"-XX:+UnlockExperimentalVMOptions"},
			[]string{"fuse", "--fuse-opts=kernel_cache,rw", "/mnt/runtime", "/"},
			false},
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{},
		}, &Alluxio{
			Properties: map[string]string{},
			Fuse: Fuse{
				MountPath: "/mnt/runtime",
			},
		}, false, []string{"-Xmx16G",
			"-Xms16G",
			"-XX:+UseG1GC",
			"-XX:MaxDirectMemorySize=32g",
			"-XX:+UnlockExperimentalVMOptions"},
			[]string{"fuse", "--fuse-opts=kernel_cache,rw"},
			true},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{}

		engine.optimizeDefaultFuse(test.runtime, test.alluxioValue, test.isNewFuseArgVersion)
		if !reflect.DeepEqual(test.alluxioValue.Fuse.JvmOptions, test.expect) {
			t.Errorf("expected %v, got %v", test.expect, test.alluxioValue.Fuse.JvmOptions)
		}

		if !reflect.DeepEqual(test.alluxioValue.Fuse.Args, test.expectArgs) {
			t.Errorf("expected fuse arg %v, got fuse arg %v", test.expectArgs, test.alluxioValue.Fuse.Args)
		}

		// _, found := test.alluxioValue.Fuse.Env["MOUNT_PATH"]
		// if found != test.foundMountPathEnv {
		// 	t.Errorf("expected fuse env %v, got fuse env %v", test.foundMountPathEnv, test.alluxioValue.Fuse.Env)
		// }
	}
}

// Tests the behavior of the optimizeDefaultFuse method under different configurations
// This test case verifies whether the JVM options for Alluxio Fuse are correctly optimized and merged
func TestOptimizeDefaultForFuseWithValue(t *testing.T) {
    // Defines the test case structure
    // runtime: Input Alluxio runtime configuration
    // alluxioValue: Input Alluxio configuration values
    // isNewFuseArgVersion: Indicates whether to use the new Fuse argument version
    // expect: Expected Fuse JVM options result
    var tests = []struct {
        runtime             *datav1alpha1.AlluxioRuntime
        alluxioValue        *Alluxio
        isNewFuseArgVersion bool
        expect              []string
    }{
        // Test case 1: Verifies that when JvmOptions are defined in runtime, they are correctly preserved after optimization
        {&datav1alpha1.AlluxioRuntime{
            Spec: datav1alpha1.AlluxioRuntimeSpec{
                Fuse: datav1alpha1.AlluxioFuseSpec{
                    JvmOptions: []string{"-Xmx4G"}, // Sets the maximum heap memory for Fuse JVM to 4G
                },
            },
        }, &Alluxio{
            Properties: map[string]string{}, // Empty property map
        }, true, []string{"-Xmx4G"}}, // Expects the optimized JVM options to remain unchanged
    }
    
    // Iterates through all test cases
    for _, test := range tests {
        // Creates a test engine instance
        engine := &AlluxioEngine{}
        
        // Executes the optimization method
        engine.optimizeDefaultFuse(test.runtime, test.alluxioValue, test.isNewFuseArgVersion)
        
        // Verifies if the optimization result matches the expectation
        if test.alluxioValue.Fuse.JvmOptions[0] != test.expect[0] {
            t.Errorf("expected %v, got %v", test.expect, test.alluxioValue.Fuse.JvmOptions)
        }
    }
}

func TestAlluxioEngine_setPortProperties(t *testing.T) {
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
						Resources: common.Resources{
							Requests: common.ResourceList{
								corev1.ResourceCPU:    "100m",
								corev1.ResourceMemory: "100Mi",
							},
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
			e.setPortProperties(tt.args.runtime, tt.args.alluxioValue)
			key := tt.args.alluxioValue.Properties["alluxio.master.rpc.port"]
			if key != strconv.Itoa(port) {
				t.Errorf("expected %d, got %s", port, tt.args.alluxioValue.Properties["alluxio.master.rpc.port"])
			}
		})
	}
}

func TestOptimizeDefaultForFuseWithArgs(t *testing.T) {
	var tests = []struct {
		runtime             *datav1alpha1.AlluxioRuntime
		alluxioValue        *Alluxio
		isNewFuseArgVersion bool
		expectArgs          []string
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Fuse: datav1alpha1.AlluxioFuseSpec{
					Args: []string{"fuse", "--fuse-opts=kernel_cache,rw,max_read=131072"},
				},
			},
		}, &Alluxio{
			Properties: map[string]string{},
			Fuse: Fuse{
				MountPath: "/mnt/runtime",
			},
		}, true, []string{"fuse", "--fuse-opts=kernel_cache,rw,max_read=131072", "/mnt/runtime", "/"}},
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Fuse: datav1alpha1.AlluxioFuseSpec{
					Args: []string{"fuse", "--fuse-opts=kernel_cache,rw,max_read=131072"},
				},
			},
		}, &Alluxio{
			Properties: map[string]string{},
		}, false, []string{"fuse", "--fuse-opts=kernel_cache,rw,max_read=131072"}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{}
		engine.optimizeDefaultFuse(test.runtime, test.alluxioValue, test.isNewFuseArgVersion)
		if !reflect.DeepEqual(test.alluxioValue.Fuse.Args, test.expectArgs) {
			t.Errorf("expected fuse arg %v, got fuse arg %v", test.expectArgs, test.alluxioValue.Fuse.Args)
		}
	}
}
