package utils

import (
	"testing"
)

func TestGetOrDefault(t *testing.T) {
	var defaultStr = "default string"
	var nonnullStr = "non-null string"
	var tests = []struct {
		pstr        *string
		defaultStr  string
		expectedStr string
	} {
		{&nonnullStr, defaultStr, nonnullStr},
		{nil, defaultStr, defaultStr},
	}

	for _, test := range tests {
		if str := GetOrDefault(test.pstr, test.defaultStr); str != test.expectedStr {
			t.Errorf("expected %s, got %s", test.expectedStr, str)
		}
	}
}