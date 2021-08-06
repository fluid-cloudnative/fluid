package utils

// FillSliceWithString fills a slice with repeated given string
func FillSliceWithString(str string, num int) *[]string {
	retSlice := make([]string, num)

	for i := 0; i < num; i++ {
		retSlice[i] = str
	}

	return &retSlice
}

// SubtractString returns the subtraction between two string slice
func SubtractString(x []string, y []string) []string {
	if len(x) == 0 {
		return []string{}
	}

	if len(y) == 0 {
		return x
	}

	var slice []string
	hash := map[string]struct{}{}

	for _, v := range x {
		hash[v] = struct{}{}
	}

	for _, v := range y {
		_, ok := hash[v]
		if ok {
			delete(hash, v)
		}
	}

	for _, v := range x {
		_, ok := hash[v]
		if ok {
			slice = append(slice, v)
		}
	}

	return slice
}
