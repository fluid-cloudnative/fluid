/*
Copyright 2023 The Fluid Author.

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
	"strings"

	units "github.com/docker/go-units"
	"k8s.io/apimachinery/pkg/api/resource"
)

// TransformQuantityToAlluxioUnit transform a given input quantity to another one
// that can be recognized by Alluxio. This is necessary because Alluxio takes decimal byte units(e.g. KB, MB, GB, etc.)
// as binary byte units(e.g. Ki, Mi, Gi)
func TransformQuantityToAlluxioUnit(q *resource.Quantity) (value string) {
	value = q.String()

	if strings.HasSuffix(value, "i") {
		value = strings.ReplaceAll(value, "i", "B")
	}
	return
	// return units.BytesSize(units.BytesSize(float64(q.Value())))

}

// TransfromQuantityToJindoUnit transform a given input quantity to another one
// that can be recognized by Jindo.
func TransformQuantityToJindoUnit(q *resource.Quantity) (value string) {
	value = q.String()
	if strings.HasSuffix(value, "Gi") {
		value = strings.ReplaceAll(value, "Gi", "g")
	}
	if strings.HasSuffix(value, "Mi") {
		value = strings.ReplaceAll(value, "Mi", "m")
	}
	return
}

// TransformQuantityToGooseFSUnit transform a given input quantity to another one
// that can be recognized by GooseFS. This is necessary because GooseFS takes decimal byte units(e.g. KB, MB, GB, etc.)
// as binary byte units(e.g. Ki, Mi, Gi)
func TransformQuantityToGooseFSUnit(q *resource.Quantity) (value string) {
	value = q.String()

	if strings.HasSuffix(value, "i") {
		value = strings.ReplaceAll(value, "i", "B")
	}
	return
	// return units.BytesSize(units.BytesSize(float64(q.Value())))

}

// TransformQuantityToEFCUnit transform a given input quantity to another one
// that can be recognized by EFC. This is necessary because EFC takes decimal byte units(e.g. KB, MB, GB, etc.)
// as binary byte units(e.g. Ki, Mi, Gi)
func TransformQuantityToEFCUnit(q *resource.Quantity) (value string) {
	value = q.String()
	if strings.HasSuffix(value, "i") {
		value = strings.ReplaceAll(value, "i", "B")
	}
	return
}

func TransformEFCUnitToQuantity(value string) (q *resource.Quantity) {
	if strings.HasSuffix(value, "B") {
		value = strings.ReplaceAll(value, "B", "i")
	}
	result := resource.MustParse(value)
	return &result
}

// TransformQuantityToUnits returns a human-readable size in bytes, kibibytes,
// mebibytes, gibibytes, or tebibytes (eg. "44kiB", "17MiB").
func TranformQuantityToUnits(q *resource.Quantity) (value string) {
	// value = q.String()

	// if strings.HasSuffix(value, "i") {
	// 	strings.ReplaceAll(value, "i", "B")
	// }

	return units.BytesSize(float64(q.Value()))
}
