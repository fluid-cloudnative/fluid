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

	for _, testCase := range testCases {
		if testCase.expect != IsDeprecated(testCase.Err) {
			t.Errorf("testCase %s: expected %v ,got %v", testCase.Name, testCase.expect, IsDeprecated(testCase.Err))
		}
	}
}
