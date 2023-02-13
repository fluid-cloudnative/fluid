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
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
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
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
				Owner: &datav1alpha1.User{
					UID: &x,
					GID: &x,
				},
			},
		}, &Alluxio{}, []string{"fuse", "--fuse-opts=kernel_cache,rw,max_read=131072,uid=1000,gid=1000,allow_other"}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{}
		engine.Log = ctrl.Log
		err := engine.transformFuse(test.runtime, test.dataset, test.value)
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
	}
}

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
				},
			},
			wantValue: &Alluxio{
				Worker: Worker{
					HostNetwork: true,
				},
			},
		},
	}

	engine := &AlluxioEngine{Log: fake.NullLogger()}
	for k, v := range testCases {
		gotValue := &Alluxio{}
		if err := engine.transformWorkers(v.runtime, gotValue); err == nil {
			if gotValue.Worker.HostNetwork != v.wantValue.Worker.HostNetwork {
				t.Errorf("check %s failure, got:%t,want:%t",
					k,
					gotValue.Worker.HostNetwork,
					v.wantValue.Worker.HostNetwork,
				)
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
