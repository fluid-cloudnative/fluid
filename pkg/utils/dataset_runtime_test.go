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
package utils

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

func TestGetRuntimeByCategory(t *testing.T) {
	testCases := map[string]struct {
		runtimes  []datav1alpha1.Runtime
		wantIndex int
	}{
		"test get runtime by category case 1": {
			runtimes:  mockThreeRuntimes(2, common.AccelerateCategory),
			wantIndex: 2,
		},
		"test get runtime by category case 2": {
			runtimes:  mockThreeRuntimes(0, common.AccelerateCategory),
			wantIndex: 0,
		},
		"test get runtime by category case 3": {
			runtimes:  mockThreeRuntimes(4, common.AccelerateCategory),
			wantIndex: -1,
		},
		"test get runtime by category case 4": {
			runtimes:  mockThreeRuntimes(1, common.AccelerateCategory),
			wantIndex: 1,
		},
		"test get runtime by category case 5": {
			runtimes:  nil,
			wantIndex: -1,
		},
	}

	for k, item := range testCases {
		gotIndex, _ := GetRuntimeByCategory(item.runtimes, common.AccelerateCategory)
		if gotIndex != item.wantIndex {
			t.Errorf("%s check failure, want index:%v,got index:%v", k, item.wantIndex, gotIndex)
		}

	}
}

func mockThreeRuntimes(index int, category common.Category) []datav1alpha1.Runtime {
	list := make([]datav1alpha1.Runtime, 0)

	r1 := datav1alpha1.Runtime{}
	list = append(list, r1)

	r2 := datav1alpha1.Runtime{}
	list = append(list, r2)

	r3 := datav1alpha1.Runtime{}
	list = append(list, r3)

	if index < len(list) && index >= 0 {
		list[index].Category = category
	}

	return list
}
