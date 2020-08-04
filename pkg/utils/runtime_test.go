package utils

import (
	data "github.com/cloudnativefluid/fluid/api/v1alpha1"
	"github.com/cloudnativefluid/fluid/pkg/common"
	"testing"
)

func TestAddRuntimesIfNotExist(t *testing.T) {
	var runtime1 = data.Runtime{
		Name:     "imagenet",
		Category: common.AccelerateCategory,
	}
	var runtime2 = data.Runtime{
		Name:     "mock-name",
		Category: "mock-category",
	}
	var runtime3 = data.Runtime{
		Name:     "cifar10",
		Category: common.AccelerateCategory,
	}
	var testCases = []struct {
		description string
		runtimes    []data.Runtime
		newRuntime  data.Runtime
		expected    []data.Runtime
	}{
		{"add runtime to an empty slices successfully",
			[]data.Runtime{}, runtime1, []data.Runtime{runtime1}},
		{"duplicate runtime will not be added",
			[]data.Runtime{runtime1}, runtime1, []data.Runtime{runtime1}},
		{"add runtime of different name and category successfully",
			[]data.Runtime{runtime1}, runtime2, []data.Runtime{runtime1, runtime2}},
		{"runtime of the same category but different name will not be added",
			[]data.Runtime{runtime1}, runtime3, []data.Runtime{runtime1}},
	}
	var runtimeSliceEqual = func(a, b []data.Runtime) bool {
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
		if updatedRuntimes := AddRuntimesIfNotExist(tc.runtimes, tc.newRuntime); !runtimeSliceEqual(tc.expected, updatedRuntimes) {
			t.Errorf("%s, expected %#v, got %#v",
				tc.description, tc.expected, updatedRuntimes)
		}
	}
}
