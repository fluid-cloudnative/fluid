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
