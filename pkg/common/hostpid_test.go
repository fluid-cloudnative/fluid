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
	"k8s.io/component-base/featuregate"

	"github.com/fluid-cloudnative/fluid/pkg/common/features"
	utilfeature "github.com/fluid-cloudnative/fluid/pkg/utils/feature"
)

func setFeatureGate(feature featuregate.Feature, enabled bool) {
	original := utilfeature.DefaultFeatureGate.Enabled(feature)
	err := utilfeature.DefaultMutableFeatureGate.Set(string(feature) + "=" + boolToString(enabled))
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	DeferCleanup(func() {
		err := utilfeature.DefaultMutableFeatureGate.Set(string(feature) + "=" + boolToString(original))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

var _ = Describe("HostPIDEnabled", func() {
	Context("when RuntimeFuseHostPID feature gate is disabled (default)", func() {
		BeforeEach(func() {
			setFeatureGate(features.RuntimeFuseHostPID, false)
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
			setFeatureGate(features.RuntimeFuseHostPID, true)
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
			gomega.Expect(utilfeature.DefaultFeatureGate.Enabled(features.RuntimeFuseHostPID)).To(gomega.BeFalse(),
				"Security default: RuntimeFuseHostPID must be disabled by default")
		})
	})
})
