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

package common

import (
	"strconv"
)

// HostPIDEnabled check if HostPID is true for runtime fuse pod.
func HostPIDEnabled(annotations map[string]string) bool {
	if annotations == nil {
		return false
	}
	value, exist := annotations[RuntimeFuseHostPIDKey]
	if !exist {
		return false
	}
	enabled, err := strconv.ParseBool(value)
	// If parse failed, return false
	if err != nil {
		return false
	}
	return enabled
}
