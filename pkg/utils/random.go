/*

Copyright 2023 The Fluid Authors.

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

package utils

import (
	"math/rand"
)

// RandomString returns a string of length l which is made up of runes randomly selected from `source`.
func RandomString(source []rune, l int32) string {
	res := make([]rune, l)
	for i := range res {
		idx := rand.Intn(len(source))
		res[i] = source[idx]
	}
	return string(res)
}

// RandomAlphaNumberString returns a string of length l
// which is made up of runes randomly selected from [0-9a-z].
func RandomAlphaNumberString(l int32) string {
	source := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	return RandomString(source, l)
}

func RandomReplacePrefix(input string, l int) (output string) {
	length := len(input)
	prefix := RandomAlphaNumberString(int32(l))
	if length <= l {
		output = prefix
	} else {
		output = prefix + input[l:]
	}

	return
}

// ReplacePrefix replaces the  input with suffix string
func ReplacePrefix(input, suffix string) (output string) {
	if len(input)+1 <= len(suffix) {
		output = suffix
	} else {
		output = suffix + "-" + input[len(suffix)+1:]
	}
	return
}
