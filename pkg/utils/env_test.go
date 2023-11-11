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
