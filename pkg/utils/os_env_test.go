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

	"github.com/fluid-cloudnative/fluid/pkg/common"
)

func TestGetEnvByKey(t *testing.T) {
	testCases := map[string]struct {
		key       string
		envKey    string
		value     string
		wantValue string
	}{
		"test get env value by key case 1": {
			key:       "MY_POD_NAMESPACE",
			envKey:    "MY_POD_NAMESPACE",
			value:     common.NamespaceFluidSystem,
			wantValue: common.NamespaceFluidSystem,
		},
		"test get env value by key case 2": {
			key:       "MY_POD_NAMESPACE",
			envKey:    "MY_POD_NAMESPACES",
			value:     common.NamespaceFluidSystem,
			wantValue: "",
		},
	}

	for k, item := range testCases {
		// prepare env value
		t.Setenv(item.key, item.value)
		gotValue, _ := GetEnvByKey(item.envKey)
		if gotValue != item.wantValue {
			t.Errorf("%s check failure, want:%v,got:%v", k, item.wantValue, gotValue)
		}
	}
}

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

func TestCheckValidateEnvName(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"MY_VAR_1", true},
		{"my_var_1", true},
		{"my-var-1", false},
		{"MY-VAR-1", false},
		{"MyVar1", true},
		{"1_MY_VAR", false},
		{"_MY_VAR", true},
	}

	for _, c := range cases {
		err := CheckValidateEnvName(c.in)
		got := (err == nil)
		if got != c.want {
			t.Errorf("CheckValidateEnvName(%q) == %v, want %v, error: %v", c.in, got, c.want, err)
		}
	}
}

func TestConvertDashToUnderscore(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "Single dash", input: "hello-world", expected: "hello_world"},
		{name: "Multiple dashes", input: "hello-world-again", expected: "hello_world_again"},
		{name: "No dash", input: "helloworld", expected: "helloworld"},
		{name: "Empty string", input: "", expected: ""},
		{name: "Only dashes", input: "---", expected: "___"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ConvertDashToUnderscore(tt.input)
			if actual != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, actual)
			}
		})
	}
}
