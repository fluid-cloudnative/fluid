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
