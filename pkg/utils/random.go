package utils

import (
	"math/rand"
)

func RandomString(source []rune, l int32) string {
	res := make([]rune, l)
	for i := range res {
		idx := rand.Intn(len(source))
		res[i] = source[idx]
	}
	return string(res)
}

func RandomAlphaNumberString(l int32) string {
	source := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	return RandomString(source, l)
}
