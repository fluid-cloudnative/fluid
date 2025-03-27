/*
Copyright 2024 The Fluid Authors.

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

package validation

import (
	"fmt"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
)

const invalidPartOfPathErrMsg string = "every part of the path shuold follow the relaxed DNS (RFC 1123) rule which additionally allows upper case alphabetic character and character '_'"

func isValidPartsOfPath(partsOfPath []string) error {
	for _, part := range partsOfPath {
		// Convert characters to lowercase and replace underscores with hyphens
		part = strings.ToLower(part)
		part = strings.Replace(part, "_", "-", -1)

		// If the component fails the DNS 1123 conformity test, the function returns an error
		if len(validation.IsDNS1123Label(part)) > 0 {
			return fmt.Errorf(invalidPartOfPathErrMsg)
		}
	}
	return nil
}

const invalidMountRootErrMsgFmt string = "invalid mount root path '%s': %s"

func IsValidMountRoot(path string) error {
	if len(path) == 0 {
		return fmt.Errorf(invalidMountRootErrMsgFmt, path, "the mount root path is empty")
	}
	if !filepath.IsAbs(path) {
		return fmt.Errorf(invalidMountRootErrMsgFmt, path, "the mount root path must be an absolute path")
	}
	// Normalize the path and split it into components
	// The path is an absolute path and to avoid an empty part, we omit the first '/'
	parts := strings.Split(filepath.Clean(path)[1:], string(filepath.Separator))

	// Validate every part of the path
	if err := isValidPartsOfPath(parts); err != nil {
		return fmt.Errorf(invalidMountRootErrMsgFmt, path, err.Error())
	}

	// If all components pass the check, the function returns nil
	return nil
}

func IsValidMountPath(path string) error {
	if len(path) == 0 {
		return fmt.Errorf("the mount path is empty")
	}
	// Normalize the path and split it into components
	path = filepath.Clean(path)
	if filepath.IsAbs(path) {
		// The path is an absolute path and to avoid an empty part, we omit the first '/'
		path = path[1:]
	}
	parts := strings.Split(path, string(filepath.Separator))

	// Validate every part of the path
	if err := isValidPartsOfPath(parts); err != nil {
		return err
	}

	// If all components pass the check, the function returns nil
	return nil
}
