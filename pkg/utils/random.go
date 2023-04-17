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
