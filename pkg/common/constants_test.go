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

package common

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetDefaultTieredStoreOrder", func() {
	type testCase struct {
		mediumType MediumType
		want       int
	}

	DescribeTable("should return correct order for each medium type",
		func(tc testCase) {
			Expect(GetDefaultTieredStoreOrder(tc.mediumType)).To(Equal(tc.want))
		},
		Entry("returns 0 for Memory", testCase{mediumType: Memory, want: 0}),
		Entry("returns 1 for SSD", testCase{mediumType: SSD, want: 1}),
		Entry("returns 2 for HDD", testCase{mediumType: HDD, want: 2}),
		Entry("returns 0 for unknown", testCase{mediumType: "unknown", want: 0}),
		Entry("returns 0 for empty string", testCase{mediumType: "", want: 0}),
		Entry("returns 0 for nil MediumType", testCase{mediumType: MediumType(""), want: 0}),
		Entry("returns 0 for random string", testCase{mediumType: MediumType("ramdisk"), want: 0}),
	)
})
