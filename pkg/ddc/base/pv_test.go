/*
Copyright 2021 The Fluid Authors.

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

package base

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
)

// TestGetPersistentVolumeName tests the GetPersistentVolumeName method of RuntimeInfo.
// This function primarily verifies that the persistent volume name is correctly generated
// based on the runtime namespace and runtime name with a hyphen separator.
//
// Parameters:
// - t (*testing.T): The testing context used for reporting test failures.
//
// Returns:
// - None (void function): Test results are reported directly via the testing.T instance.
//
// The test validates that:
//  1. RuntimeInfo can be built successfully for different runtime configurations.
//  2. The generated PV name follows the format: {namespace}-{runtimeName}.
//  3. The method handles various namespace and runtime name combinations correctly.
func TestGetPersistentVolumeName(t *testing.T) {
	var testCases = []struct {
		runtimeName      string
		runtimeNamespace string
		expectedPVName   string
	}{
		{
			runtimeName:      "spark",
			runtimeNamespace: "fluid",
			expectedPVName:   "fluid-spark",
		},
		{
			runtimeName:      "hadoop",
			runtimeNamespace: "test",
			expectedPVName:   "test-hadoop",
		},
	}
	for _, testCase := range testCases {
		runtimeInfo, err := BuildRuntimeInfo(testCase.runtimeName, testCase.runtimeNamespace, common.AlluxioRuntime)
		if err != nil {
			t.Errorf("fail to create the runtimeInfo with error %v", err)
		}
		result := runtimeInfo.GetPersistentVolumeName()
		if result != testCase.expectedPVName {
			t.Errorf("get failure, expected %s, get %s", testCase.expectedPVName, result)
		}
	}
}
