package utils

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestFieldNameByType(t *testing.T) {
	type testCase struct {
		name     string
		original interface{}
		target   interface{}
		expect   []string
	}

	testcases := []testCase{
		{
			name:     "Original is Ptr, and search pod",
			original: &corev1.Pod{},
			target:   corev1.PodSpec{},
			expect: []string{
				"Spec",
			},
		},
		{
			name:     "Original is Ptr",
			original: &corev1.Pod{},
			target:   corev1.Container{},
			expect: []string{
				"InitContainers", "Containers",
			},
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
		got := FieldNameByType(testcase.original, testcase.target)

		if !reflect.DeepEqual(got, testcase.expect) {
			t.Errorf("testcase %s failed due to expected %v, but got %v", testcase.name, testcase.expect, got)
		}
	}

}
