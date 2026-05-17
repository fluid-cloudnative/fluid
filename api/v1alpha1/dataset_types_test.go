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

package v1alpha1

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const dataLoadOperation = "DataLoad"

var _ = Describe("Dataset operation references", func() {
	Describe("RemoveDataOperationInProgress", func() {
		DescribeTable("removes the target operation reference",
			func(dataset *Dataset, operationType string, name string, expected string) {
				Expect(dataset.RemoveDataOperationInProgress(operationType, name)).To(Equal(expected))
				Expect(dataset.GetDataOperationInProgress(operationType)).To(Equal(expected))
			},
			Entry("removes the only in-progress operation",
				&Dataset{Status: DatasetStatus{OperationRef: map[string]string{dataLoadOperation: "test1"}}},
				dataLoadOperation,
				"test1",
				"",
			),
			Entry("removes one operation from a list",
				&Dataset{Status: DatasetStatus{OperationRef: map[string]string{dataLoadOperation: "test1,test2"}}},
				dataLoadOperation,
				"test1",
				"test2",
			),
			Entry("returns empty when no operation refs are recorded",
				&Dataset{Status: DatasetStatus{}},
				dataLoadOperation,
				"test1",
				"",
			),
		)
	})

	Describe("SetDataOperationInProgress", func() {
		DescribeTable("tracks the operation reference for the requested operation type",
			func(dataset *Dataset, operationType string, name string, expected string) {
				dataset.SetDataOperationInProgress(operationType, name)

				Expect(dataset.GetDataOperationInProgress(operationType)).To(Equal(expected))
			},
			Entry("creates the first operation ref",
				&Dataset{Status: DatasetStatus{}},
				dataLoadOperation,
				"test1",
				"test1",
			),
			Entry("appends a new operation ref for the same type",
				&Dataset{Status: DatasetStatus{OperationRef: map[string]string{dataLoadOperation: "test1"}}},
				dataLoadOperation,
				"test2",
				"test1,test2",
			),
			Entry("records a different operation type independently",
				&Dataset{Status: DatasetStatus{OperationRef: map[string]string{dataLoadOperation: "test1"}}},
				"DataMigrate",
				"test",
				"test",
			),
			Entry("keeps an existing operation ref without duplication",
				&Dataset{Status: DatasetStatus{OperationRef: map[string]string{dataLoadOperation: "test"}}},
				dataLoadOperation,
				"test",
				"test",
			),
		)
	})

	Describe("CanbeBound", func() {
		It("returns true when no runtime is recorded", func() {
			dataset := &Dataset{}

			Expect(dataset.CanbeBound("runtime", "fluid", common.AccelerateCategory)).To(BeTrue())
		})

		DescribeTable("matches the runtime identity",
			func(name, namespace string, category common.Category, expected bool) {
				dataset := &Dataset{
					Status: DatasetStatus{
						Runtimes: []Runtime{
							{Name: "target", Namespace: "fluid", Category: common.AccelerateCategory},
						},
					},
				}

				Expect(dataset.CanbeBound(name, namespace, category)).To(Equal(expected))
			},
			Entry("matching identity", "target", "fluid", common.AccelerateCategory, true),
			Entry("name mismatch", "other", "fluid", common.AccelerateCategory, false),
			Entry("namespace mismatch", "target", "other", common.AccelerateCategory, false),
			Entry("category mismatch", "target", "fluid", common.Category("other"), false),
		)
	})

	DescribeTable("IsExclusiveMode",
		func(mode PlacementMode, expected bool) {
			dataset := &Dataset{Spec: DatasetSpec{PlacementMode: mode}}

			Expect(dataset.IsExclusiveMode()).To(Equal(expected))
		},
		Entry("default placement mode", DefaultMode, true),
		Entry("exclusive placement mode", ExclusiveMode, true),
		Entry("shared placement mode", ShareMode, false),
	)
})
