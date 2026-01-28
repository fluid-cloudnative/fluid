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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

var _ = Describe("TransformPermission", func() {
	type testCase struct {
		runtime *datav1alpha1.GooseFSRuntime
		value   *GooseFS
		expect  map[string]string
	}

	DescribeTable("should transform permission properties correctly",
		func(tc testCase) {
			keys := []string{
				"goosefs.master.security.impersonation.root.users",
				"goosefs.master.security.impersonation.root.groups",
				"goosefs.security.authorization.permission.enabled",
			}

			engine := &GooseFSEngine{}
			engine.transformPermission(tc.runtime, tc.value)

			for _, key := range keys {
				Expect(tc.value.Properties[key]).To(Equal(tc.expect[key]))
			}
		},
		Entry("default fuse spec",
			testCase{
				runtime: &datav1alpha1.GooseFSRuntime{
					Spec: datav1alpha1.GooseFSRuntimeSpec{
						Fuse: datav1alpha1.GooseFSFuseSpec{},
					},
				},
				value: &GooseFS{},
				expect: map[string]string{
					"goosefs.master.security.impersonation.root.users":  "*",
					"goosefs.master.security.impersonation.root.groups": "*",
					"goosefs.security.authorization.permission.enabled": "false",
				},
			},
		),
	)
})
