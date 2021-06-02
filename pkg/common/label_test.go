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
package common

import "testing"

func TestHitTarget(t *testing.T) {
	testCases := map[string]struct {
		labels  map[string]string
		target  string
		wantHit bool
	}{
		"test label target hit case 1": {
			labels:  map[string]string{LabelFluidSchedulingStrategyFlag: "true"},
			target:  LabelFluidSchedulingStrategyFlag,
			wantHit: true,
		},
		"test label target hit case 2": {
			labels:  map[string]string{LabelFluidSchedulingStrategyFlag: "false"},
			target:  LabelFluidSchedulingStrategyFlag,
			wantHit: false,
		},
	}

	for index, item := range testCases {
		gotHit := HitTarget(item.labels, item.target)
		if gotHit != item.wantHit {
			t.Errorf("%s check failure, want:%t,got:%t", index, item.wantHit, gotHit)
		}
	}

}
