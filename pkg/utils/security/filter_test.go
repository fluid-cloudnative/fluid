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

package security

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("FilterCommand", func() {
	DescribeTable("should filter sensitive keys from command",
		func(input, expect []string) {
			got := FilterCommand(input)
			Expect(got).To(Equal(expect))
		},
		Entry("withSensitiveKey",
			[]string{"mount", "fs", "aws.secretKey=xxxxxxxxx"},
			[]string{"mount", "fs", "aws.secretKey=[ redacted ]"},
		),
		Entry("withOutSensitiveKey",
			[]string{"mount", "fs", "alluxio.underfs.s3.inherit.acl=false"},
			[]string{"mount", "fs", "alluxio.underfs.s3.inherit.acl=false"},
		),
		Entry("key",
			[]string{"mount", "fs", "alluxio.underfs.s3.inherit.acl=false", "aws.secretKey=xxxxxxxxx"},
			[]string{"mount", "fs", "alluxio.underfs.s3.inherit.acl=false", "aws.secretKey=[ redacted ]"},
		),
	)
})

var _ = Describe("FilterCommandWithSensitive", func() {
	DescribeTable("should filter custom sensitive keys from command",
		func(filterKey string, input, expect []string) {
			UpdateSensitiveKey(filterKey)
			DeferCleanup(func() {
				// Reset by removing the added key
				delete(sensitiveKeys, filterKey)
			})
			got := FilterCommand(input)
			Expect(got).To(Equal(expect))
		},
		Entry("NotAddSensitiveKey",
			"test",
			[]string{"mount", "fs", "fs.azure.account.key=xxxxxxxxx"},
			[]string{"mount", "fs", "fs.azure.account.key=xxxxxxxxx"},
		),
		Entry("AddSensitiveKey",
			"fs.azure.account.key",
			[]string{"mount", "fs", "fs.azure.account.key=false"},
			[]string{"mount", "fs", "fs.azure.account.key=[ redacted ]"},
		),
	)
})
