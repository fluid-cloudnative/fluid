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
				"Spec": "v1.PodSpec",
			},
		},
		{
			name:     "Original is Ptr, and search container",
			original: &corev1.Pod{},
			target:   corev1.Container{},
			expect: map[string]string{
				"InitContainers": "[]v1.Container",
				"Containers":     "[]v1.Container",
			},
		}, {
			name:     "Both struct, and search volume",
			original: &corev1.Pod{},
			target:   []corev1.Volume{},
			expect: map[string]string{
				"Volumes": "[]v1.Volume",
			},
		}, {
			name:     "TargetType struct, and search containers",
			original: &corev1.Pod{},
			target:   []corev1.Container{},
			expect: map[string]string{
				"InitContainers": "[]v1.Container",
				"Containers":     "[]v1.Container",
			},
		}, {
			name:     "TargetType struct, and search *int64",
			original: &corev1.Pod{},
			target:   utilpointer.Int64Ptr(1),
			expect: map[string]string{
				"ActiveDeadlineSeconds":         "*int64",
				"FSGroup":                       "*int64",
				"TolerationSeconds":             "*int64",
				"DeletionGracePeriodSeconds":    "*int64",
				"ExpirationSeconds":             "*int64",
				"RunAsUser":                     "*int64",
				"RunAsGroup":                    "*int64",
				"TerminationGracePeriodSeconds": "*int64"},
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

	for k, v := range slice1 {
		if typeValue, found := slice2[k]; found {
			if v.Type().String() != typeValue {
				diff = append(diff, k)
			}
		} else {
			diff = append(diff, k)
		}
	}

	return diff
}
