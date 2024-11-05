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
	"reflect"
	"testing"
)

func TestFilterCommand(t *testing.T) {

	type testCase struct {
		name   string
		input  []string
		expect []string
	}

	testCases := []testCase{
		{
			name:   "withSensitiveKey",
			input:  []string{"mount", "fs", "aws.secretKey=xxxxxxxxx"},
			expect: []string{"mount", "fs", "aws.secretKey=[ redacted ]"},
		}, {
			name:   "withOutSensitiveKey",
			input:  []string{"mount", "fs", "alluxio.underfs.s3.inherit.acl=false"},
			expect: []string{"mount", "fs", "alluxio.underfs.s3.inherit.acl=false"},
		}, {
			name:   "key",
			input:  []string{"mount", "fs", "alluxio.underfs.s3.inherit.acl=false", "aws.secretKey=xxxxxxxxx"},
			expect: []string{"mount", "fs", "alluxio.underfs.s3.inherit.acl=false", "aws.secretKey=[ redacted ]"},
		},
	}

	for _, test := range testCases {
		got := FilterCommand(test.input)
		if !reflect.DeepEqual(got, test.expect) {
			t.Errorf("testcase %s is failed due to expect %v, but got %v", test.name, test.expect, got)
		}
	}

}

func TestFilterCommandWithSensitive(t *testing.T) {

	type testCase struct {
		name      string
		filterKey string
		input     []string
		expect    []string
	}

	testCases := []testCase{
		{
			name:      "NotAddSensitiveKey",
			filterKey: "test",
			input:     []string{"mount", "fs", "fs.azure.account.key=xxxxxxxxx"},
			expect:    []string{"mount", "fs", "fs.azure.account.key=xxxxxxxxx"},
		}, {
			name:      "AddSensitiveKey",
			filterKey: "fs.azure.account.key",
			input:     []string{"mount", "fs", "fs.azure.account.key=false"},
			expect:    []string{"mount", "fs", "fs.azure.account.key=[ redacted ]"},
		},
	}

	for _, test := range testCases {
		UpdateSensitiveKey(test.filterKey)
		got := FilterCommand(test.input)
		if !reflect.DeepEqual(got, test.expect) {
			t.Errorf("testcase %s is failed due to expect %v, but got %v", test.name, test.expect, got)
		}
	}

}
