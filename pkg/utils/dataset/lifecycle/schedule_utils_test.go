/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
