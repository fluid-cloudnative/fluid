package utils

// ContainsAll checks if a map contains all the elements in a slice
func ContainsAll(m map[string]string, slice []string) bool {
	if len(slice) == 0 {
		return true
	}
	if len(m) == 0 {
		return false
	}
	for _, elem := range slice {
		if _, ok := m[elem]; !ok {
			return false
		}
	}
	return true
}

// ContainsKey judges whether the arr contains the elem.
func ContainsKey(elems map[string]string, target string) bool {
	for elem := range elems {
		if elem == target {
			return true
		}
	}
	return false
}
