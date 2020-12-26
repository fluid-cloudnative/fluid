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
