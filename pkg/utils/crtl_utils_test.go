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
	}{
		{&nonnullStr, defaultStr, nonnullStr},
		{nil, defaultStr, defaultStr},
	}
	for _, test := range tests {
		if str := GetOrDefault(test.pstr, test.defaultStr); str != test.expectedStr {
			t.Errorf("expected %s, got %s", test.expectedStr, str)
		}
	}
}

func TestContainsString(t *testing.T)  {
	var aaa, bbb, ccc, ddd, empty = "aaa", "bbb", "ccc", "ddd", ""
	var slice = []string{aaa, bbb, ccc, empty}
	var testCases = []struct {
		slice    []string
		s        string
		expected bool
	} {
		{slice, aaa, true},
		{slice, bbb, true},
		{slice, ccc, true},
		{slice, ddd, false},
		{slice, empty, true},
	}
	for _, tc := range testCases {
		if ret := ContainsString(tc.slice, tc.s); ret != tc.expected {
			t.Errorf("ContainsString(%#v, %s), expected %t， got %t", tc.slice, tc.s, tc.expected, ret)
		}
	}
}

func TestRemoveString(t *testing.T) {
	var aaa, bbb, ccc, ddd, empty = "aaa", "bbb", "ccc", "ddd", ""
	var slice = []string{aaa, bbb, ccc, ddd, empty}
	var testCases = []struct {
		slice    []string
		s        string
		expected []string
	} {
		{slice, aaa, []string{bbb, ccc, ddd, empty}},
		{slice, bbb, []string{aaa, ccc, ddd, empty}},
		{slice, ccc, []string{aaa, bbb, ddd, empty}},
		{slice, ddd, []string{aaa, bbb, ccc, empty}},
		{slice, empty, []string{aaa, bbb, ccc, ddd}},
	}
	var stringSliceEqual = func(a, b []string) bool {
		if len(a) != len(b) || (a == nil) != (b == nil) {
			return false
		}
		for i, s := range a {
			if s != b[i] {
				return false
			}
		}
		return true
	}
	for _, tc := range testCases {
		if result := RemoveString(tc.slice, tc.s); !stringSliceEqual(tc.expected, result) {
			t.Errorf("RemoveString(%#v, %s), expected %#v, got %#v", tc.slice, tc.s, tc.expected, result)
		}
	}
}
