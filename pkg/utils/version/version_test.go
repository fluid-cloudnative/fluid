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
	Describe("RuntimeVersion", func() {
		Context("when parsing valid version strings", func() {
			validVersions := []string{
				"2.7.2-SNAPSHOT-3714f2b",
				"release-2.7.2-SNAPSHOT-3714f2b",
				"2.8.0",
			}

			for _, versionStr := range validVersions {
				versionStr := versionStr // capture range variable
				It("should successfully parse "+versionStr, func() {
					ver, err := RuntimeVersion(versionStr)
					Expect(err).NotTo(HaveOccurred())
					Expect(ver).NotTo(BeNil())
					GinkgoWriter.Printf("Valid: %s, %v\n", versionStr, ver)
				})
			}
		})
	})

	Describe("Compare", func() {
		Context("when comparing version strings", func() {
			It("should return -1 when current version is less than other", func() {
				current := "release-2.7.2-SNAPSHOT-3714f2b"
				other := "2.8.0"

				result, err := Compare(current, other)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(-1))
			})

			It("should return error for invalid version format", func() {
				current := "test-2.7.2-SNAPSHOT-3714f2b"
				other := "2.8.0"

				result, err := Compare(current, other)

				Expect(err).To(HaveOccurred())
				Expect(result).To(Equal(0))
			})
		})

		Context("when comparing equal versions", func() {
			It("should return 0", func() {
				current := "2.8.0"
				other := "2.8.0"

				result, err := Compare(current, other)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(0))
			})
		})

		Context("when comparing greater versions", func() {
			It("should return 1 when current version is greater than other", func() {
				current := "2.9.0"
				other := "2.8.0"

				result, err := Compare(current, other)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(1))
			})
		})
	})
})
