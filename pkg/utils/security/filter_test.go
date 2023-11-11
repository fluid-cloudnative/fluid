/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
