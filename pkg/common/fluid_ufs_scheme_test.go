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

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCommon(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Common Suite")
}

var _ = Describe("IsFluidNativeScheme", func() {
	DescribeTable("should return correct bool for endpoint",
		func(endpoint string, want bool) {
			Expect(IsFluidNativeScheme(endpoint)).To(Equal(want))
		},
		Entry("pvc://mnt/fluid/data", "pvc://mnt/fluid/data", true),
		Entry("local://mnt/fluid/data", "local://mnt/fluid/data", true),
		Entry("http://mnt/fluid/data", "http://mnt/fluid/data", false),
		Entry("https://mnt/fluid/data", "https://mnt/fluid/data", false),
	)
})

var _ = Describe("IsFluidWebScheme", func() {
	DescribeTable("should return correct bool for endpoint",
		func(endpoint string, want bool) {
			Expect(IsFluidWebScheme(endpoint)).To(Equal(want))
		},
		Entry("pvc://mnt/fluid/data", "pvc://mnt/fluid/data", false),
		Entry("local://mnt/fluid/data", "local://mnt/fluid/data", false),
		Entry("http://mnt/fluid/data", "http://mnt/fluid/data", true),
		Entry("https://mnt/fluid/data", "https://mnt/fluid/data", true),
	)
})

var _ = Describe("IsFluidRefSchema", func() {
	DescribeTable("should return correct bool for endpoint",
		func(endpoint string, want bool) {
			Expect(IsFluidRefSchema(endpoint)).To(Equal(want))
		},
		Entry("dataset://mnt/fluid/data", "dataset://mnt/fluid/data", true),
		Entry("local://mnt/fluid/data", "local://mnt/fluid/data", false),
	)
})
