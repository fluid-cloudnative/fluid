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

package testutil

import (
	"reflect"

	"github.com/stretchr/testify/assert"
)

// DeepEqualIgnoringSliceOrder is much like reflect.DeepEqual but ignores order of slices
// in any structs. This is a function only used for unit tests since many objects in Kubernetes
// should be considered same ignoring the slice order(e.g. VolumeMounts, Volumes, EnvVar, etc.)
// NOTE: The func cannot handle recursive composite types like [][]interface, map[X][]interface, etc.
func DeepEqualIgnoringSliceOrder(t assert.TestingT, x interface{}, y interface{}) bool {
	if x == nil || y == nil {
		return x == y
	}
	v1 := reflect.ValueOf(x)
	v2 := reflect.ValueOf(y)
	if v1.Type() != v2.Type() {
		return false
	}

	switch v1.Kind() {
	case reflect.Array:
		return assert.ElementsMatch(t, v1.Interface(), v2.Interface())
	case reflect.Slice:
		return assert.ElementsMatch(t, v1.Interface(), v2.Interface())
	case reflect.Ptr:
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		return DeepEqualIgnoringSliceOrder(t, v1.Elem().Interface(), v2.Elem().Interface())
	case reflect.Struct:
		for i, n := 0, v1.NumField(); i < n; i++ {
			if !DeepEqualIgnoringSliceOrder(t, v1.Field(i).Interface(), v2.Field(i).Interface()) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(x, y)
	}
}
