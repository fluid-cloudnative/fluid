/*
Copyright 2026 The Fluid Authors.

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

package base_test

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const testNamespace = "default"

var _ = Describe("RuntimeInfo.GetWorkerStatefulsetName", func() {
	DescribeTable("returns correct statefulset name",
		func(runtimeName, runtimeType, suffix string) {
			info, err := base.BuildRuntimeInfo(runtimeName, testNamespace, runtimeType)
			Expect(err).NotTo(HaveOccurred())
			Expect(info.GetWorkerStatefulsetName()).To(Equal(runtimeName + suffix))
		},
		Entry("JindoRuntime uses jindofs suffix", "mydata", common.JindoRuntime, "-jindofs-worker"),
		Entry("JindoCacheEngineImpl uses jindofs suffix", "cache", common.JindoCacheEngineImpl, "-jindofs-worker"),
		Entry("JindoFSxEngineImpl uses jindofs suffix", "fsx", common.JindoFSxEngineImpl, "-jindofs-worker"),
		Entry("AlluxioRuntime uses default suffix", "alluxio-data", common.AlluxioRuntime, "-worker"),
		Entry("JuiceFSRuntime uses default suffix", "juice", common.JuiceFSRuntime, "-worker"),
		Entry("empty runtime type uses default suffix", "test", "", "-worker"),
		Entry("unknown runtime type uses default suffix", "unknown-data", "UnknownRuntime", "-worker"),
	)
})
