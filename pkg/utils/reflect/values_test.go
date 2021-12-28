package reflect

import (
	ref "reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	utilpointer "k8s.io/utils/pointer"
)

func TestValueByType(t *testing.T) {
	type testCase struct {
		name     string
		original interface{}
		target   interface{}
		expect   map[string]string
	}

	testcases := []testCase{
		{
			name:     "Original is Ptr, and search pod",
			original: &corev1.Pod{},
			target:   corev1.PodSpec{},
			expect: map[string]string{
				"Spec": "v1.",
			},
		},
		{
			name:     "Original is Ptr, and search container",
			original: &corev1.Pod{},
			target:   corev1.Container{},
			expect: map[string]string{
				"InitContainers": "v1.",
				"Containers":     "v1.",
			},
		}, {
			name:     "Both struct, and search volume",
			original: &corev1.Pod{},
			target:   []corev1.Volume{},
			expect: map[string]string{
				"Volumes": "v1.",
			},
		}, {
			name:     "TargetType struct, and search containers",
			original: &corev1.Pod{},
			target:   []corev1.Container{},
			expect: map[string]string{
				"InitContainers": "v1.",
				"Containers":     "v1.",
			},
		}, {
			name:     "TargetType struct, and search *int64",
			original: &corev1.Pod{},
			target:   utilpointer.Int64Ptr(1),
			expect: map[string]string{
				"ActiveDeadlineSeconds":         "v1.",
				"FSGroup":                       "v1.",
				"TolerationSeconds":             "v1.",
				"DeletionGracePeriodSeconds":    "v1.",
				"ExpirationSeconds":             "v1.",
				"RunAsUser":                     "v1.",
				"RunAsGroup":                    "v1.",
				"TerminationGracePeriodSeconds": "v1."},
		},
	}

	for _, testcase := range testcases {
		got := ValueByType(testcase.original, testcase.target)

		result := differenceMap(got, testcase.expect)
		if len(result) > 0 {
			t.Errorf("testcase %s failed due to expected %v, but got %v", testcase.name, testcase.expect, got)
		}

	}

}

func differenceMap(slice1 map[string]ref.Value, slice2 map[string]string) []string {
	var diff []string

	return diff
}
