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

package alluxio

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

var _ = Describe("Alluxio", func() {
	DescribeTable("transformAPIGateway",
		func(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio, expectError bool, shouldMatch bool) {
			engine := &AlluxioEngine{}
			err := engine.transformAPIGateway(runtime, value)

			if expectError {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
				if shouldMatch {
					Expect(runtime.Spec.APIGateway.Enabled).To(Equal(value.APIGateway.Enabled))
				}
			}
		},
		Entry("should sync when runtime enabled and value disabled",
			&datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					APIGateway: datav1alpha1.AlluxioCompTemplateSpec{
						Enabled: true,
					},
				},
			},
			&Alluxio{
				APIGateway: APIGateway{
					Enabled: false,
				},
			},
			false,
			false,
		),
		Entry("should sync when runtime disabled and value enabled",
			&datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					APIGateway: datav1alpha1.AlluxioCompTemplateSpec{
						Enabled: false,
					},
				},
			},
			&Alluxio{
				APIGateway: APIGateway{
					Enabled: true,
				},
			},
			false,
			false,
		),
		Entry("should return error when runtime is nil",
			nil,
			&Alluxio{
				APIGateway: APIGateway{
					Enabled: false,
				},
			},
			true,
			false,
		),
		Entry("should return error when value is nil",
			&datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					APIGateway: datav1alpha1.AlluxioCompTemplateSpec{
						Enabled: true,
					},
				},
			},
			nil,
			true,
			false,
		),
	)
})
