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

package jindofsx

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
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
		engine := &JindoFSxEngine{Log: fake.NullLogger()}
		engine.transformFuse(test.runtime, test.jindoValue)
		if test.jindoValue.Fuse.FuseProperties["fs.jindofsx.data.cache.enable"] != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Fuse.FuseProperties["jfs.cache.data-cache.enable"])
		}
	}
}

func TestTransformFuseWithSecret(t *testing.T) {
	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     string
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Secret: "test",
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Jindo{}, "JSON"},
	}
	for _, test := range tests {
		engine := &JindoFSxEngine{Log: fake.NullLogger()}
		engine.transformFuse(test.runtime, test.jindoValue)
		if test.jindoValue.Fuse.FuseProperties["fs.oss.provider.format"] != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Fuse.FuseProperties["fs.oss.provider.format"])
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
		engine := &JindoFSxEngine{Log: fake.NullLogger()}
		engine.transformRunAsUser(test.runtime, test.jindoValue)
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
		engine := &JindoFSxEngine{Log: fake.NullLogger()}
		engine.transformSecret(test.runtime, test.jindoValue)
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
				Fuse: datav1alpha1.JindoFuseSpec{
					Args: []string{"-okernel_cache"},
				},
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
					Path:       "/",
				}},
			}}, &Jindo{}, "-okernel_cache"},
	}
	for _, test := range tests {
		engine := &JindoFSxEngine{Log: fake.NullLogger()}
		properties := engine.transformFuseArg(test.runtime, test.dataset)
		if properties[0] != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Fuse.RunAs)
		}
	}
}

func TestParseFuseImage(t *testing.T) {
	var tests = []struct {
		runtime               *datav1alpha1.JindoRuntime
		dataset               *datav1alpha1.Dataset
		jindoValue            *Jindo
		expect                string
		expectImagePullPolicy string
	}{
		{
			runtime: &datav1alpha1.JindoRuntime{
				Spec: datav1alpha1.JindoRuntimeSpec{
					Secret: "secret",
				},
			},
			dataset: &datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: "local:///mnt/test",
						Name:       "test",
						Path:       "/",
					}},
				}},
			jindoValue:            &Jindo{},
			expect:                "registry.cn-shanghai.aliyuncs.com/jindofs/jindo-fuse:4.5.2",
			expectImagePullPolicy: "Always",
		},
		{
			runtime: &datav1alpha1.JindoRuntime{
				Spec: datav1alpha1.JindoRuntimeSpec{
					Secret: "secret",
					Fuse: datav1alpha1.JindoFuseSpec{
						Image:           "jindofs/jindo-fuse",
						ImageTag:        "testtag",
						ImagePullPolicy: "IfNotPresent",
					},
				},
			},
			dataset: &datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: "local:///mnt/test",
						Name:       "test",
						Path:       "/",
					}},
				}},
			jindoValue:            &Jindo{},
			expect:                "jindofs/jindo-fuse:testtag",
			expectImagePullPolicy: "IfNotPresent",
		},
	}
	for _, test := range tests {
		engine := &JindoFSxEngine{Log: fake.NullLogger()}
		imageR, tagR, imagePullPolicyR := engine.parseFuseImage(test.runtime)
		registryVersion := imageR + ":" + tagR
		if registryVersion != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, registryVersion)
		}
		if imagePullPolicyR != test.expectImagePullPolicy {
			t.Errorf("expected image pull policy %v, but got %v", test.expectImagePullPolicy, imagePullPolicyR)
		}
	}
}
