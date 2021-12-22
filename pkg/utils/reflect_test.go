package utils

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestFieldNameByType(t *testing.T) {
	type testCase struct {
		name        string
		original    interface{}
		target      interface{}
		expect      []string
		expectFound bool
	}

	testcases := []testCase{
		{
			name:     "Original is Ptr",
			original: &corev1.Pod{},
			target:   corev1.Container{},
		}, {
			name:     "Both struct",
			original: corev1.Pod{},
			target:   corev1.Container{},
		}, {
			name:     "targetType struct",
			original: corev1.Pod{},
			target:   []corev1.Container{},
		},
	}

	for _, testcase := range testcases {
		got, found := FieldNameByType(testcase.original, testcase.target)
		if found != testcase.expectFound {
			t.Errorf("testcase %s failed due to expected %b, but got %b", testcase.name, testcase.expectFound, found)
		}

		if !reflect.DeepEqual(got, testcase.expect) {
			t.Errorf("testcase %s failed due to expected %v, but got %v", testcase.name, got, testcase.expect)
		}
	}

}
