/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

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

func TestRandomReplace(t *testing.T) {
	testCases := map[string]struct {
		input   string
		len     int
		wantLen int
	}{
		"test RandomReplace case 1": {
			input:   "a",
			len:     5,
			wantLen: 5,
		},
		"test RandomReplace case 2": {
			input:   "abcdef",
			len:     3,
			wantLen: len("abcdef"),
		},
	}

	for k, item := range testCases {
		got := RandomReplacePrefix(item.input, item.len)
		if len(got) != item.wantLen || item.input == got {
			t.Errorf("%s check failure,want length is:%d,got:%d", k, item.wantLen, len(got))
		}
	}
}

func TestReplacePrefix(t *testing.T) {
	testCases := map[string]struct {
		input   string
		replace string
		want    string
	}{
		"test RandomReplace case 1": {
			input:   "a",
			replace: "abc",
			want:    "abc",
		},
		"test RandomReplace case 2": {
			input:   "abcdef",
			replace: "efg",
			want:    "efg-ef",
		},
	}

	for k, item := range testCases {
		got := ReplacePrefix(item.input, item.replace)
		if item.want != got {
			t.Errorf("%s check failure,want is:%s,got:%s", k, item.want, got)
		}
	}
}
