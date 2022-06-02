/*
Copyright 2022 The Fluid Authors.

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

package utils

import (
	"os"
	"testing"
)

func TestGetBoolValueFormEnv(t *testing.T) {
	// 1. env is not set
	testEnvNameNotFound := "envnotfound"
	expect := false

	got := GetBoolValueFormEnv(testEnvNameNotFound, false)

	if got != expect {
		t.Errorf("test failed due to expect %v, but got %v", expect, got)
	}

	// 2. env is set in true
	testEnvNameFound := "envFound"
	os.Setenv(testEnvNameFound, "true")
	expect = true

	got = GetBoolValueFormEnv(testEnvNameFound, false)
	if got != expect {
		t.Errorf("test failed due to expect %v, but got %v", expect, got)
	}

	// env is set T, which also means true
	os.Setenv(testEnvNameFound, "T")
	expect = true

	got = GetBoolValueFormEnv(testEnvNameFound, false)
	if got != expect {
		t.Errorf("test failed due to expect %v, but got %v", expect, got)
	}
}
