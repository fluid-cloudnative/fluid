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

package options

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
)

var _ = Describe("CriticalFusePodEnabled", func() {

	Context("when environment variable is not set", func() {
		BeforeEach(func() {
			os.Unsetenv(EnvCriticalFusePodEnabled)
		})

		It("should return false", func() {
			setCriticalFusePodOption()
			got := CriticalFusePodEnabled()
			Expect(got).To(BeFalse())
		})
	})

	Context("when environment variable is set to true", func() {
		BeforeEach(func() {
			os.Setenv(EnvCriticalFusePodEnabled, "true")
		})

		AfterEach(func() {
			os.Unsetenv(EnvCriticalFusePodEnabled)
		})

		It("should return true", func() {
			setCriticalFusePodOption()
			got := CriticalFusePodEnabled()
			Expect(got).To(BeTrue())
		})
	})

	Context("when environment variable is set to false", func() {
		BeforeEach(func() {
			os.Setenv(EnvCriticalFusePodEnabled, "false")
		})

		AfterEach(func() {
			os.Unsetenv(EnvCriticalFusePodEnabled)
		})

		It("should return false", func() {
			setCriticalFusePodOption()
			got := CriticalFusePodEnabled()
			Expect(got).To(BeFalse())
		})
	})
})
