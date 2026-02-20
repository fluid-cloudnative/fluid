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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Label Functions", func() {
	Describe("GetStorageLabelName", func() {
		It("should return correct storage label name", func() {
			info := RuntimeInfo{
				name:        "spark",
				namespace:   "default",
				runtimeType: common.AlluxioRuntime,
			}
			result := utils.GetStorageLabelName(common.HumanReadType, common.MemoryStorageType, info.runtimeType, info.namespace, info.name, "")
			Expect(result).To(Equal("fluid.io/s-h-alluxio-m-default-spark"))
		})
	})

	Describe("GetLabelNameForMemory", func() {
		It("should return correct memory label name", func() {
			info := RuntimeInfo{
				name:        "spark",
				namespace:   "default",
				runtimeType: common.AlluxioRuntime,
			}
			result := info.GetLabelNameForMemory()
			Expect(result).To(Equal("fluid.io/s-h-alluxio-m-default-spark"))
		})
	})

	Describe("GetLabelNameForDisk", func() {
		It("should return correct disk label name", func() {
			info := RuntimeInfo{
				name:        "spark",
				namespace:   "default",
				runtimeType: common.AlluxioRuntime,
			}
			result := info.GetLabelNameForDisk()
			Expect(result).To(Equal("fluid.io/s-h-alluxio-d-default-spark"))
		})
	})

	Describe("GetLabelNameForTotal", func() {
		It("should return correct total label name", func() {
			info := RuntimeInfo{
				name:        "spark",
				namespace:   "default",
				runtimeType: common.AlluxioRuntime,
			}
			result := info.GetLabelNameForTotal()
			Expect(result).To(Equal("fluid.io/s-h-alluxio-t-default-spark"))
		})
	})

	Describe("GetCommonLabelName", func() {
		It("should return correct common label name", func() {
			info := RuntimeInfo{
				name:      "spark",
				namespace: "default",
			}
			result := info.GetCommonLabelName()
			Expect(result).To(Equal("fluid.io/s-default-spark"))
		})
	})

	Describe("GetRuntimeLabelName", func() {
		It("should return correct runtime label name", func() {
			info := RuntimeInfo{
				name:        "spark",
				namespace:   "default",
				runtimeType: common.AlluxioRuntime,
			}
			result := info.GetRuntimeLabelName()
			Expect(result).To(Equal("fluid.io/s-alluxio-default-spark"))
		})
	})

	Describe("GetDatasetNumLabelName", func() {
		It("should return dataset num label name", func() {
			info := RuntimeInfo{}
			result := info.GetDatasetNumLabelName()
			Expect(result).To(Equal(common.LabelAnnotationDatasetNum))
		})
	})
})
