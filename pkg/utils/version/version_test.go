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
	"testing"
)

func TestVersion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Version Suite")
}

var _ = Describe("Version", func() {
	DescribeTable("RuntimeVersion",
		func(versionStr string) {
			ver, err := RuntimeVersion(versionStr)
			Expect(err).NotTo(HaveOccurred())
			Expect(ver).NotTo(BeNil())
			GinkgoWriter.Printf("Valid: %s, %v\n", versionStr, ver)
		},
		Entry("should parse snapshot version", "2.7.2-SNAPSHOT-3714f2b"),
		Entry("should parse release snapshot version", "release-2.7.2-SNAPSHOT-3714f2b"),
		Entry("should parse simple version", "2.8.0"),
	)

	DescribeTable("Compare",
		func(current, other string, expectedResult int, expectError bool) {
			result, err := Compare(current, other)
			if expectError {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(result).To(Equal(expectedResult))
		},
		Entry("should return -1 when current version is less than other", "release-2.7.2-SNAPSHOT-3714f2b", "2.8.0", -1, false),
		Entry("should return 0 when versions are equal", "2.8.0", "2.8.0", 0, false),
		Entry("should return 1 when current version is greater than other", "2.9.0", "2.8.0", 1, false),
		Entry("should return error for invalid version format", "test-2.7.2-SNAPSHOT-3714f2b", "2.8.0", 0, true),
	)
})
