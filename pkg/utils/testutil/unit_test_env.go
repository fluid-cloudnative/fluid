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
	"os"
	"strings"
)

const FluidUnitTestEnv = "FLUID_UNIT_TEST"

func init() {
	// Automatically detect if running inside a test binary
	if strings.HasSuffix(os.Args[0], ".test") || strings.HasSuffix(os.Args[0], ".test.exe") {
		os.Setenv(FluidUnitTestEnv, "true")
	}
}

func IsUnitTest() bool {
	_, exists := os.LookupEnv(FluidUnitTestEnv)
	return exists
}
