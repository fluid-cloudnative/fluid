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

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			}}, &Alluxio{}, []string{"fuse", "--fuse-opts=kernel_cache,rw,max_read=131072,allow_other"}, true},
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Fuse: datav1alpha1.AlluxioFuseSpec{
					ImageTag: "v2.8.0",
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
			}}, &Alluxio{}, []string{"fuse", "--fuse-opts=kernel_cache,rw,max_read=131072,allow_other", "/alluxio/default/test/alluxio-fuse", "/"}, false},
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
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Alluxio{}, []string{"fuse", "--fuse-opts=kernel_cache,allow_other"}, true},
		{&datav1alpha1.AlluxioRuntime{
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
			}}, &Alluxio{}, []string{"fuse", "--fuse-opts=kernel_cache,allow_other"}, false},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{Log: fake.NullLogger()}
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
