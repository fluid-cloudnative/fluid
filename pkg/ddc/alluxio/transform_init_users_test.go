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

func TestTransformInitUsersWithoutRunAs(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{},
		}, &Alluxio{}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{Log: log.NullLogger{}}
		engine.transformInitUsers(test.runtime, test.alluxioValue)
		if test.alluxioValue.InitUsers.Enabled {
			t.Errorf("expected init users are disabled, but got %v", test.alluxioValue.InitUsers.Enabled)
		}
	}
}

func TestTransformInitUsersWithRunAs(t *testing.T) {

	value := int64(1000)
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				RunAs: &datav1alpha1.User{
					UID:       &value,
					GID:       &value,
					UserName:  "user1",
					GroupName: "group1",
				},
			},
		}, &Alluxio{}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{Log: log.NullLogger{}}
		engine.transformInitUsers(test.runtime, test.alluxioValue)
		if !test.alluxioValue.InitUsers.Enabled {
			t.Errorf("expected init users are enabled, but got %v", test.alluxioValue.InitUsers.Enabled)
		}
	}
}
