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
	"testing"
	"time"
)

func TestGetBoolValueFormEnv(t *testing.T) {
	t.Run("env is not set", func(t *testing.T) {
		testEnvNameNotFound := "envnotfound"
		expect := false

		got := GetBoolValueFromEnv(testEnvNameNotFound, false)

		if got != expect {
			t.Errorf("test failed due to expect %v, but got %v", expect, got)
		}
	})

	t.Run("env is set in true", func(t *testing.T) {
		testEnvNameFound := "envFound"
		t.Setenv(testEnvNameFound, "true")
		expect := true

		got := GetBoolValueFromEnv(testEnvNameFound, false)
		if got != expect {
			t.Errorf("test failed due to expect %v, but got %v", expect, got)
		}
	})

	t.Run("env is set T, which also means true", func(t *testing.T) {
		testEnvNameFound := "envFound"
		t.Setenv(testEnvNameFound, "T")
		expect := true

		got := GetBoolValueFromEnv(testEnvNameFound, false)
		if got != expect {
			t.Errorf("test failed due to expect %v, but got %v", expect, got)
		}
	})
}

func TestGetIntValueFormEnv(t *testing.T) {
	t.Run("env is not set", func(t *testing.T) {
		testEnvNameNotFound := "envnotfound"
		expectFound := false

		_, found := GetIntValueFromEnv(testEnvNameNotFound)

		if found != expectFound {
			t.Errorf("test failed due to expect %v, but got %v", expectFound, found)
		}
	})

	t.Run("env is set in true", func(t *testing.T) {
		testEnvNameFound := "envFound"
		t.Setenv(testEnvNameFound, "10")
		expectFound := true
		expect := 10

		got, found := GetIntValueFromEnv(testEnvNameFound)

		if found != expectFound {
			t.Errorf("test failed due to expect %v, but got %v", expectFound, found)
		}

		if got != expect {
			t.Errorf("test failed due to expect %v, but got %v", expect, got)
		}
	})

	t.Run("env is set with illegal value", func(t *testing.T) {
		testEnvNameIllegal := "envIllegal"
		t.Setenv(testEnvNameIllegal, "illegal")
		expectFound := false

		_, found := GetIntValueFromEnv(testEnvNameIllegal)

		if found != expectFound {
			t.Errorf("test failed due to expect %v, but got %v", expectFound, found)
		}
	})
}

func TestGetDurationValueFormEnv(t *testing.T) {
	t.Run("env is not set", func(t *testing.T) {
		testEnvNameNotFound := "envnotfound"
		expect := 3 * time.Second

		got := GetDurationValueFromEnv(testEnvNameNotFound, 3*time.Second)

		if got != expect {
			t.Errorf("test failed due to expect %v, but got %v", expect, got)
		}
	})

	t.Run("env is set in true", func(t *testing.T) {
		testEnvNameFound := "envFound"
		t.Setenv(testEnvNameFound, "10s")
		expect := 10 * time.Second

		got := GetDurationValueFromEnv(testEnvNameFound, 3*time.Second)

		if got != expect {
			t.Errorf("test failed due to expect %v, but got %v", expect, got)
		}
	})
}

func TestGetStringValueFromEnv(t *testing.T) {
	defaultStringValue := "defaultStringValue"

	t.Run("env is not set", func(t *testing.T) {
		testEnvNameNotFound := "envnotfound"
		expect := defaultStringValue

		got := GetStringValueFromEnv(testEnvNameNotFound, defaultStringValue)

		if got != expect {
			t.Errorf("test failed due to expect %v, but got %v", expect, got)
		}
	})

	t.Run("env is set in true", func(t *testing.T) {
		testEnvNameFound := "envFound"
		t.Setenv(testEnvNameFound, "stringValue")
		expect := "stringValue"

		got := GetStringValueFromEnv(testEnvNameFound, defaultStringValue)
		if got != expect {
			t.Errorf("test failed due to expect %v, but got %v", expect, got)
		}
	})
}
