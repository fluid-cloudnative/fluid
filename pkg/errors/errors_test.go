/*
Copyright 2021 The Fluid Authors.

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
