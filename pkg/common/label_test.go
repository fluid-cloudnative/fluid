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

package common

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("LabelToModify", func() {
	ginkgo.Describe("GetLabelKey, GetLabelValue, GetOperationType", func() {
		ginkgo.It("should return correct values", func() {
			labelToModify := LabelToModify{
				labelKey:      "commonLabel",
				labelValue:    "true",
				operationType: AddLabel,
			}

			gomega.Expect(labelToModify.GetLabelKey()).To(gomega.Equal("commonLabel"))
			gomega.Expect(labelToModify.GetLabelValue()).To(gomega.Equal("true"))
			gomega.Expect(labelToModify.GetOperationType()).To(gomega.Equal(AddLabel))
		})
	})
})

var _ = ginkgo.Describe("LabelsToModify", func() {
	ginkgo.Describe("GetLabels", func() {
		ginkgo.It("should return labels after adding", func() {
			var labelsToModify LabelsToModify
			labelsToModify.Add("commonLabel", "true")

			expected := []LabelToModify{
				{
					labelKey:      "commonLabel",
					labelValue:    "true",
					operationType: AddLabel,
				},
			}

			gomega.Expect(labelsToModify.GetLabels()).To(gomega.Equal(expected))
		})
	})

	ginkgo.Describe("operator", func() {
		ginkgo.It("should add label with AddLabel operation", func() {
			var labelsToModify LabelsToModify
			labelsToModify.operator("commonLabel", "true", AddLabel)

			expected := LabelToModify{
				labelKey:      "commonLabel",
				labelValue:    "true",
				operationType: AddLabel,
			}

			gomega.Expect(labelsToModify.labels[0]).To(gomega.Equal(expected))
		})

		ginkgo.It("should add label with DeleteLabel operation without value", func() {
			var labelsToModify LabelsToModify
			labelsToModify.operator("commonLabel", "true", DeleteLabel)

			expected := LabelToModify{
				labelKey:      "commonLabel",
				operationType: DeleteLabel,
			}

			gomega.Expect(labelsToModify.labels[0]).To(gomega.Equal(expected))
		})
	})

	ginkgo.Describe("Add", func() {
		ginkgo.It("should add label to modify slice", func() {
			var labelsToModify LabelsToModify
			labelsToModify.Add("commonLabel", "true")

			expected := []LabelToModify{
				{
					labelKey:      "commonLabel",
					labelValue:    "true",
					operationType: AddLabel,
				},
			}

			gomega.Expect(labelsToModify.GetLabels()).To(gomega.Equal(expected))
		})
	})

	ginkgo.Describe("Delete", func() {
		ginkgo.It("should add delete operation to modify slice", func() {
			var labelsToModify LabelsToModify
			labelsToModify.Delete("commonLabel")

			expected := []LabelToModify{
				{
					labelKey:      "commonLabel",
					operationType: DeleteLabel,
				},
			}

			gomega.Expect(labelsToModify.GetLabels()).To(gomega.Equal(expected))
		})
	})

	ginkgo.Describe("Update", func() {
		ginkgo.It("should add update operation to modify slice", func() {
			var labelsToModify LabelsToModify
			labelsToModify.Update("commonLabel", "")

			expected := []LabelToModify{
				{
					labelKey:      "commonLabel",
					operationType: UpdateLabel,
				},
			}

			gomega.Expect(labelsToModify.GetLabels()).To(gomega.Equal(expected))
		})
	})
})

var _ = ginkgo.Describe("CheckExpectValue", func() {
	ginkgo.DescribeTable("should check if label has expected value",
		func(labels map[string]string, target, targetValue string, wantHit bool) {
			gomega.Expect(CheckExpectValue(labels, target, targetValue)).To(gomega.Equal(wantHit))
		},
		ginkgo.Entry("label matches target and value",
			map[string]string{EnableFluidInjectionFlag: "true"},
			EnableFluidInjectionFlag, "true", true),
		ginkgo.Entry("label matches target but not value",
			map[string]string{EnableFluidInjectionFlag: "false"},
			EnableFluidInjectionFlag, "true", false),
		ginkgo.Entry("nil labels",
			nil,
			EnableFluidInjectionFlag, "true", false),
	)
})

var _ = ginkgo.Describe("LabelAnnotationPodSchedRegex", func() {
	ginkgo.DescribeTable("should match correct patterns",
		func(target string, shouldMatch bool, expectedGroup string) {
			submatch := LabelAnnotationPodSchedRegex.FindStringSubmatch(target)
			if shouldMatch {
				gomega.Expect(submatch).To(gomega.HaveLen(2))
				gomega.Expect(submatch[1]).To(gomega.Equal(expectedGroup))
			} else {
				gomega.Expect(len(submatch)).NotTo(gomega.Equal(2))
			}
		},
		ginkgo.Entry("correct pattern",
			LabelAnnotationDataset+".dsA.sched", true, "dsA"),
		ginkgo.Entry("wrong fluid.io prefix",
			"fluidaio/dataset.dsA.sched", false, ""),
		ginkgo.Entry("wrong prefix",
			"a.fluid.io/dataset.dsA.sched", false, ""),
	)
})
