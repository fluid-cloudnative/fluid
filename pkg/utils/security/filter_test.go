/*
Copyright 2023 The Fluid Author.

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

package security

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestSecurity(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Security Suite")
}

var _ = Describe("FilterCommand", func() {
	Context("when filtering commands", func() {
		It("should redact sensitive keys", func() {
			input := []string{"mount", "fs", "aws.secretKey=xxxxxxxxx"}
			expect := []string{"mount", "fs", "aws.secretKey=[ redacted ]"}
			got := FilterCommand(input)
			Expect(got).To(Equal(expect))
		})

		It("should not modify commands without sensitive keys", func() {
			input := []string{"mount", "fs", "alluxio.underfs.s3.inherit.acl=false"}
			expect := []string{"mount", "fs", "alluxio.underfs.s3.inherit.acl=false"}
			got := FilterCommand(input)
			Expect(got).To(Equal(expect))
		})

		It("should redact sensitive keys while preserving other parameters", func() {
			input := []string{"mount", "fs", "alluxio.underfs.s3.inherit.acl=false", "aws.secretKey=xxxxxxxxx"}
			expect := []string{"mount", "fs", "alluxio.underfs.s3.inherit.acl=false", "aws.secretKey=[ redacted ]"}
			got := FilterCommand(input)
			Expect(got).To(Equal(expect))
		})
	})
})

var _ = Describe("FilterCommandWithSensitive", func() {
	Context("when updating sensitive keys", func() {
		It("should not redact keys that are not added as sensitive", func() {
			filterKey := "test"
			input := []string{"mount", "fs", "fs.azure.account.key=xxxxxxxxx"}
			expect := []string{"mount", "fs", "fs.azure.account.key=xxxxxxxxx"}
			UpdateSensitiveKey(filterKey)
			got := FilterCommand(input)
			Expect(got).To(Equal(expect))
		})

		It("should redact keys that are added as sensitive", func() {
			filterKey := "fs.azure.account.key"
			input := []string{"mount", "fs", "fs.azure.account.key=false"}
			expect := []string{"mount", "fs", "fs.azure.account.key=[ redacted ]"}
			UpdateSensitiveKey(filterKey)
			got := FilterCommand(input)
			Expect(got).To(Equal(expect))
		})
	})
})
