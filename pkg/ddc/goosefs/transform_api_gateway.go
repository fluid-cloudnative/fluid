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

package goosefs

import (
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

// transformAPIGateway decide whether to enable APIGateway in value according to GooseFSRuntime
func (e *GooseFSEngine) transformAPIGateway(runtime *datav1alpha1.GooseFSRuntime, value *GooseFS) (err error) {
	if runtime == nil || value == nil {
		err = fmt.Errorf("cannot transform because runtime or value will lead to nil pointer")
		return
	}
	value.APIGateway.Enabled = runtime.Spec.APIGateway.Enabled
	return
}
