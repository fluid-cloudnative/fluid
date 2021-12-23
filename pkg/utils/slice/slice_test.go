package slice

import (
	"reflect"
	"testing"
)

func TestContainsString(t *testing.T) {
	src := []string{"aa", "bb", "cc"}
	if !ContainsString(src, "bb") {
		t.Errorf("ContainsString didn't find the string as expected")
	}
}

func TestRemoveString(t *testing.T) {
	tests := []struct {
		testName string
		input    []string
		remove   string
		want     []string
	}{
		{
			testName: "Nil input slice",
			input:    nil,
			remove:   "",
			want:     nil,
		},
		{
			testName: "Slice doesn't contain the string",
			input:    []string{"a", "ab", "cdef"},
			remove:   "NotPresentInSlice",
			want:     []string{"a", "ab", "cdef"},
		},
		{
			testName: "All strings removed, result is nil",
			input:    []string{"a"},
			remove:   "a",
			want:     nil,
		},
		{
			testName: "No modifier func, one string removed",
			input:    []string{"a", "ab", "cdef"},
			remove:   "ab",
			want:     []string{"a", "cdef"},
		},
		{
			testName: "No modifier func, all(three) strings removed",
			input:    []string{"ab", "a", "ab", "cdef", "ab"},
			remove:   "ab",
			want:     []string{"a", "cdef"},
		},
	}
	for _, tt := range tests {
		if got := RemoveString(tt.input, tt.remove); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%v: RemoveString(%v, %q) = %v WANT %v", tt.testName, tt.input, tt.remove, got, tt.want)
		}
	}
}
