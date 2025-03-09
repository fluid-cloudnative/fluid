/*
Copyright 2024 The Fluid Authors.
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

package vineyard

import (
	"os"
	"reflect"
	"strings"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/net"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	LocalMount = datav1alpha1.Mount{
		MountPoint: "local:///mnt/test",
		Name:       "local",
	}
)

func TestTransformFuse(t *testing.T) {
	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	var tests = map[string]struct {
		runtime   *datav1alpha1.VineyardRuntime
		value     *Vineyard
		expect    []string
		expectEnv map[string]string
	}{
		"test image, imageTag and pullPolicy": {
			runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Fuse: datav1alpha1.VineyardClientSocketSpec{
						Image:           "dummy-fuse-image",
						ImageTag:        "dummy-tag",
						ImagePullPolicy: "IfNotPresent",
						Env:             map[string]string{"TEST_ENV": "true"},
						CleanPolicy:     "OnRuntimeDeleted",
					},
				},
			},
			value:     &Vineyard{},
			expect:    []string{"dummy-fuse-image", "dummy-tag", "IfNotPresent", "OnRuntimeDeleted"},
			expectEnv: map[string]string{"TEST_ENV": "true"},
		},
		"test image, imageTag and pullPolicy from env": {
			runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Fuse: datav1alpha1.VineyardClientSocketSpec{
						ImagePullPolicy: "IfNotPresent",
						Env:             map[string]string{"TEST_ENV": "true"},
						CleanPolicy:     "OnRuntimeDeleted",
					},
				},
			},
			value:     &Vineyard{},
			expect:    []string{"image-from-env", "image-tag-from-env", "IfNotPresent", "OnRuntimeDeleted"},
			expectEnv: map[string]string{"TEST_ENV": "true"},
		},
	}
	for k, test := range tests {
		if strings.Contains(k, "env") {
			os.Setenv("VINEYARD_FUSE_IMAGE_ENV", "image-from-env:image-tag-from-env")
		}
		runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "vineyard")
		if err != nil {
			t.Errorf("fail to create the runtimeInfo with error %v", err)
		}
		engine := &VineyardEngine{
			runtimeInfo: runtimeInfo,
		}
		engine.Log = ctrl.Log
		engine.transformFuse(test.runtime, test.value)

		if test.value.Fuse.Image != test.expect[0] {
			t.Errorf("Expected Fuse Image %s, got %s", test.expect[0], test.value.Fuse.Image)
		}
		if test.value.Fuse.ImageTag != test.expect[1] {
			t.Errorf("Expected Fuse ImageTag %s, got %s", test.expect[1], test.value.Fuse.ImageTag)
		}
		if test.value.Fuse.ImagePullPolicy != test.expect[2] {
			t.Errorf("Expected Fuse ImagePullPolicy %s, got %s", test.expect[2], test.value.Fuse.ImagePullPolicy)
		}
		if string(test.value.Fuse.CleanPolicy) != test.expect[3] {
			t.Errorf("Expected Fuse CleanPolicy %s, got %s", test.expect[3], test.value.Fuse.CleanPolicy)
		}
		if !reflect.DeepEqual(test.value.Fuse.Env, test.expectEnv) {
			t.Errorf("Expected Fuse Env %v, got %v", test.expectEnv, test.value.Fuse.Env)
		}
		if strings.Contains(k, "env") {
			os.Unsetenv("VINEYARD_FUSE_IMAGE_ENV")
		}
	}
}

func TestTransformMaster(t *testing.T) {
	testCases := map[string]struct {
		runtime   *datav1alpha1.VineyardRuntime
		wantValue *Vineyard
	}{
		"test image, imageTag and pullPolicy": {
			runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Master: datav1alpha1.MasterSpec{
						VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
							Image:           "test-image",
							ImageTag:        "test-tag",
							ImagePullPolicy: "IfNotPresent",
						},
					},
				},
			},
			wantValue: &Vineyard{
				Master: Master{
					Image:           "test-image",
					ImageTag:        "test-tag",
					ImagePullPolicy: "IfNotPresent",
				},
			},
		},
		"test image, imageTag and pullPolicy from env": {
			runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Master: datav1alpha1.MasterSpec{
						VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
							ImagePullPolicy: "IfNotPresent",
						},
					},
				},
			},
			wantValue: &Vineyard{
				Master: Master{
					Image:           "image-from-env",
					ImageTag:        "image-tag-from-env",
					ImagePullPolicy: "IfNotPresent",
				},
			},
		},
	}

	engine := &VineyardEngine{Log: fake.NullLogger()}
	ds := &datav1alpha1.Dataset{}
	for k, v := range testCases {
		if strings.Contains(k, "env") {
			os.Setenv("VINEYARD_MASTER_IMAGE_ENV", "image-from-env:image-tag-from-env")
		}
		gotValue := &Vineyard{}
		if err := engine.transformMasters(v.runtime, ds, gotValue); err == nil {
			if gotValue.Master.Image != v.wantValue.Master.Image {
				t.Errorf("check master %s failure, got:%s,want:%s",
					k,
					gotValue.Master.Image,
					v.wantValue.Master.Image,
				)
			}
			if gotValue.Master.ImageTag != v.wantValue.Master.ImageTag {
				t.Errorf("check master %s failure, got:%s,want:%s",
					k,
					gotValue.Master.ImageTag,
					v.wantValue.Master.ImageTag,
				)
			}
			if gotValue.Master.ImagePullPolicy != v.wantValue.Master.ImagePullPolicy {
				t.Errorf("check master %s failure, got:%s,want:%s",
					k,
					gotValue.Master.ImagePullPolicy,
					v.wantValue.Master.ImagePullPolicy,
				)
			}
		}
		if strings.Contains(k, "env") {
			os.Unsetenv("VINEYARD_MASTER_IMAGE_ENV")
		}
	}
}

func TestTransformWorker(t *testing.T) {
	testCases := map[string]struct {
		runtime   *datav1alpha1.VineyardRuntime
		wantValue *Vineyard
	}{
		"test replicas, image, imageTag and pullPolicy": {
			runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Replicas: 3,
					Worker: datav1alpha1.VineyardCompTemplateSpec{
						Replicas:        2,
						Image:           "test-image",
						ImageTag:        "test-tag",
						ImagePullPolicy: "IfNotPresent",
					},
				},
			},
			wantValue: &Vineyard{
				Worker: Worker{
					Image:           "test-image",
					ImageTag:        "test-tag",
					ImagePullPolicy: "IfNotPresent",
				},
			},
		},
		"test replicas, image, imageTag and pullPolicy from env": {
			runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Replicas: 3,
					Worker: datav1alpha1.VineyardCompTemplateSpec{
						Replicas:        2,
						Image:           "test-image",
						ImageTag:        "test-tag",
						ImagePullPolicy: "IfNotPresent",
					},
				},
			},
			wantValue: &Vineyard{
				Worker: Worker{
					Image:           "image-from-env",
					ImageTag:        "image-tag-from-env",
					ImagePullPolicy: "IfNotPresent",
				},
			},
		},
	}

	engine := &VineyardEngine{Log: fake.NullLogger()}
	for k, v := range testCases {
		if strings.Contains(k, "env") {
			os.Setenv("VINEYARD_WORKER_IMAGE_ENV", "image-from-env:image-tag-from-env")
		}
		gotValue := &Vineyard{}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "vineyard", base.WithTieredStore(v.runtime.Spec.TieredStore))
		if err := engine.transformWorkers(v.runtime, gotValue); err == nil {
			if gotValue.Worker.Replicas != v.wantValue.Worker.Replicas {
				t.Errorf("check %s failure, got:%d,want:%d",
					k,
					gotValue.Worker.Replicas,
					v.wantValue.Worker.Replicas,
				)
			}
			if gotValue.Worker.Image != v.wantValue.Worker.Image {
				t.Errorf("check %s failure, got:%s,want:%s",
					k,
					gotValue.Worker.Image,
					v.wantValue.Worker.Image,
				)
			}
			if gotValue.Worker.ImageTag != v.wantValue.Worker.ImageTag {
				t.Errorf("check %s failure, got:%s,want:%s",
					k,
					gotValue.Worker.ImageTag,
					v.wantValue.Worker.ImageTag,
				)
			}
			if gotValue.Worker.ImagePullPolicy != v.wantValue.Worker.ImagePullPolicy {
				t.Errorf("check %s failure, got:%s,want:%s",
					k,
					gotValue.Worker.ImagePullPolicy,
					v.wantValue.Worker.ImagePullPolicy,
				)
			}
		}
		if strings.Contains(k, "env") {
			os.Unsetenv("VINEYARD_WORKER_IMAGE_ENV")
		}
	}
}

func TestTransformMasterSelector(t *testing.T) {
	tests := []struct {
		name     string
		runtime  *datav1alpha1.VineyardRuntime
		expected map[string]string
	}{
		{
			name: "NoNodeSelector",
			runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Master: datav1alpha1.MasterSpec{},
				},
			},
			expected: map[string]string{},
		},
		{
			name: "WithNodeSelector",
			runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Master: datav1alpha1.MasterSpec{
						VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
							NodeSelector: map[string]string{"disktype": "ssd"},
						},
					},
				},
			},
			expected: map[string]string{"disktype": "ssd"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &VineyardEngine{}
			actual := engine.transformMasterSelector(tt.runtime)
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, actual)
			}
		})
	}
}

func TestTransformMasterPorts(t *testing.T) {
	tests := []struct {
		name     string
		runtime  *datav1alpha1.VineyardRuntime
		expected map[string]int
	}{
		{
			name: "NoMasterPorts",
			runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Master: datav1alpha1.MasterSpec{
						VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
							Ports: map[string]int{},
						},
					},
				},
			},
			expected: map[string]int{
				"client": 2379,
				"peer":   2380,
			},
		},
		{
			name: "WithMasterPorts",
			runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Master: datav1alpha1.MasterSpec{
						VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
							Ports: map[string]int{
								"client": 1234,
								"peer":   5678,
							},
						},
					},
				},
			},
			expected: map[string]int{
				"client": 1234,
				"peer":   5678,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &VineyardEngine{}
			actual := engine.transformMasterPorts(tt.runtime)
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, actual)
			}
		})
	}
}

func TestTransformMasterOptions(t *testing.T) {
	tests := []struct {
		name     string
		runtime  *datav1alpha1.VineyardRuntime
		expected map[string]string
	}{
		{
			name: "NoMasterOptions",
			runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Master: datav1alpha1.MasterSpec{
						VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
							Options: map[string]string{},
						},
					},
				},
			},
			expected: map[string]string{
				"vineyardd.reserve.memory": "true",
				"etcd.prefix":              "/vineyard",
			},
		},
		{
			name: "WithMasterOptions",
			runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Master: datav1alpha1.MasterSpec{
						VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
							Options: map[string]string{
								"vineyardd.reserve.memory": "false",
								"etcd.prefix":              "/vineyard-test",
							},
						},
					},
				},
			},
			expected: map[string]string{
				"vineyardd.reserve.memory": "false",
				"etcd.prefix":              "/vineyard-test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &VineyardEngine{}
			actual := engine.transformMasterOptions(tt.runtime)
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, actual)
			}
		})
	}
}

func TestTransformWorkerOptions(t *testing.T) {
	tests := []struct {
		name     string
		runtime  *datav1alpha1.VineyardRuntime
		expected map[string]string
	}{
		{
			name: "NoWorkerOptions",
			runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Worker: datav1alpha1.VineyardCompTemplateSpec{
						Options: map[string]string{},
					},
				},
			},
			expected: map[string]string{},
		},
		{
			name: "WithWorkerOptions",
			runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Worker: datav1alpha1.VineyardCompTemplateSpec{
						Options: map[string]string{
							"dummy-key": "dummy-value",
						},
					},
				},
			},
			expected: map[string]string{
				"dummy-key": "dummy-value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &VineyardEngine{}
			actual := engine.transformWorkerOptions(tt.runtime)
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, actual)
			}
		})
	}
}

func TestTransformFuseOptions(t *testing.T) {
	tests := []struct {
		name     string
		runtime  *datav1alpha1.VineyardRuntime
		value    *Vineyard
		expected map[string]string
	}{
		{
			name: "NoFuseOptions",
			runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Fuse: datav1alpha1.VineyardClientSocketSpec{
						Options: map[string]string{},
					},
				},
			},
			value: &Vineyard{
				FullnameOverride: "vineyard",
				Master: Master{
					Ports: map[string]int{
						"client": 2379,
					},
				},
			},
			expected: map[string]string{
				"size":          "0",
				"etcd_endpoint": "http://vineyard-master-0.vineyard-master.default:2379",
				"etcd_prefix":   "/vineyard",
			},
		},
		{
			name: "WithFuseOptions",
			runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Fuse: datav1alpha1.VineyardClientSocketSpec{
						Options: map[string]string{
							"size":           "10Gi",
							"etcd_endpoint":  "http://vineyard-master-0.vineyard-master.default:12379",
							"reserve_memory": "true",
						},
					},
				},
			},
			value: &Vineyard{
				Master: Master{
					Ports: map[string]int{
						"client": 12379,
					},
				},
			},
			expected: map[string]string{
				"size":           "10Gi",
				"etcd_endpoint":  "http://vineyard-master-0.vineyard-master.default:12379",
				"etcd_prefix":    "/vineyard",
				"reserve_memory": "true",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &VineyardEngine{
				namespace: "default",
			}
			actual := engine.transformFuseOptions(tt.runtime, tt.value)
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, actual)
			}
		})
	}
}

func TestTransformWorkerPorts(t *testing.T) {
	tests := []struct {
		name     string
		runtime  *datav1alpha1.VineyardRuntime
		expected map[string]int
	}{
		{
			name: "NoWorkerPorts",
			runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Worker: datav1alpha1.VineyardCompTemplateSpec{
						Ports: map[string]int{},
					},
				},
			},
			expected: map[string]int{
				"rpc":      9600,
				"exporter": 9144,
			},
		},
		{
			name: "WithWorkerPorts",
			runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Worker: datav1alpha1.VineyardCompTemplateSpec{
						Ports: map[string]int{
							"rpc":      1234,
							"exporter": 5678,
						},
					},
				},
			},
			expected: map[string]int{
				"rpc":      1234,
				"exporter": 5678,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &VineyardEngine{}
			actual := engine.transformWorkerPorts(tt.runtime)
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, actual)
			}
		})
	}
}

func TestTransformFuseNodeSelector(t *testing.T) {
	tests := []struct {
		name     string
		runtime  *datav1alpha1.VineyardRuntime
		expected map[string]string
	}{
		{
			name: "NoWorkerPorts",
			runtime: &datav1alpha1.VineyardRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vineyard",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Worker: datav1alpha1.VineyardCompTemplateSpec{
						NodeSelector: map[string]string{},
					},
				},
			},
			expected: map[string]string{
				"fluid.io/f-fluid-vineyard": "true",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "vineyard")
			if err != nil {
				t.Errorf("fail to create the runtimeInfo with error %v", err)
			}
			engine := &VineyardEngine{
				name:        "vineyard",
				namespace:   "fluid",
				runtimeInfo: runtimeInfo,
			}
			actual := engine.transformFuseNodeSelector(tt.runtime)
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, actual)
			}
		})
	}
}

func TestTransformFuseWithLaunchMode(t *testing.T) {
	testCases := map[string]struct {
		runtime   *datav1alpha1.VineyardRuntime
		wantValue *Vineyard
	}{
		"test fuse launch mode case 1": {
			runtime: &datav1alpha1.VineyardRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Fuse: datav1alpha1.VineyardClientSocketSpec{
						LaunchMode: datav1alpha1.EagerMode,
					},
				},
			},
			wantValue: &Vineyard{
				Fuse: Fuse{
					NodeSelector: map[string]string{},
				},
			},
		},
		"test fuse launch mode case 2": {
			runtime: &datav1alpha1.VineyardRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Fuse: datav1alpha1.VineyardClientSocketSpec{
						LaunchMode: datav1alpha1.LazyMode,
					},
				},
			},
			wantValue: &Vineyard{
				Fuse: Fuse{
					NodeSelector: map[string]string{
						utils.GetFuseLabelName("fluid", "hbase", ""): "true",
					},
				},
			},
		},
		"test fuse launch mode case 3": {
			runtime: &datav1alpha1.VineyardRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Fuse: datav1alpha1.VineyardClientSocketSpec{
						LaunchMode: "",
					},
				},
			},
			wantValue: &Vineyard{
				Fuse: Fuse{
					NodeSelector: map[string]string{
						utils.GetFuseLabelName("fluid", "hbase", ""): "true",
					},
				},
			},
		},
	}

	runtimeInfo, err := base.BuildRuntimeInfo("hbase", "fluid", "Vineyard")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	engine := &VineyardEngine{
		Log:         fake.NullLogger(),
		runtimeInfo: runtimeInfo,
		Client:      fake.NewFakeClientWithScheme(testScheme),
	}

	for k, v := range testCases {
		gotValue := &Vineyard{}
		engine.transformFuse(v.runtime, gotValue)
		if !reflect.DeepEqual(gotValue.Fuse.NodeSelector, v.wantValue.Fuse.NodeSelector) {
			t.Errorf("check %s failure, got:%+v,want:%+v",
				k,
				gotValue.Fuse.NodeSelector,
				v.wantValue.Fuse.NodeSelector,
			)
		}
	}
}

func TestTransformTieredStore(t *testing.T) {
	defaultQuota := resource.MustParse("4Gi")
	quota := resource.MustParse("20Gi")
	tests := []struct {
		name     string
		runtime  *datav1alpha1.VineyardRuntime
		expected TieredStore
	}{
		{
			name: "Notieredstore",
			runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					TieredStore: datav1alpha1.TieredStore{},
				},
			},
			expected: TieredStore{
				Levels: []Level{
					{
						MediumType: "MEM",
						Level:      0,
						Quota:      &defaultQuota,
					},
				},
			},
		},
		{
			name: "Withtieredstore",
			runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					TieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{
							{
								MediumType: "MEM",
								Quota:      &quota,
							},
						},
					},
				},
			},
			expected: TieredStore{
				Levels: []Level{
					{
						MediumType: "MEM",
						Level:      0,
						Quota:      &quota,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &VineyardEngine{}
			actual, err := engine.transformTieredStore(tt.runtime)
			if tt.name == "Notieredstore" {
				if err == nil {
					t.Errorf("expected error, get nil")
				}
			} else {
				if !reflect.DeepEqual(actual, tt.expected) {
					t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, actual)
				}
			}
		})
	}
}

func TestVineyardEngineAllocatePorts(t *testing.T) {
	pr := net.ParsePortRangeOrDie("14000-16000")
	dummyPorts := func(client client.Client) (ports []int, err error) {
		return []int{14000, 14001, 14002, 14003}, nil
	}
	err := portallocator.SetupRuntimePortAllocator(nil, pr, "bitmap", dummyPorts)
	if err != nil {
		t.Fatal(err.Error())
	}
	type args struct {
		runtime *datav1alpha1.VineyardRuntime
		value   *Vineyard
	}
	tests := []struct {
		name            string
		args            args
		wantErr         bool
		wantMasterPorts map[string]int
		wantWorkerPorts map[string]int
	}{
		{
			name: "Expect not to allocate ports",
			args: args{
				runtime: &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Master: datav1alpha1.MasterSpec{
							VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
								NetworkMode: "ContainerNetwork",
							},
						},
						Worker: datav1alpha1.VineyardCompTemplateSpec{
							NetworkMode: "ContainerNetwork",
						},
					},
				},
				value: &Vineyard{
					Master: Master{
						Ports: map[string]int{
							MasterClientName: MasterClientPort,
							MasterPeerName:   MasterPeerPort,
						},
					},
					Worker: Worker{
						Ports: map[string]int{
							WorkerRPCName:      WorkerRPCPort,
							WorkerExporterName: WorkerExporterPort,
						},
					},
				},
			},
			wantErr: false,
			wantMasterPorts: map[string]int{
				MasterClientName: MasterClientPort,
				MasterPeerName:   MasterPeerPort,
			},
			wantWorkerPorts: map[string]int{
				WorkerRPCName:      WorkerRPCPort,
				WorkerExporterName: WorkerExporterPort,
			},
		},
		{
			name: "Expect to allocate ports",
			args: args{
				runtime: &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Master: datav1alpha1.MasterSpec{
							VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
								NetworkMode: "HostNetwork",
							},
						},
						Worker: datav1alpha1.VineyardCompTemplateSpec{
							NetworkMode: "HostNetwork",
						},
					},
				},
				value: &Vineyard{
					Master: Master{
						Ports: map[string]int{
							MasterClientName: MasterClientPort,
							MasterPeerName:   MasterPeerPort,
						},
					},
					Worker: Worker{
						Ports: map[string]int{
							WorkerRPCName:      WorkerRPCPort,
							WorkerExporterName: WorkerExporterPort,
						},
					},
				},
			},
			wantErr: false,
			wantMasterPorts: map[string]int{
				MasterClientName: MasterClientPort,
				MasterPeerName:   MasterPeerPort,
			},
			wantWorkerPorts: map[string]int{
				WorkerRPCName:      WorkerRPCPort,
				WorkerExporterName: WorkerExporterPort,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &VineyardEngine{}
			if err := v.allocatePorts(tt.args.value, tt.args.runtime); (err != nil) != tt.wantErr {
				t.Errorf("allocatePorts() error = %v, wantErr %v", err, tt.wantErr)
			}
			if strings.Contains(tt.name, "Expect not to allocate ports") {
				if !reflect.DeepEqual(tt.wantMasterPorts, tt.args.value.Master.Ports) {
					t.Errorf("allocatePorts() got master ports = %v, want %v", tt.args.value.Master.Ports, tt.wantMasterPorts)
				}
				if !reflect.DeepEqual(tt.wantWorkerPorts, tt.args.value.Worker.Ports) {
					t.Errorf("allocatePorts() got worker ports = %v, want %v", tt.args.value.Worker.Ports, tt.wantWorkerPorts)
				}
			} else {
				if len(tt.args.value.Master.Ports) != 2 {
					t.Errorf("allocatePorts() got master ports = %v, want 2", len(tt.args.value.Master.Ports))
				}
				if len(tt.args.value.Worker.Ports) != 2 {
					t.Errorf("allocatePorts() got worker ports = %v, want 2", len(tt.args.value.Worker.Ports))
				}
				if tt.args.value.Master.Ports[MasterClientName] < 14000 || tt.args.value.Master.Ports[MasterClientName] > 16000 {
					t.Errorf("allocatePorts() got master client port = %v, want between 14000 and 16000", tt.args.value.Master.Ports[MasterClientName])
				}
				if tt.args.value.Master.Ports[MasterPeerName] < 14000 || tt.args.value.Master.Ports[MasterPeerName] > 16000 {
					t.Errorf("allocatePorts() got master peer port = %v, want between 14000 and 16000", tt.args.value.Master.Ports[MasterPeerName])
				}
				if tt.args.value.Worker.Ports[WorkerRPCName] < 14000 || tt.args.value.Worker.Ports[WorkerRPCName] > 16000 {
					t.Errorf("allocatePorts() got worker rpc port = %v, want between 14000 and 16000", tt.args.value.Worker.Ports[WorkerRPCName])
				}
				if tt.args.value.Worker.Ports[WorkerExporterName] < 14000 || tt.args.value.Worker.Ports[WorkerExporterName] > 16000 {
					t.Errorf("allocatePorts() got worker exporter port = %v, want between 14000 and 16000", tt.args.value.Worker.Ports[WorkerExporterName])
				}
			}
		})
	}
}

func TestVineyardEngineTransformPodMetadata(t *testing.T) {
	engine := &VineyardEngine{Log: fake.NullLogger()}

	type testCase struct {
		Name    string
		Runtime *datav1alpha1.VineyardRuntime
		Value   *Vineyard

		wantValue *Vineyard
	}

	testCases := []testCase{
		{
			Name: "set_common_labels_and_annotations",
			Runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					PodMetadata: datav1alpha1.PodMetadata{
						Labels:      map[string]string{"common-key": "common-value"},
						Annotations: map[string]string{"common-annotation": "val"},
					},
				},
			},
			Value: &Vineyard{},
			wantValue: &Vineyard{
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
			Runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					PodMetadata: datav1alpha1.PodMetadata{
						Labels:      map[string]string{"common-key": "common-value"},
						Annotations: map[string]string{"common-annotation": "val"},
					},
					Master: datav1alpha1.MasterSpec{
						VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
							PodMetadata: datav1alpha1.PodMetadata{
								Labels:      map[string]string{"common-key": "master-value"},
								Annotations: map[string]string{"common-annotation": "master-val"},
							},
						},
					},
					Worker: datav1alpha1.VineyardCompTemplateSpec{
						PodMetadata: datav1alpha1.PodMetadata{
							Labels:      map[string]string{"common-key": "worker-value"},
							Annotations: map[string]string{"common-annotation": "worker-val"},
						},
					},
				},
			},
			Value: &Vineyard{},
			wantValue: &Vineyard{
				Master: Master{
					Labels:      map[string]string{"common-key": "master-value"},
					Annotations: map[string]string{"common-annotation": "master-val"},
				},
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
		engine.transformPodMetadata(tt.Runtime, tt.Value)

		if !reflect.DeepEqual(tt.Value, tt.wantValue) {
			t.Fatalf("test name: %s. Expect value %v, but got %v", tt.Name, tt.wantValue, tt.Value)
		}
	}
}
