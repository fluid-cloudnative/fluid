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
