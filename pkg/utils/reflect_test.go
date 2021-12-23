package utils

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	utilpointer "k8s.io/utils/pointer"
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
			name:     "Original is Ptr, and search container",
			original: &corev1.Pod{},
			target:   corev1.Container{},
			expect: []string{
				"InitContainers", "Containers",
			},
		}, {
			name:     "Both struct, and search volume",
			original: &corev1.Pod{},
			target:   []corev1.Volume{},
			expect: []string{
				"Volumes",
			},
		}, {
			name:     "TargetType struct, and search containers",
			original: &corev1.Pod{},
			target:   []corev1.Container{},
			expect: []string{
				"InitContainers", "Containers",
			},
		}, {
			name:     "TargetType struct, and search *int64",
			original: &corev1.Pod{},
			target:   utilpointer.Int64Ptr(1),
			expect: []string{
				"ActiveDeadlineSeconds", "FSGroup", "TolerationSeconds", "DeletionGracePeriodSeconds", "ExpirationSeconds", "RunAsUser", "RunAsGroup", "TerminationGracePeriodSeconds",
			},
		},
	}

	for _, testcase := range testcases {
		got := FieldNameByType(testcase.original, testcase.target)

		result := difference(got, testcase.expect)
		if len(result) > 0 {
			t.Errorf("testcase %s failed due to expected %v, but got %v", testcase.name, testcase.expect, got)
		}

	}

}

func difference(slice1 []string, slice2 []string) []string {
	var diff []string

	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}

			if !found {
				diff = append(diff, s1)
			}
		}
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return diff
}
