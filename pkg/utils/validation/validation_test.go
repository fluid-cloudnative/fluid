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

package validation_test

import (
	"fmt"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/fluid-cloudnative/fluid/pkg/utils/validation"
)

func TestValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Validation Suite")
}

var _ = Describe("IsValidMountRoot", func() {
	Context("with valid mount root paths", func() {
		DescribeTable("should accept valid paths",
			func(input string) {
				err := validation.IsValidMountRoot(input)
				Expect(err).To(BeNil())
			},
			Entry("validPath-1: path with double slashes", "/runtime-mnt//alluxio/default/hbase"),
			Entry("validPath-2: path with numbers, underscores and dot", "/opt/20-Runtime-Mnt_1/./alluxio/default/hbase"),
		)

		It("should have paths that start with /", func() {
			validPaths := []string{
				"/runtime-mnt//alluxio/default/hbase",
				"/opt/20-Runtime-Mnt_1/./alluxio/default/hbase",
			}

			for _, path := range validPaths {
				err := validation.IsValidMountRoot(path)
				if err == nil {
					Expect(path).To(HavePrefix(string(filepath.Separator)))
				}
			}
		})
	})

	Context("with invalid mount root paths", func() {
		DescribeTable("should reject invalid paths",
			func(input string, expectedErrorMsg string) {
				err := validation.IsValidMountRoot(input)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(expectedErrorMsg))
			},
			Entry("invalidPath-1: path with $",
				"/$test/alluxio/default/hbase",
				fmt.Sprintf(validation.InvalidMountRootErrMsgFmt, "/$test/alluxio/default/hbase", "every directory name in the mount root path should follow the relaxed DNS (RFC 1123) rule which additionally allows upper case alphabetic character and character '_'"),
			),
			Entry("invalidPath-2: path with parentheses",
				"/test/(alluxio)/default/hbase",
				fmt.Sprintf(validation.InvalidMountRootErrMsgFmt, "/test/(alluxio)/default/hbase", "every directory name in the mount root path should follow the relaxed DNS (RFC 1123) rule which additionally allows upper case alphabetic character and character '_'"),
			),
			Entry("invalidPath-3: path with semicolon",
				"/test/alluxio/def;ault/hbase",
				fmt.Sprintf(validation.InvalidMountRootErrMsgFmt, "/test/alluxio/def;ault/hbase", "every directory name in the mount root path should follow the relaxed DNS (RFC 1123) rule which additionally allows upper case alphabetic character and character '_'"),
			),
			Entry("invalidPath-4: empty path",
				"",
				fmt.Sprintf(validation.InvalidMountRootErrMsgFmt, "", "the mount root path is empty"),
			),
			Entry("invalidPath-5: relative path",
				"runtime-mnt/default",
				fmt.Sprintf(validation.InvalidMountRootErrMsgFmt, "runtime-mnt/default", "the mount root path must be an absolute path"),
			),
		)
	})

})
