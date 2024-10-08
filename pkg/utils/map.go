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
	"cmp"
	"slices"
)

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

// UnionMapsWithOverride unions two maps into one. If both maps are empty or nil, return an empty map.
// If both maps share the same key, the value in map2 overrides the value in map1.
func UnionMapsWithOverride(map1 map[string]string, map2 map[string]string) map[string]string {
	retMap := map[string]string{}
	for k, v := range map1 {
		retMap[k] = v
	}

	for k, v := range map2 {
		retMap[k] = v
	}

	return retMap
}

// IntersectIntegerSets returns the intersection of integer set 1 and set 2.
func IntersectIntegerSets(map1 map[int]bool, map2 map[int]bool) map[int]bool {
	ret := map[int]bool{}
	if len(map1) == 0 || len(map2) == 0 {
		return ret
	}

	for elem := range map1 {
		if _, exists := map2[elem]; exists {
			ret[elem] = true
		}
	}

	return ret
}

// SetValueIfKeyAbsent sets value when key is not found in the map.
func SetValueIfKeyAbsent(m map[string]string, key string, value string) {
	if _, found := m[key]; !found {
		m[key] = value
	}
}

// OrderedKeys returns a sorted slice with all keys defined in the map.
func OrderedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	keys := []K{}
	for k := range m {
		keys = append(keys, k)
	}

	slices.Sort(keys)

	return keys
}
