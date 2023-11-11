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

	"github.com/fluid-cloudnative/fluid/pkg/common"
)

func TestGetStoragetLabelName(t *testing.T) {
	tests := []struct {
		info           RuntimeInfo
		expectedResult string
	}{
		{
			info: RuntimeInfo{
				name:        "spark",
				namespace:   "default",
				runtimeType: common.AlluxioRuntime,
			},

			expectedResult: "fluid.io/s-h-alluxio-m-default-spark",
		},
		{
			info: RuntimeInfo{
				name:                "hadoop",
				namespace:           "default",
				runtimeType:         common.AlluxioRuntime,
				deprecatedNodeLabel: true,
			},
			expectedResult: "data.fluid.io/storage-h-alluxio-m-default-hadoop",
		},
	}

	for _, test := range tests {
		result := test.info.getStoragetLabelName(common.HumanReadType, common.MemoryStorageType)
		if test.expectedResult != result {
			t.Errorf("check failure, expected %s, get %s", test.expectedResult, result)
		}
	}
}

func TestGetLabelNameForMemory(t *testing.T) {
	tests := []struct {
		info         RuntimeInfo
		expectResult string
	}{
		{
			info: RuntimeInfo{
				name:        "spark",
				namespace:   "default",
				runtimeType: common.AlluxioRuntime,
			},
			expectResult: "fluid.io/s-h-alluxio-m-default-spark",
		},
		{
			info: RuntimeInfo{
				name:                "hadoop",
				namespace:           "default",
				runtimeType:         common.AlluxioRuntime,
				deprecatedNodeLabel: true,
			},
			expectResult: "data.fluid.io/storage-human-alluxio-mem-default-hadoop",
		},
	}

	for _, test := range tests {
		result := test.info.GetLabelNameForMemory()
		if test.expectResult != result {
			t.Errorf("check failure, expected %s, get %s", test.expectResult, result)
		}
	}
}

func TestGetLabelNameForDisk(t *testing.T) {
	tests := []struct {
		info         RuntimeInfo
		expectResult string
	}{
		{
			info: RuntimeInfo{
				name:        "spark",
				namespace:   "default",
				runtimeType: common.AlluxioRuntime,
			},
			expectResult: "fluid.io/s-h-alluxio-d-default-spark",
		},
		{
			info: RuntimeInfo{
				name:                "hadoop",
				namespace:           "default",
				runtimeType:         common.AlluxioRuntime,
				deprecatedNodeLabel: true,
			},
			expectResult: "data.fluid.io/storage-human-alluxio-disk-default-hadoop",
		},
	}

	for _, test := range tests {
		result := test.info.GetLabelNameForDisk()
		if result != test.expectResult {
			t.Errorf("check failure, expected %s, get %s", test.expectResult, result)
		}
	}
}

func TestGetLabelNameForTotal(t *testing.T) {
	tests := []struct {
		info         RuntimeInfo
		expectResult string
	}{
		{
			info: RuntimeInfo{
				name:        "spark",
				namespace:   "default",
				runtimeType: common.AlluxioRuntime,
			},
			expectResult: "fluid.io/s-h-alluxio-t-default-spark",
		},
		{
			info: RuntimeInfo{
				name:                "hadoop",
				namespace:           "default",
				runtimeType:         common.AlluxioRuntime,
				deprecatedNodeLabel: true,
			},
			expectResult: "data.fluid.io/storage-human-alluxio-total-default-hadoop",
		},
	}

	for _, test := range tests {
		result := test.info.GetLabelNameForTotal()
		if result != test.expectResult {
			t.Errorf("check failure, expected %s, get %s", test.expectResult, result)
		}
	}
}

func TestGetCommonLabelName(t *testing.T) {
	tests := []struct {
		info         RuntimeInfo
		expectResult string
	}{
		{
			info: RuntimeInfo{
				name:      "spark",
				namespace: "default",
			},
			expectResult: "fluid.io/s-default-spark",
		},
		{
			info: RuntimeInfo{
				name:                "hadoop",
				namespace:           "default",
				deprecatedNodeLabel: true,
			},
			expectResult: "data.fluid.io/storage-default-hadoop",
		},
	}

	for _, test := range tests {
		result := test.info.GetCommonLabelName()
		if result != test.expectResult {
			t.Errorf("check failure, expected %s, get %s", test.expectResult, result)
		}
	}
}

func TestGetRuntimeLabelName(t *testing.T) {
	tests := []struct {
		info         RuntimeInfo
		expectResult string
	}{
		{
			info: RuntimeInfo{
				name:        "spark",
				namespace:   "default",
				runtimeType: common.AlluxioRuntime,
			},
			expectResult: "fluid.io/s-alluxio-default-spark",
		},
		{
			info: RuntimeInfo{
				name:                "hadoop",
				namespace:           "default",
				runtimeType:         common.AlluxioRuntime,
				deprecatedNodeLabel: true,
			},
			expectResult: "data.fluid.io/storage-alluxio-default-hadoop",
		},
	}

	for _, test := range tests {
		result := test.info.GetRuntimeLabelName()
		if result != test.expectResult {
			t.Errorf("check failure, expected %s, get %s", test.expectResult, result)
		}
	}
}

func TestGetDatasetNumLabelName(t *testing.T) {

	tests := []struct {
		info           RuntimeInfo
		expectedResult string
	}{
		{
			info:           RuntimeInfo{},
			expectedResult: common.LabelAnnotationDatasetNum,
		},
	}
	for _, test := range tests {
		result := test.info.GetDatasetNumLabelName()
		if result != test.expectedResult {
			t.Errorf("check failure, expected %s, get %s", test.expectedResult, result)
		}
	}

}
