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

package goosefs

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func TestTransformPermission(t *testing.T) {

	keys := []string{
		"goosefs.master.security.impersonation.root.users",
		"goosefs.master.security.impersonation.root.groups",
		"goosefs.security.authorization.permission.enabled",
	}

	var tests = []struct {
		runtime *datav1alpha1.GooseFSRuntime
		value   *GooseFS
		expect  map[string]string
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Fuse: datav1alpha1.GooseFSFuseSpec{},
			},
		}, &GooseFS{}, map[string]string{
			"goosefs.master.security.impersonation.root.users":  "*",
			"goosefs.master.security.impersonation.root.groups": "*",
			"goosefs.security.authorization.permission.enabled": "false",
		}},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{}
		engine.transformPermission(test.runtime, test.value)
		for _, key := range keys {
			if test.value.Properties[key] != test.expect[key] {
				t.Errorf("The key %s expected %s, got %s", key, test.value.Properties[key], test.expect[key])
			}
		}

	}
}
