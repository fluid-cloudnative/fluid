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

package alluxio

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestTransformFuseWithNoArgs(t *testing.T) {
	var tests = []struct {
		runtime           *datav1alpha1.AlluxioRuntime
		dataset           *datav1alpha1.Dataset
		alluxioValue      *Alluxio
		expect            []string
		foundMountPathEnv bool
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Alluxio{}, []string{"fuse", "--fuse-opts=kernel_cache,rw,allow_other"}, true},
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Fuse: datav1alpha1.AlluxioFuseSpec{
					ImageTag: "release-2.8.1-SNAPSHOT-0433ade",
				},
			},
		}, &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Alluxio{}, []string{"fuse", "--fuse-opts=kernel_cache,rw,allow_other", "/alluxio/default/test/alluxio-fuse", "/"}, false},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{
			name:      "test",
			namespace: "default",
			Log:       fake.NullLogger()}
		err := engine.transformFuse(test.runtime, test.dataset, test.alluxioValue)
		if err != nil {
			t.Errorf("Got err %v", err)
		}
		if !reflect.DeepEqual(test.alluxioValue.Fuse.Args, test.expect) {
			t.Errorf("expected value %v, but got %v", test.expect, test.alluxioValue.Fuse.Args)
		}

		_, found := test.alluxioValue.Fuse.Env["MOUNT_POINT"]
		if found != test.foundMountPathEnv {
			t.Errorf("expected fuse env %v, got fuse env %v", test.foundMountPathEnv, test.alluxioValue.Fuse.Env)
		}
	}
}

func TestTransformFuseWithArgs(t *testing.T) {
	var tests = []struct {
		runtime           *datav1alpha1.AlluxioRuntime
		dataset           *datav1alpha1.Dataset
		alluxioValue      *Alluxio
		expect            []string
		foundMountPathEnv bool
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Fuse: datav1alpha1.AlluxioFuseSpec{
					Args: []string{
						"fuse",
						"--fuse-opts=kernel_cache",
					},
				},
			},
		}, &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Alluxio{}, []string{"fuse", "--fuse-opts=kernel_cache,allow_other"}, true},
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Fuse: datav1alpha1.AlluxioFuseSpec{
					ImageTag: "release-2.8.1-SNAPSHOT-0433ade",
					Args: []string{
						"fuse",
						"--fuse-opts=kernel_cache",
					},
				},
			},
		}, &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Alluxio{}, []string{"fuse", "--fuse-opts=kernel_cache,allow_other", "/alluxio/default/test/alluxio-fuse", "/"}, false},
		{&datav1alpha1.AlluxioRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Fuse: datav1alpha1.AlluxioFuseSpec{
					ImageTag: "v2.8.0",
					Args: []string{
						"fuse",
						"--fuse-opts=kernel_cache",
					},
				},
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Alluxio{}, []string{"fuse", "--fuse-opts=kernel_cache,allow_other", "/alluxio/default/test/alluxio-fuse", "/"}, false},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{
			name:      "test",
			namespace: "default",
			Log:       fake.NullLogger()}
		err := engine.transformFuse(test.runtime, test.dataset, test.alluxioValue)
		if err != nil {
			t.Errorf("Got err %v", err)
		}
		if !reflect.DeepEqual(test.alluxioValue.Fuse.Args, test.expect) {
			t.Errorf("expected value %v, but got %v", test.expect, test.alluxioValue.Fuse.Args)
		}

		_, found := test.alluxioValue.Fuse.Env["MOUNT_POINT"]
		if found != test.foundMountPathEnv {
			t.Errorf("expected fuse env %v, got fuse env %v", test.foundMountPathEnv, test.alluxioValue.Fuse.Env)
		}
	}
}

func TestTransformFuseWithNetwork(t *testing.T) {
	testCases := map[string]struct {
		runtime   *datav1alpha1.AlluxioRuntime
		wantValue *Alluxio
	}{
		"test network mode case 1": {
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Fuse: datav1alpha1.AlluxioFuseSpec{
						ImageTag:        "2.8.0",
						Image:           "fluid/alluixo-fuse",
						ImagePullPolicy: "always",
						NetworkMode:     datav1alpha1.ContainerNetworkMode,
					},
				},
			},
			wantValue: &Alluxio{
				Fuse: Fuse{
					HostNetwork: false,
				},
			},
		},
		"test network mode case 2": {
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Fuse: datav1alpha1.AlluxioFuseSpec{
						ImageTag:        "2.8.0",
						Image:           "fluid/alluixo-fuse",
						ImagePullPolicy: "always",
						NetworkMode:     datav1alpha1.HostNetworkMode,
					},
				},
			},
			wantValue: &Alluxio{
				Fuse: Fuse{
					HostNetwork: true,
				},
			},
		},
		"test network mode case 3": {
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Fuse: datav1alpha1.AlluxioFuseSpec{
						ImageTag:        "2.8.0",
						Image:           "fluid/alluixo-fuse",
						ImagePullPolicy: "always",
						NetworkMode:     "",
					},
				},
			},
			wantValue: &Alluxio{
				Fuse: Fuse{
					HostNetwork: true,
				},
			},
		},
	}

	engine := &AlluxioEngine{Log: fake.NullLogger()}
	ds := &datav1alpha1.Dataset{}
	for k, v := range testCases {
		gotValue := &Alluxio{}
		if err := engine.transformFuse(v.runtime, ds, gotValue); err == nil {
			if gotValue.Fuse.HostNetwork != v.wantValue.Fuse.HostNetwork {
				t.Errorf("check %s failure, got:%t,want:%t",
					k,
					gotValue.Fuse.HostNetwork,
					v.wantValue.Fuse.HostNetwork,
				)
			}
		}

	}
}
