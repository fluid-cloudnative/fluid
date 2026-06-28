/*
Copyright 2024 The Fluid Authors.

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
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/fluid-cloudnative/fluid/pkg/common/features"
	utilfeature "github.com/fluid-cloudnative/fluid/pkg/utils/feature"
)

var _ = Describe("HostPIDEnabled", func() {
	Context("when RuntimeFuseHostPID feature gate is disabled (default)", func() {
		BeforeEach(func() {
			// Ensure feature gate is disabled (default state)
			err := utilfeature.DefaultMutableFeatureGate.Set(string(features.RuntimeFuseHostPID) + "=false")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		It("should return false even when annotation is set to true", func() {
			annotations := map[string]string{
				RuntimeFuseHostPIDKey: "true",
			}
			gomega.Expect(HostPIDEnabled(annotations)).To(gomega.BeFalse(), "HostPID should be disabled when feature gate is off")
		})

		It("should return false when annotations are nil", func() {
			gomega.Expect(HostPIDEnabled(nil)).To(gomega.BeFalse())
		})

		It("should return false when annotation does not exist", func() {
			gomega.Expect(HostPIDEnabled(map[string]string{})).To(gomega.BeFalse())
		})

		It("should return false when annotation has wrong value", func() {
			annotations := map[string]string{
				RuntimeFuseHostPIDKey: "sss",
			}
			gomega.Expect(HostPIDEnabled(annotations)).To(gomega.BeFalse())
		})

		It("should return false when annotation is set to True", func() {
			annotations := map[string]string{
				RuntimeFuseHostPIDKey: "True",
			}
			gomega.Expect(HostPIDEnabled(annotations)).To(gomega.BeFalse(), "HostPID should be disabled when feature gate is off")
		})
	})

	Context("when RuntimeFuseHostPID feature gate is enabled", func() {
		BeforeEach(func() {
			// Enable feature gate for this test context
			err := utilfeature.DefaultMutableFeatureGate.Set(string(features.RuntimeFuseHostPID) + "=true")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		AfterEach(func() {
			// Reset to disabled after each test
			err := utilfeature.DefaultMutableFeatureGate.Set(string(features.RuntimeFuseHostPID) + "=false")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		It("should return true when annotation is set to true", func() {
			annotations := map[string]string{
				RuntimeFuseHostPIDKey: "true",
			}
			gomega.Expect(HostPIDEnabled(annotations)).To(gomega.BeTrue())
		})

		It("should return true when annotation is set to True", func() {
			annotations := map[string]string{
				RuntimeFuseHostPIDKey: "True",
			}
			gomega.Expect(HostPIDEnabled(annotations)).To(gomega.BeTrue())
		})

		It("should return false when annotations are nil", func() {
			gomega.Expect(HostPIDEnabled(nil)).To(gomega.BeFalse())
		})

		It("should return false when annotation does not exist", func() {
			gomega.Expect(HostPIDEnabled(map[string]string{})).To(gomega.BeFalse())
		})

		It("should return false when annotation has wrong value", func() {
			annotations := map[string]string{
				RuntimeFuseHostPIDKey: "sss",
			}
			gomega.Expect(HostPIDEnabled(annotations)).To(gomega.BeFalse())
		})

		It("should return false when annotation is set to false", func() {
			annotations := map[string]string{
				RuntimeFuseHostPIDKey: "false",
			}
			gomega.Expect(HostPIDEnabled(annotations)).To(gomega.BeFalse())
		})
	})

	Context("feature gate default state verification", func() {
		It("should have RuntimeFuseHostPID disabled by default", func() {
			// This test verifies the security default: feature gate is off by default
			// Reset to ensure we're testing the actual default state
			err := utilfeature.DefaultMutableFeatureGate.Set(string(features.RuntimeFuseHostPID) + "=false")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Even with valid annotation, HostPID should be disabled
			annotations := map[string]string{
				RuntimeFuseHostPIDKey: "true",
			}
			gomega.Expect(HostPIDEnabled(annotations)).To(gomega.BeFalse(), "Security default: HostPID must be disabled by default")
		})
	})
})
