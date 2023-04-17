package utils

import (
	"fmt"
	"testing"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestNoRequeue(t *testing.T) {
	result, err := NoRequeue()
	if err != nil {
		t.Errorf("err should be nil")

	}
	if result.Requeue != false || result.RequeueAfter != 0 {
		t.Errorf("resuld should be ctrl.Result{}")
	}
}

func TestRequeueAfterInterval(t *testing.T) {
	testCases := map[string]string{
		"test calculate duration case 1": "5m10s",
		"test calculate duration case 2": "6h7m0s",
		"test calculate duration case 3": "2h7m2s",
	}

	for k, item := range testCases {
		mockDuration, err := time.ParseDuration(item)
		if err != nil {
			t.Errorf("%s is not suitable", k)
		}
		result, err := RequeueAfterInterval(mockDuration)
		if err != nil {
			t.Errorf("err should be nil")
		}
		if result.RequeueAfter.String() != item {
			t.Errorf("%s is wrong, want %s, get %s", k, item, result.RequeueAfter.String())
		}

	}
}

func TestRequeueImmediately(t *testing.T) {
	result, err := RequeueImmediately()
	if err != nil {
		t.Errorf("err should be nil")
	}
	if result.Requeue != true || result.RequeueAfter != 0 {
		t.Errorf("should requeue immediately")
	}
}

func TestRequeueIfError(t *testing.T) {
	var testcases = []error{
		fmt.Errorf("err1"),
		fmt.Errorf("err2"),
		fmt.Errorf("err3"),
		nil,
	}
	for _, testcase := range testcases {
		result, err := RequeueIfError(testcase)
		if err != testcase {
			t.Errorf("should not change the err")
		}
		if result.Requeue != false || result.RequeueAfter != 0 {
			t.Errorf("resuld should be ctrl.Result{}")
		}
	}

}

func TestRequeueImmediatelyUnlessGenerationChanged(t *testing.T) {
	var tests = []struct {
		prevGeneration int64
		curGeneration  int64
	}{
		{
			prevGeneration: 35,
			curGeneration:  35,
		},
		{
			prevGeneration: 35,
			curGeneration:  34,
		},
	}
	for _, test := range tests {
		result, err := RequeueImmediatelyUnlessGenerationChanged(test.prevGeneration, test.curGeneration)
		if test.prevGeneration == test.curGeneration {
			if err != nil {
				t.Errorf("err should be nil if prevGeneration == test.curGeneration")
			}
			if result.Requeue != true || result.RequeueAfter != 0 {
				t.Errorf("should requeue immediately if prevGeneration == test.curGeneration")
			}
		} else {
			if err != nil {
				t.Errorf("err should be nil if prevGeneration != test.curGeneration")
			}
			if result.Requeue != false || result.RequeueAfter != 0 {
				t.Errorf("resuld should be ctrl.Result{} != if prevGeneration ！= test.curGeneration")
			}
		}
	}
}

func TestGetOrDefault(t *testing.T) {
	var defaultStr = "default string"
	var nonnullStr = "non-null string"
	var tests = []struct {
		pstr        *string
		defaultStr  string
		expectedStr string
	}{
		{&nonnullStr, defaultStr, nonnullStr},
		{nil, defaultStr, defaultStr},
	}
	for _, test := range tests {
		if str := GetOrDefault(test.pstr, test.defaultStr); str != test.expectedStr {
			t.Errorf("expected %s, got %s", test.expectedStr, str)
		}
	}
}

func TestContainsString(t *testing.T) {
	var aaa, bbb, ccc, ddd, empty = "aaa", "bbb", "ccc", "ddd", ""
	var slice = []string{aaa, bbb, ccc, empty}
	var testCases = []struct {
		slice    []string
		s        string
		expected bool
	}{
		{slice, aaa, true},
		{slice, bbb, true},
		{slice, ccc, true},
		{slice, ddd, false},
		{slice, empty, true},
	}
	for _, tc := range testCases {
		if ret := ContainsString(tc.slice, tc.s); ret != tc.expected {
			t.Errorf("ContainsString(%#v, %s), expected %t， got %t", tc.slice, tc.s, tc.expected, ret)
		}
	}
}

func TestContainsOwners(t *testing.T) {
	var testCases = map[string]struct {
		owners   []metav1.OwnerReference
		dataset  *datav1alpha1.Dataset
		expected bool
	}{
		"test calculate duration case 1": {
			owners:   []metav1.OwnerReference{},
			dataset:  &datav1alpha1.Dataset{},
			expected: false,
		},
	}
	for k, tc := range testCases {
		if ret := ContainsOwners(tc.owners, tc.dataset); ret != tc.expected {
			t.Errorf("%s check failure, want %t, get %t", k, tc.expected, ret)
		}
	}
}

func TestRemoveString(t *testing.T) {
	var aaa, bbb, ccc, ddd, empty = "aaa", "bbb", "ccc", "ddd", ""
	var slice = []string{aaa, bbb, ccc, ddd, empty}
	var testCases = []struct {
		slice    []string
		s        string
		expected []string
	}{
		{slice, aaa, []string{bbb, ccc, ddd, empty}},
		{slice, bbb, []string{aaa, ccc, ddd, empty}},
		{slice, ccc, []string{aaa, bbb, ddd, empty}},
		{slice, ddd, []string{aaa, bbb, ccc, empty}},
		{slice, empty, []string{aaa, bbb, ccc, ddd}},
	}
	var stringSliceEqual = func(a, b []string) bool {
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
		if result := RemoveString(tc.slice, tc.s); !stringSliceEqual(tc.expected, result) {
			t.Errorf("RemoveString(%#v, %s), expected %#v, got %#v", tc.slice, tc.s, tc.expected, result)
		}
	}
}

func TestHasDeletionTimestamp(t *testing.T) {
	var pod = corev1.Pod{}
	if HasDeletionTimestamp(pod.ObjectMeta) {
		t.Errorf("result of checking if the DeletionTimestamp exists is wrong")
	}
	mockTime := metav1.NewTime(time.Now())
	pod.DeletionTimestamp = &mockTime
	if !HasDeletionTimestamp(pod.ObjectMeta) {
		t.Errorf("result of checking if the DeletionTimestamp exists is wrong")
	}
}

func TestCalculateDuration(t *testing.T) {
	startTime := time.Now()
	var result string
	testCases := map[string]string{
		"test calculate duration case 1": "5m10s",
		"test calculate duration case 2": "6h7m0s",
		"test calculate duration case 3": "2h7m2s",
	}

	for k, item := range testCases {
		mockDuration, err := time.ParseDuration(item)
		if err != nil {
			t.Errorf("testcase %s is not suitable", item)
		}
		finishTime := startTime.Add(mockDuration)
		result = CalculateDuration(startTime, finishTime)
		if result != item {
			t.Errorf("%s check failure, want %s, get %s", k, item, result)
		}

	}
}

func TestContainsSubString(t *testing.T) {
	type args struct {
		slice []string
		s     string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test",
			args: args{
				slice: []string{"aaa"},
				s:     "a",
			},
			want: true,
		},
		{
			name: "test-b",
			args: args{
				slice: []string{"aaa"},
				s:     "b",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsSubString(tt.args.slice, tt.args.s); got != tt.want {
				t.Errorf("ContainsSubString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIgnoreAlreadyExists(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name: "already_exists_error",
			err: apierrs.NewAlreadyExists(schema.GroupResource{
				Group:    "",
				Resource: "pod",
			}, "mypod"),
			wantErr: false,
		},
		{
			name: "not_found_error",
			err: apierrs.NewNotFound(schema.GroupResource{
				Group:    "",
				Resource: "pod",
			}, "mypod"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := IgnoreAlreadyExists(tt.err); (err != nil) != tt.wantErr {
				t.Errorf("IgnoreAlreadyExists() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIgnoreNotFound(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name: "already_exists_error",
			err: apierrs.NewAlreadyExists(schema.GroupResource{
				Group:    "",
				Resource: "pod",
			}, "mypod"),
			wantErr: true,
		},
		{
			name: "not_found_error",
			err: apierrs.NewNotFound(schema.GroupResource{
				Group:    "",
				Resource: "pod",
			}, "mypod"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := IgnoreNotFound(tt.err); (err != nil) != tt.wantErr {
				t.Errorf("IgnoreNotFound() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIgnoreNoKindMatchError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name: "not_found_error",
			err: apierrs.NewNotFound(schema.GroupResource{
				Group:    "",
				Resource: "pod",
			}, "mypod"),
			wantErr: true,
		},
		{
			name: "no_kind_match_error",
			err: &apimeta.NoKindMatchError{
				GroupKind: schema.GroupKind{
					Group: "data.fluid.io",
					Kind:  "AlluxioRuntime",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := IgnoreNoKindMatchError(tt.err); (err != nil) != tt.wantErr {
				t.Errorf("IgnoreNoKindMatchError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
