/*
Copyright 2021 The Fluid Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package reflect

import (
	"fmt"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
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

func TestContainersFieldNameFromObject(t *testing.T) {
	type testCase struct {
		name           string
		object         interface{}
		nominateName   string
		excludeMatches []string
		expect         string
		wantErr        error
	}

	testcases := []testCase{
		{
			name:           "with Exclude names",
			object:         &appsv1.DaemonSet{},
			nominateName:   "",
			excludeMatches: []string{"init"},
			expect:         "Containers",
			wantErr:        nil,
		}, {
			name:           "with nominate names",
			object:         &appsv1.DaemonSet{},
			nominateName:   "InitContainers",
			excludeMatches: []string{""},
			expect:         "InitContainers",
			wantErr:        nil,
		}, {
			name:           "Empty",
			object:         &appsv1.DaemonSet{},
			nominateName:   "",
			excludeMatches: []string{""},
			expect:         "",
			wantErr:        fmt.Errorf("can't determine the names in [InitContainers Containers]"),
		},
	}

	for _, testcase := range testcases {
		got, err := ContainersFieldNameFromObject(testcase.object, testcase.nominateName, testcase.excludeMatches)
		if testcase.wantErr != err {
			if testcase.wantErr != nil && err != nil {
				if testcase.wantErr.Error() != err.Error() {
					t.Errorf("testcase %s failed due to expected err %v, but got err %v", testcase.name, testcase.wantErr, err)
				}
			}

		}

		if testcase.expect != got {
			t.Errorf("testcase %s failed due to expected %v, but got %v", testcase.name, testcase.expect, got)
		}

	}

}

func TestVolumesFieldNameFromObject(t *testing.T) {
	type testCase struct {
		name           string
		object         interface{}
		nominateName   string
		excludeMatches []string
		expect         string
		wantErr        error
	}

	testcases := []testCase{
		{
			name:           "with Exclude names",
			object:         &appsv1.DaemonSet{},
			nominateName:   "",
			excludeMatches: []string{"init"},
			expect:         "Volumes",
			wantErr:        nil,
		}, {
			name:           "with nominate names",
			object:         &appsv1.DaemonSet{},
			nominateName:   "Volumes",
			excludeMatches: []string{""},
			expect:         "Volumes",
			wantErr:        nil,
		}, {
			name:           "Empty",
			object:         &appsv1.DaemonSet{},
			nominateName:   "",
			excludeMatches: []string{""},
			expect:         "",
			wantErr:        nil,
		},
	}

	for _, testcase := range testcases {
		got, err := VolumesFieldNameFromObject(testcase.object, testcase.nominateName, testcase.excludeMatches)
		if testcase.wantErr != err {
			if testcase.wantErr != nil && err != nil {
				if testcase.wantErr.Error() != err.Error() {
					t.Errorf("testcase %s failed due to expected err %v, but got err %v", testcase.name, testcase.wantErr, err)
				}
			}

		}

		if testcase.expect != got {
			t.Errorf("testcase %s failed due to expected %v, but got %v", testcase.name, testcase.expect, got)
		}

	}
}
