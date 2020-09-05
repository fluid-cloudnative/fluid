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
	var defaultStr = "default string"
	var nonnullStr = "non-null string"
	var tests = []struct {
		runtime *datav1alpha1.AlluxioRuntime
		dataset *datav1alpha1.Dataset
		value   *Alluxio
		expect  UFSPath
	}{
		{&datav1alpha1.AlluxioRuntime{}, &datav1alpha1.Dataset{
			Spec: {
				Mounts: []datav1alpha1.Mount{
					datav1alpha1.Mount{
						MountPoint: "path:/mnt/test",
						Name:       "test",
					},
				},
			},
		}, &Alluxio{}, UFSPath{
			Name:          "test",
			ContainerPath: "/opt/alluxio/underFSStorage/test",
			HostPath:      "/mnt/test",
		}},
	}
	for _, test := range tests {

		transformDatasetToVolume(test.runtime, test.dataset, test.value)
		if len(test.value.UFSPaths) != 1 {
			t.Errorf("expected %v, got %v", test.expect, test.value)
		}
	}
}
