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
	"testing"
)

func TestIsSafePathWithSafePath(t *testing.T) {

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
		tt := filepath.Clean(test.input)
		print(tt)
		got := IsValidMountRoot(test.input)
		if got != nil {
			t.Errorf("testcase %s failed, expect no error happened, but got an error: %s", test.name, got.Error())
		}
	}
}

func TestIsSafePathWithInvalidPath(t *testing.T) {

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
		got := IsValidMountRoot(test.input)
		if got == nil {
			t.Errorf("testcase %s failed, expect an error happened, but got no error", test.name)
		}
		if got.Error() != test.expect.Error() {
			t.Errorf("testcase %s failed, expect error: %v, but got error: %v", test.name, test.expect, got)
		}
	}
}
