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
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("IsFluidNativeScheme", func() {
	ginkgo.DescribeTable("should correctly identify fluid native schemes",
		func(endpoint string, expected bool) {
			gomega.Expect(IsFluidNativeScheme(endpoint)).To(gomega.Equal(expected))
		},
		ginkgo.Entry("pvc scheme is native", "pvc://mnt/fluid/data", true),
		ginkgo.Entry("local scheme is native", "local://mnt/fluid/data", true),
		ginkgo.Entry("http scheme is not native", "http://mnt/fluid/data", false),
		ginkgo.Entry("https scheme is not native", "https://mnt/fluid/data", false),
	)
})

var _ = ginkgo.Describe("IsFluidWebScheme", func() {
	ginkgo.DescribeTable("should correctly identify fluid web schemes",
		func(endpoint string, expected bool) {
			gomega.Expect(IsFluidWebScheme(endpoint)).To(gomega.Equal(expected))
		},
		ginkgo.Entry("pvc scheme is not web", "pvc://mnt/fluid/data", false),
		ginkgo.Entry("local scheme is not web", "local://mnt/fluid/data", false),
		ginkgo.Entry("http scheme is web", "http://mnt/fluid/data", true),
		ginkgo.Entry("https scheme is web", "https://mnt/fluid/data", true),
	)
})

var _ = ginkgo.Describe("IsFluidRefSchema", func() {
	ginkgo.DescribeTable("should correctly identify fluid ref schemes",
		func(endpoint string, expected bool) {
			gomega.Expect(IsFluidRefSchema(endpoint)).To(gomega.Equal(expected))
		},
		ginkgo.Entry("dataset scheme is ref", "dataset://mnt/fluid/data", true),
		ginkgo.Entry("local scheme is not ref", "local://mnt/fluid/data", false),
	)
})
