/*
Copyright 2021 The Fluid Authors.

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

import "strings"

type FluidUFSScheme string

const (
	// native
	PathScheme   FluidUFSScheme = "local://"
	VolumeScheme FluidUFSScheme = "pvc://"
	// web
	HttpScheme  FluidUFSScheme = "http://"
	HttpsScheme FluidUFSScheme = "https://"
	// ref
	RefSchema FluidUFSScheme = "dataset://"
)

func (fns FluidUFSScheme) String() string {
	return string(fns)
}

func IsFluidNativeScheme(s string) bool {
	if strings.HasPrefix(s, PathScheme.String()) {
		return true
	} else if strings.HasPrefix(s, VolumeScheme.String()) {
		return true
	}

	return false
}

func IsFluidWebScheme(s string) bool {
	if strings.HasPrefix(s, HttpScheme.String()) {
		return true
	} else if strings.HasPrefix(s, HttpsScheme.String()) {
		return true
	}
	return false
}

func IsFluidRefSchema(s string) bool {
	return strings.HasPrefix(s, RefSchema.String())
}
