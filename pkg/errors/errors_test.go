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

package errors

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func resource(resource string) schema.GroupResource {
	return schema.GroupResource{Group: "", Resource: resource}
}

func TestIsDeprecated(t *testing.T) {

	testCases := []struct {
		Name   string
		Err    error
		expect bool
	}{
		{
			Name:   "deprecated",
			Err:    NewDeprecated(resource("test"), types.NamespacedName{}),
			expect: true,
		},
		{
			Name:   "no deprecated",
			Err:    fmt.Errorf("test"),
			expect: false,
		},
	}

	err := NewDeprecated(resource("test"), types.NamespacedName{})
	if err.Details() == nil {
		t.Errorf("expect error details %v is not nil", err.Details())
	}

	if len(err.Error()) == 0 {
		t.Errorf("expect error is not empty, but %v", err.Error())
	}

	for _, testCase := range testCases {
		if testCase.expect != IsDeprecated(testCase.Err) {
			t.Errorf("testCase %s: expected %v ,got %v", testCase.Name, testCase.expect, IsDeprecated(testCase.Err))
		}
	}
}

func TestIsNotSupported(t *testing.T) {
	testCases := []struct {
		Name   string
		Err    error
		expect bool
	}{
		{
			Name:   "notSupported",
			Err:    NewNotSupported(resource("DataBackup"), "ecaRuntime"),
			expect: true,
		},
		{
			Name:   "no notSupported",
			Err:    fmt.Errorf("test"),
			expect: false,
		},
	}

	err := NewNotSupported(resource("DataBackup"), "ecaRuntime")
	if err.Details() == nil {
		t.Errorf("expect error details %v is not nil", err.Details())
	}

	if len(err.Error()) == 0 {
		t.Errorf("expect error is not empty, but %v", err.Error())
	}

	for _, testCase := range testCases {
		if testCase.expect != IsNotSupported(testCase.Err) {
			t.Errorf("testCase %s: expected %v ,got %v", testCase.Name, testCase.expect, IsNotSupported(testCase.Err))
		}
	}
}
