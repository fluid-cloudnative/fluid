package reflect

import (
	"fmt"
	ref "reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
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

func TestContainersValueFromObject(t *testing.T) {
	type testCase struct {
		name           string
		object         interface{}
		nominateName   string
		excludeMatches []string
		expectName     string
		expectType     string
		wantErr        error
	}

	testcases := []testCase{
		{
			name:           "with Exclude names",
			object:         &appsv1.DaemonSet{},
			nominateName:   "",
			excludeMatches: []string{"init"},
			expectName:     "Containers",
			expectType:     "[]v1.Container",
			wantErr:        nil,
		}, {
			name:           "with nominate names",
			object:         &appsv1.DaemonSet{},
			nominateName:   "InitContainers",
			excludeMatches: []string{""},
			expectName:     "InitContainers",
			expectType:     "[]v1.Container",
			wantErr:        nil,
		}, {
			name:           "Empty",
			object:         &appsv1.DaemonSet{},
			nominateName:   "",
			excludeMatches: []string{""},
			expectName:     "",
			expectType:     "[]v1.Container",
			wantErr:        fmt.Errorf("can't determine the names in [InitContainers Containers]"),
		},
	}

	for _, testcase := range testcases {
		name, value, err := ContainersValueFromObject(testcase.object, testcase.nominateName, testcase.excludeMatches)
		if testcase.wantErr != err {
			if testcase.wantErr != nil && err != nil {
				if len(testcase.wantErr.Error()) != len(err.Error()) {
					t.Errorf("testcase %s failed due to expected err %v, but got err %v", testcase.name, testcase.wantErr, err)
				}
			}

		}

		if err == nil {
			if testcase.expectName != name {
				t.Errorf("testcase %s failed due to expected %v, but got %v", testcase.name, testcase.expectName, name)
			}

			// if !value.IsValid() {
			// 	sliceType := ref.TypeOf(value)
			// 	value.Set(ref.MakeSlice(sliceType, 0, 0))
			// }

			if testcase.expectType != value.Type().String() {
				t.Errorf("testcase %s failed due to expected type %v, but got type %v", testcase.name, testcase.expectType, value.Type().String())
			}
		}

	}

}

func TestVolumesValueFromObject(t *testing.T) {
	type testCase struct {
		name           string
		object         interface{}
		nominateName   string
		excludeMatches []string
		expect         string
		expectType     string
		wantErr        error
	}

	testcases := []testCase{
		{
			name:           "with Exclude names",
			object:         &appsv1.DaemonSet{},
			nominateName:   "",
			excludeMatches: []string{"init"},
			expect:         "Volumes",
			expectType:     "[]v1.Volume",
			wantErr:        nil,
		}, {
			name:           "with nominate names",
			object:         &appsv1.DaemonSet{},
			nominateName:   "Volumes",
			excludeMatches: []string{""},
			expect:         "Volumes",
			expectType:     "[]v1.Volume",
			wantErr:        nil,
		}, {
			name:           "Empty",
			object:         &appsv1.DaemonSet{},
			nominateName:   "",
			excludeMatches: []string{""},
			expect:         "Volumes",
			expectType:     "[]v1.Volume",
			wantErr:        nil,
		},
	}

	for _, testcase := range testcases {
		name, value, err := VolumesValueFromObject(testcase.object, testcase.nominateName, testcase.excludeMatches)
		if testcase.wantErr != err {
			if testcase.wantErr != nil && err != nil {
				if testcase.wantErr.Error() != err.Error() {
					t.Errorf("testcase %s failed due to expected err %v, but got err %v", testcase.name, testcase.wantErr, err)
				}
			}

		}

		if testcase.expect != name {
			t.Errorf("testcase %s failed due to expected %v, but got %v", testcase.name, testcase.expect, name)
		}

		if testcase.expectType != value.Type().String() {
			t.Errorf("testcase %s failed due to expected type %v, but got type %v", testcase.name, testcase.expectType, value.Type().String())
		}

	}
}
