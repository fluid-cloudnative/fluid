package utils

import (
	"os"
	"testing"
)

func TestMountRootWithEnvSet(t *testing.T) {
	var testCases = []struct {
		input    string
		expected string
	}{
		{"/var/lib/mymount", "/var/lib/mymount"},
	}
	for _, tc := range testCases {
		os.Setenv(MountRoot, tc.input)
		if tc.expected != GetMountRoot() {
			t.Errorf("expected %#v, got %#v",
				tc.expected, GetMountRoot())
		}
	}
}

func TestMountRootWithoutEnvSet(t *testing.T) {
	var testCases = []struct {
		input    string
		expected string
	}{
		{"/var/lib/mymount", ""},
	}
	for _, tc := range testCases {
		if tc.expected != GetMountRoot() {
			t.Errorf("expected %#v, got %#v",
				tc.expected, GetMountRoot())
		}
	}
}
