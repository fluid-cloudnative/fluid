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
