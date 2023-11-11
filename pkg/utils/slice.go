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

// RemoveDuplicateStr removes duplicate string
func RemoveDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
