/*
Copyright 2023 The Fluid Authors.

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

package lifecycle

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestToleratesTaints(t *testing.T) {
	taintsNull := []v1.Taint{}

	taintAll := []v1.Taint{
		{
			Effect: v1.TaintEffectNoExecute,
			Value:  "taint_test_value",
		},
	}

	tolerationAll := []v1.Toleration{
		{
			Operator: "",
			Value:    "taint_test_value",
		},
	}

	tolerationNone := []v1.Toleration{
		{
			Operator: "",
			Value:    "taint_test_none",
		},
	}

	var tests = []struct {
		taints         []v1.Taint
		tolerations    []v1.Toleration
		expectedResult bool
	}{
		{
			taints:         taintsNull,
			tolerations:    nil,
			expectedResult: false,
		},
		{
			taints:         taintAll,
			tolerations:    tolerationAll,
			expectedResult: true,
		},
		{
			taints:         taintAll,
			tolerations:    tolerationNone,
			expectedResult: false,
		},
	}

	for _, test := range tests {
		if result := toleratesTaints(test.taints, test.tolerations); result != test.expectedResult {
			t.Errorf("expected %t, get %t", test.expectedResult, result)
		}
	}

}
