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
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestTransformFuseWithNoArgs(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		dataset      *datav1alpha1.Dataset
		alluxioValue *Alluxio
		expect       string
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{datav1alpha1.Mount{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Alluxio{}, "--fuse-opts=kernel_cache,rw,max_read=131072,attr_timeout=7200,entry_timeout=7200,nonempty,allow_other"},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{Log: log.NullLogger{}}
		err := engine.transformFuse(test.runtime, test.dataset, test.alluxioValue)
		if err != nil {
			t.Errorf("Got err %v", err)
		}
		if test.alluxioValue.Fuse.Args[1] != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.alluxioValue.Fuse.Args[1])
		}
	}
}

func TestTransformFuseWithArgs(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		dataset      *datav1alpha1.Dataset
		alluxioValue *Alluxio
		expect       string
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
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{datav1alpha1.Mount{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Alluxio{}, "--fuse-opts=kernel_cache,allow_other"},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{Log: log.NullLogger{}}
		err := engine.transformFuse(test.runtime, test.dataset, test.alluxioValue)
		if err != nil {
			t.Errorf("Got err %v", err)
		}
		if test.alluxioValue.Fuse.Args[1] != test.expect {
			t.Errorf("expected fuse %v, but got %v", test.expect, test.alluxioValue.Fuse.Args[1])
		}
	}
}
