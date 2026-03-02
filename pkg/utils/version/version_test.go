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

package version

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version Utils", func() {
	Describe("RuntimeVersion", func() {
		validVersions := []string{
			"2.7.2-SNAPSHOT-3714f2b",
			"release-2.7.2-SNAPSHOT-3714f2b",
			"2.8.0",
		}
		for _, s := range validVersions {
			versionStr := s
			It("should parse valid version: "+versionStr, func() {
				ver, err := RuntimeVersion(versionStr)
				GinkgoWriter.Println("Valid: ", versionStr, ver, err)
				Expect(err).NotTo(HaveOccurred())
			})
		}
	})

	Describe("Compare", func() {
		DescribeTable("should compare versions correctly",
			func(current, other string, wantError bool, want int) {
				got, err := Compare(current, other)
				if wantError {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
				Expect(got).To(Equal(want))
			},
			Entry("lessThan", "release-2.7.2-SNAPSHOT-3714f2b", "2.8.0", false, -1),
			Entry("error", "test-2.7.2-SNAPSHOT-3714f2b", "2.8.0", true, 0),
		)
	})
})
