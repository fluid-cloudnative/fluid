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

// UnionMapsWithOverride unions two maps into one. If either of the maps is empty, return the other one.
// If both maps share the same key, the value in map2 overrides the corresponding value in map1.
func UnionMapsWithOverride(map1 map[string]string, map2 map[string]string) map[string]string {
	if len(map1) == 0 || len(map2) == 0 {
		if len(map1) == 0 {
			return map2
		} else {
			return map1
		}
	}

	retMap := map[string]string{}
	for k, v := range map1 {
		retMap[k] = v
	}

	for k, v := range map2 {
		retMap[k] = v
	}

	return retMap
}
