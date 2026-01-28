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

package validation

import (
	"fmt"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IsValidMountRoot", func() {
	Describe("Safe paths", func() {
		type testCase struct {
			name  string
			input string
		}
		testCases := []testCase{
			{
				name:  "validPath-1",
				input: "/runtime-mnt//alluxio/default/hbase",
			},
			{
				name:  "validPath-2",
				input: "/opt/20-Runtime-Mnt_1/./alluxio/default/hbase",
			},
		}
		for _, test := range testCases {
			tc := test
			It(fmt.Sprintf("should accept %s", tc.name), func() {
				tt := filepath.Clean(tc.input)
				GinkgoWriter.Println(tt)
				got := IsValidMountRoot(tc.input)
				Expect(got).To(BeNil())
			})
		}
	})

	Describe("Invalid paths", func() {
		type testCase struct {
			name   string
			input  string
			expect error
		}
		testCases := []testCase{
			{
				name:   "invalidPath-1",
				input:  "/$test/alluxio/default/hbase",
				expect: fmt.Errorf(invalidMountRootErrMsgFmt, "/$test/alluxio/default/hbase", "every directory name in the mount root path shuold follow the relaxed DNS (RFC 1123) rule which additionally allows upper case alphabetic character and character '_'"),
			},
			{
				name:   "invalidPath-2",
				input:  "/test/(alluxio)/default/hbase",
				expect: fmt.Errorf(invalidMountRootErrMsgFmt, "/test/(alluxio)/default/hbase", "every directory name in the mount root path shuold follow the relaxed DNS (RFC 1123) rule which additionally allows upper case alphabetic character and character '_'"),
			},
			{
				name:   "invalidPath-3",
				input:  "/test/alluxio/def;ault/hbase",
				expect: fmt.Errorf(invalidMountRootErrMsgFmt, "/test/alluxio/def;ault/hbase", "every directory name in the mount root path shuold follow the relaxed DNS (RFC 1123) rule which additionally allows upper case alphabetic character and character '_'"),
			},
			{
				name:   "invalidPath-4",
				input:  "",
				expect: fmt.Errorf(invalidMountRootErrMsgFmt, "", "the mount root path is empty"),
			},
			{
				name:   "invalidPath-5",
				input:  "runtime-mnt/default",
				expect: fmt.Errorf(invalidMountRootErrMsgFmt, "runtime-mnt/default", "the mount root path must be an absolute path"),
			},
		}
		for _, test := range testCases {
			tc := test
			It(fmt.Sprintf("should reject %s", tc.name), func() {
				got := IsValidMountRoot(tc.input)
				Expect(got).To(HaveOccurred())
				Expect(got.Error()).To(Equal(tc.expect.Error()))
			})
		}
	})

	Describe("Fuzz: path must start with / if valid", func() {
		It("should only accept valid paths starting with /", func() {
			// Example fuzz-like check for a few cases
			inputs := []string{"", "foo", "/foo", "foo/bar", "/foo/bar"}
			for _, input := range inputs {
				err := IsValidMountRoot(input)
				if err == nil {
					Expect(strings.HasPrefix(input, string(filepath.Separator))).To(BeTrue(), "valid input must start with '/'")
				}
			}
		})
	})
})
