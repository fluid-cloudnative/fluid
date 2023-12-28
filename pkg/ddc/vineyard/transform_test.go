/*
Copyright 2023 The Fluid Authors.

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
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"k8s.io/apimachinery/pkg/api/resource"
	ctrl "sigs.k8s.io/controller-runtime"
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

	var tests = []struct {
		runtime *datav1alpha1.VineyardRuntime
		value   *Vineyard
		expect  []string
	}{
		{&datav1alpha1.VineyardRuntime{
			Spec: datav1alpha1.VineyardRuntimeSpec{
				Fuse: datav1alpha1.VineyardSockSpec{
					Image:           "dummy-fuse-image",
					ImageTag:        "dummy-tag",
					ImagePullPolicy: "IfNotPresent",
					CleanPolicy:     "OnRuntimeDeleted",
				},
			},
		}, &Vineyard{}, []string{"dummy-fuse-image", "dummy-tag", "IfNotPresent", "OnRuntimeDeleted"}},
	}
	for _, test := range tests {
		engine := &VineyardEngine{}
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
	}

	engine := &VineyardEngine{Log: fake.NullLogger()}
	ds := &datav1alpha1.Dataset{}
	for k, v := range testCases {
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
	}
}

func TestTransformWorker(t *testing.T) {
	testCases := map[string]struct {
		runtime   *datav1alpha1.VineyardRuntime
		wantValue *Vineyard
	}{
		"test image, imageTag and pullPolicy": {
			runtime: &datav1alpha1.VineyardRuntime{
				Spec: datav1alpha1.VineyardRuntimeSpec{
					Worker: datav1alpha1.VineyardCompTemplateSpec{
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
	}

	engine := &VineyardEngine{Log: fake.NullLogger()}
	for k, v := range testCases {
		gotValue := &Vineyard{}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "vineyard", v.runtime.Spec.TieredStore)
		if err := engine.transformWorkers(v.runtime, gotValue); err == nil {
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
			engine := &VineyardEngine{
				name:      "vineyard",
				namespace: "fluid",
			}
			actual := engine.transformFuseNodeSelector(tt.runtime)
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, actual)
			}
		})
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
			actual := engine.transformTieredStore(tt.runtime)
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, actual)
			}
		})
	}
}
