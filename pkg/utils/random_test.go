package utils

import (
	"regexp"
	"testing"
)

func TestRandomString(t *testing.T) {
	testCases := map[string]struct {
		base    []rune
		len     int32
		wantLen int
		regex   *regexp.Regexp
	}{
		"test RandomString case 1": {
			base:    []rune("abcdefghijklmnopqrstuvwxyz"),
			len:     3,
			wantLen: 3,
			regex:   regexp.MustCompile(".*[a-z]+.*"),
		},
		"test RandomString case 2": {
			base:    []rune("0123456789"),
			len:     5,
			wantLen: 5,
			regex:   regexp.MustCompile(".*[0-9]+.*"),
		},
		"test RandomString case 3": {
			base:    []rune("abcdefghijklmnopqrstuvwxyz0123456789"),
			len:     30,
			wantLen: 30,
			regex:   regexp.MustCompile("^[0-9a-z]+$"),
		},
	}

	for k, item := range testCases {
		got := RandomString(item.base, item.len)

		if len(got) != item.wantLen || !item.regex.MatchString(got) {
			t.Errorf("%s test failure, want length:%d,got length:%d, got string:%s",
				k,
				item.wantLen,
				len(got),
				got,
			)
		}
	}
}

func TestRandomAlphaNumberString(t *testing.T) {
	testCases := map[string]struct {
		len     int32
		wantLen int
		regex   *regexp.Regexp
	}{
		"test RandomAlphaNumberString case 1": {
			len:     10,
			wantLen: 10,
			regex:   regexp.MustCompile("^[0-9a-z]+?"),
		},
		"test RandomAlphaNumberString case 2": {
			len:     20,
			wantLen: 20,
			regex:   regexp.MustCompile("^[0-9a-z]+?"),
		},
	}

	for k, item := range testCases {
		got := RandomAlphaNumberString(item.len)
		if len(got) != item.wantLen || !item.regex.MatchString(got) {
			t.Errorf("%s check failure,want length is:%d,got:%d", k, item.wantLen, len(got))
		}
	}
}
