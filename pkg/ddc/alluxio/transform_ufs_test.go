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
)

func TestTransformDatasetToVolume(t *testing.T) {
	var ufsPath = UFSPath{}
	ufsPath.Name = "test"
	ufsPath.HostPath = "/mnt/test"
	ufsPath.ContainerPath = "/opt/alluxio/underFSStorage/test"
	var ufsPath2 = UFSPath{}
	ufsPath2.Name = "test2"
	ufsPath2.ContainerPath = "/opt/alluxio/underFSStorage/test2"

	var tests = []struct {
		runtime *datav1alpha1.AlluxioRuntime
		dataset *datav1alpha1.Dataset
		value   *Alluxio
		expect  UFSPath
	}{
		{&datav1alpha1.AlluxioRuntime{}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{datav1alpha1.Mount{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			},
		}, &Alluxio{}, ufsPath},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{}
		engine.transformDatasetToVolume(test.runtime, test.dataset, test.value)
		if test.value.UFSPaths[0].HostPath != ufsPath.HostPath ||
			test.value.UFSPaths[0].ContainerPath != ufsPath.ContainerPath {
			t.Errorf("expected %v, got %v", test.expect, test.value)
		}
	}
}

func TestTransformDatasetToPVC(t *testing.T) {
	var ufsVolume = UFSVolume{}
	ufsVolume.Name = "test2"
	ufsVolume.ContainerPath = "/opt/alluxio/underFSStorage/test2"

	var tests = []struct {
		runtime *datav1alpha1.AlluxioRuntime
		dataset *datav1alpha1.Dataset
		value   *Alluxio
		expect  UFSVolume
	}{
		{&datav1alpha1.AlluxioRuntime{}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{datav1alpha1.Mount{
					MountPoint: "pvc://test2",
					Name:       "test2",
				}},
			},
		}, &Alluxio{}, ufsVolume},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{}
		engine.transformDatasetToVolume(test.runtime, test.dataset, test.value)
		if test.value.UFSVolumes[0].ContainerPath != ufsVolume.ContainerPath &&
			test.value.UFSVolumes[0].Name != ufsVolume.Name {
			t.Errorf("expected %v, got %v", test.expect, test.value)
		}
	}
}
