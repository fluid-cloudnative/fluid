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

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func TestGetPersistentVolumeName(t *testing.T) {
	var testCases = []struct {
		runtimeName        string
		runtimeNamespace   string
		isDeprecatedPVName bool
		expectedPVName     string
	}{
		{
			runtimeName:        "spark",
			runtimeNamespace:   "fluid",
			isDeprecatedPVName: false,
			expectedPVName:     "fluid-spark",
		},
		{
			runtimeName:        "hadoop",
			runtimeNamespace:   "test",
			isDeprecatedPVName: false,
			expectedPVName:     "test-hadoop",
		},
		{
			runtimeName:        "hbase",
			runtimeNamespace:   "fluid",
			isDeprecatedPVName: true,
			expectedPVName:     "hbase",
		},
	}
	for _, testCase := range testCases {
		runtimeInfo, err := BuildRuntimeInfo(testCase.runtimeName, testCase.runtimeNamespace, "alluxio", datav1alpha1.TieredStore{})
		if err != nil {
			t.Errorf("fail to create the runtimeInfo with error %v", err)
		}
		runtimeInfo.SetDeprecatedPVName(testCase.isDeprecatedPVName)
		result := runtimeInfo.GetPersistentVolumeName()
		if result != testCase.expectedPVName {
			t.Errorf("get failure, expected %s, get %s", testCase.expectedPVName, result)
		}
	}
}
