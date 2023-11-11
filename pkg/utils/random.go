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
