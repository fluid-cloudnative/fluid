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
