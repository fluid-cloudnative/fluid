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
