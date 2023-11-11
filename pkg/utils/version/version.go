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

package version

import (
	"strings"

	versionutil "k8s.io/apimachinery/pkg/util/version"
)

const (
	releasePrefix = "release-"
)

func RuntimeVersion(str string) (*versionutil.Version, error) {
	// return versionutil.ParseSemantic(str)
	return versionutil.ParseGeneric(strings.TrimPrefix(strings.ToLower(str), releasePrefix))
}

// Compare compares v against a version string (which will be parsed as either Semantic
// or non-Semantic depending on v). On success it returns -1 if v is less than other, 1 if
// it is greater than other, or 0 if they are equal.
func Compare(current, other string) (compare int, err error) {
	v1, err := RuntimeVersion(current)
	if err != nil {
		return
	}

	return v1.Compare(other)
}
