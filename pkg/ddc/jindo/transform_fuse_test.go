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

package jindo

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
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     string
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Jindo{}, "true"},
	}
	for _, test := range tests {
		engine := &JindoEngine{Log: fake.NullLogger()}
		err := engine.transformFuse(test.runtime, test.jindoValue)
		if err != nil {
			t.Errorf("Got err %v", err)
		}
		if test.jindoValue.Fuse.FuseProperties["jfs.cache.data-cache.enable"] != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Fuse.FuseProperties["jfs.cache.data-cache.enable"])
		}
	}
}

func TestTransformFuseNodeSelectorWithLaunchMode(t *testing.T) {
	testCases := map[string]struct {
		runtime   *datav1alpha1.JindoRuntime
		wantValue *Jindo
	}{
		"test fuse launch mode case 1": {
			runtime: &datav1alpha1.JindoRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.JindoRuntimeSpec{
					Fuse: datav1alpha1.JindoFuseSpec{
						LaunchMode: datav1alpha1.EagerMode,
						NodeSelector: map[string]string{
							"fuse_node": "true",
						},
					},
				},
			},
			wantValue: &Jindo{
				Fuse: Fuse{
					NodeSelector: map[string]string{
						"fuse_node": "true",
					},
				},
			},
		},
		"test fuse launch mode case 2": {
			runtime: &datav1alpha1.JindoRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.JindoRuntimeSpec{
					Fuse: datav1alpha1.JindoFuseSpec{
						LaunchMode: datav1alpha1.LazyMode,
						NodeSelector: map[string]string{
							"fuse_node": "true",
						},
					},
				},
			},
			wantValue: &Jindo{
				Fuse: Fuse{
					NodeSelector: map[string]string{
						utils.GetFuseLabelName("fluid", "hbase", ""): "true",
						"fuse_node": "true",
					},
				},
			},
		},
		"test fuse launch mode case 3": {
			runtime: &datav1alpha1.JindoRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.JindoRuntimeSpec{
					Fuse: datav1alpha1.JindoFuseSpec{
						LaunchMode: "",
					},
				},
			},
			wantValue: &Jindo{
				Fuse: Fuse{
					NodeSelector: map[string]string{
						utils.GetFuseLabelName("fluid", "hbase", ""): "true",
					},
				},
			},
		},
	}

	runtimeInfo, err := base.BuildRuntimeInfo("hbase", "fluid", "Jindo")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	engine := &JindoEngine{
		Log:         fake.NullLogger(),
		runtimeInfo: runtimeInfo,
		Client:      fake.NewFakeClientWithScheme(testScheme),
	}

	for k, v := range testCases {
		gotValue := &Jindo{}
		if err := engine.transformFuseNodeSelector(v.runtime, gotValue); err == nil {
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

func TestTransformRunAsUser(t *testing.T) {
	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     string
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				User: "user",
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Jindo{}, "user"},
	}
	for _, test := range tests {
		engine := &JindoEngine{Log: fake.NullLogger()}
		err := engine.transformRunAsUser(test.runtime, test.jindoValue)
		if err != nil {
			t.Errorf("Got err %v", err)
		}
		if test.jindoValue.Fuse.RunAs != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Fuse.RunAs)
		}
	}
}

func TestTransformSecret(t *testing.T) {
	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     string
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Secret: "secret",
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Jindo{}, "secret"},
	}
	for _, test := range tests {
		engine := &JindoEngine{Log: fake.NullLogger()}
		err := engine.transformSecret(test.runtime, test.jindoValue)
		if err != nil {
			t.Errorf("Got err %v", err)
		}
		if test.jindoValue.Secret != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Fuse.RunAs)
		}
	}
}

func TestTransformFuseArg(t *testing.T) {
	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     string
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Secret: "secret",
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
					Path:       "/",
				}},
			}}, &Jindo{}, "-ocredential_provider=secrets:///token/ -oroot_ns=jindo -okernel_cache"},
	}
	for _, test := range tests {
		engine := &JindoEngine{Log: fake.NullLogger()}
		properties := engine.transformFuseArg(test.runtime, test.dataset)
		if properties[0] != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Fuse.RunAs)
		}
	}
}

func TestParseFuseImage(t *testing.T) {
	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     string
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Secret: "secret",
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
					Path:       "/",
				}},
			}}, &Jindo{}, "registry.cn-shanghai.aliyuncs.com/jindofs/jindo-fuse:3.8.0"},
	}
	for _, test := range tests {
		engine := &JindoEngine{Log: fake.NullLogger()}
		imageR, tagR := engine.parseFuseImage()
		registryVersion := imageR + ":" + tagR
		if registryVersion != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Fuse.RunAs)
		}
	}
}
