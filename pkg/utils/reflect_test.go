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

func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
