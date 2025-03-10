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
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTransformFuseWithNoArgs(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		dataset      *datav1alpha1.Dataset
		goosefsValue *GooseFS
		expect       string
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &GooseFS{}, "--fuse-opts=rw,direct_io,allow_other"},
	}
	for _, test := range tests {
		runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "goosefs")
		if err != nil {
			t.Errorf("fail to create the runtimeInfo with error %v", err)
		}
		engine := &GooseFSEngine{
			Log:         fake.NullLogger(),
			runtimeInfo: runtimeInfo,
			Client:      fake.NewFakeClientWithScheme(testScheme),
		}
		err = engine.transformFuse(test.runtime, test.dataset, test.goosefsValue)
		if err != nil {
			t.Errorf("Got err %v", err)
		}
		if test.goosefsValue.Fuse.Args[1] != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.goosefsValue.Fuse.Args[1])
		}
	}
}

func TestTransformFuseWithArgs(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		dataset      *datav1alpha1.Dataset
		goosefsValue *GooseFS
		expect       string
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Fuse: datav1alpha1.GooseFSFuseSpec{
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
			}}, &GooseFS{}, "--fuse-opts=kernel_cache,allow_other"},
	}
	for _, test := range tests {
		runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "goosefs")
		if err != nil {
			t.Errorf("fail to create the runtimeInfo with error %v", err)
		}
		engine := &GooseFSEngine{
			Log:         fake.NullLogger(),
			runtimeInfo: runtimeInfo,
			Client:      fake.NewFakeClientWithScheme(testScheme),
		}
		err = engine.transformFuse(test.runtime, test.dataset, test.goosefsValue)
		if err != nil {
			t.Errorf("Got err %v", err)
		}
		if test.goosefsValue.Fuse.Args[1] != test.expect {
			t.Errorf("expected fuse %v, but got %v", test.expect, test.goosefsValue.Fuse.Args[1])
		}
	}
}

func TestTransformFuseWithLaunchMode(t *testing.T) {
	testCases := map[string]struct {
		runtime   *datav1alpha1.GooseFSRuntime
		wantValue *GooseFS
	}{
		"test fuse launch mode case 1": {
			runtime: &datav1alpha1.GooseFSRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					Fuse: datav1alpha1.GooseFSFuseSpec{
						LaunchMode: datav1alpha1.EagerMode,
						NodeSelector: map[string]string{
							"fuse_node": "true",
						},
					},
				},
			},
			wantValue: &GooseFS{
				Fuse: Fuse{
					NodeSelector: map[string]string{
						"fuse_node": "true",
					},
				},
			},
		},
		"test fuse launch mode case 2": {
			runtime: &datav1alpha1.GooseFSRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					Fuse: datav1alpha1.GooseFSFuseSpec{
						LaunchMode: datav1alpha1.LazyMode,
						NodeSelector: map[string]string{
							"fuse_node": "true",
						},
					},
				},
			},
			wantValue: &GooseFS{
				Fuse: Fuse{
					NodeSelector: map[string]string{
						utils.GetFuseLabelName("fluid", "hbase", ""): "true",
						"fuse_node": "true",
					},
				},
			},
		},
		"test fuse launch mode case 3": {
			runtime: &datav1alpha1.GooseFSRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					Fuse: datav1alpha1.GooseFSFuseSpec{
						LaunchMode: "",
					},
				},
			},
			wantValue: &GooseFS{
				Fuse: Fuse{
					NodeSelector: map[string]string{
						utils.GetFuseLabelName("fluid", "hbase", ""): "true",
					},
				},
			},
		},
	}

	runtimeInfo, err := base.BuildRuntimeInfo("hbase", "fluid", "GooseFS")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	engine := &GooseFSEngine{
		Log:         fake.NullLogger(),
		runtimeInfo: runtimeInfo,
		Client:      fake.NewFakeClientWithScheme(testScheme),
	}
	ds := &datav1alpha1.Dataset{}
	for k, v := range testCases {
		gotValue := &GooseFS{}
		if err := engine.transformFuse(v.runtime, ds, gotValue); err == nil {
			if !reflect.DeepEqual(gotValue.Fuse.NodeSelector, v.wantValue.Fuse.NodeSelector) {
				t.Errorf("check %s failure, got:%+v,want:%+v",
					k,
					gotValue.Fuse.NodeSelector,
					v.wantValue.Fuse.NodeSelector,
				)
			}
		}

	}
}
