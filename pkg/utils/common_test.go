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
	"os"
	"testing"
)

func TestGetEnvByKey(t *testing.T) {
	testCases := map[string]struct {
		key       string
		envKey    string
		value     string
		wantValue string
	}{
		"test get env value by key case 1": {
			key:       "MY_POD_NAMESPACE",
			envKey:    "MY_POD_NAMESPACE",
			value:     "fluid-system",
			wantValue: "fluid-system",
		},
		"test get env value by key case 2": {
			key:       "MY_POD_NAMESPACE",
			envKey:    "MY_POD_NAMESPACES",
			value:     "fluid-system",
			wantValue: "",
		},
	}

	for k, item := range testCases {
		// prepare env value
		os.Setenv(item.key, item.value)
		gotValue, _ := GetEnvByKey(item.envKey)
		if gotValue != item.wantValue {
			t.Errorf("%s check failure, want:%v,got:%v", k, item.wantValue, gotValue)
		}
	}
}
