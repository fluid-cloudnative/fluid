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
)

func TestAlluxioEngine_calculateMountPointsChanges(t *testing.T) {

	testCases := map[string]struct {
		mounted []string
		current []string
		expect  map[string][]string
	}{
		"calculate mount point changes test case 1": {
			mounted: []string{"hadoop3.3.0"},
			current: []string{"hadoopcurrent", "hadoop3.3.0"},
			expect:  map[string][]string{"added": {"hadoopcurrent"}},
		},
		"calculate mount point changes test case 2": {
			mounted: []string{"hadoopcurrent", "hadoop3.3.0"},
			current: []string{"hadoop3.3.0"},
			expect:  map[string][]string{"removed": {"hadoopcurrent"}},
		},
		"calculate mount point changes test case 3": {
			mounted: []string{"hadoopcurrent", "hadoop3.2.2"},
			current: []string{"hadoop3.3.0", "hadoop3.2.2"},
			expect:  map[string][]string{"added": {"hadoop3.3.0"}, "removed": {"hadoopcurrent"}},
		},
		"calculate mount point changes test case 4": {
			mounted: []string{"hadoop3.3.0"},
			current: []string{"hadoop3.3.0"},
			expect:  map[string][]string{},
		},
		"calculate mount point changes test case 5": {
			mounted: []string{"hadoopcurrent", "hadoop3.2.2"},
			current: []string{"hadoop3.3.0", "hadoop3.2.2", "hadoop3.3.1"},
			expect:  map[string][]string{"added": {"hadoop3.3.0", "hadoop3.3.1"}, "removed": {"hadoopcurrent"}},
		},
	}

	for _, item := range testCases {
		engine := &AlluxioEngine{}
		added, removed := engine.calculateMountPointsChanges(item.mounted, item.current)

		if !ArrayEqual(added, item.expect["added"]) {
			t.Errorf("expected added %v, got %v", item.expect["added"], added)
		}
		if !ArrayEqual(removed, item.expect["removed"]) {
			t.Errorf("expected removed %v, got %v", item.expect["removed"], removed)
		}
	}

}

func ArrayEqual(a, b []string) bool {
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for key, val := range a {
		if val != b[key] {
			return false
		}
	}
	return true
}
