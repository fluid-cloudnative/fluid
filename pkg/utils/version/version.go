/*
Copyright 2022 The Fluid Authors.

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
