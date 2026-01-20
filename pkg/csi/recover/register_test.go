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

package recover

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/csi/features"
	utilfeature "github.com/fluid-cloudnative/fluid/pkg/utils/feature"
)

var _ = Describe("Enabled", func() {
	var originalEnv string

	AfterEach(func() {
		if originalEnv == "" {
			os.Unsetenv("NODEPUBLISH_METHOD")
		} else {
			os.Setenv("NODEPUBLISH_METHOD", originalEnv)
		}
	})

	Context("when NODEPUBLISH_METHOD is symlink", func() {
		BeforeEach(func() {
			os.Setenv("NODEPUBLISH_METHOD", common.NodePublishMethodSymlink)
		})

		It("should return false regardless of feature gate", func() {
			Expect(Enabled()).To(BeFalse())
		})

		It("should return false when FuseRecovery feature is enabled", func() {
			err := utilfeature.DefaultMutableFeatureGate.Set(string(features.FuseRecovery) + "=true")
			Expect(err).NotTo(HaveOccurred())
			Expect(Enabled()).To(BeFalse())
		})
	})

	Context("when NODEPUBLISH_METHOD is bind", func() {
		BeforeEach(func() {
			os.Setenv("NODEPUBLISH_METHOD", "bind")
		})

		It("should return true when FuseRecovery feature is enabled", func() {
			err := utilfeature.DefaultMutableFeatureGate.Set(string(features.FuseRecovery) + "=true")
			Expect(err).NotTo(HaveOccurred())
			Expect(Enabled()).To(BeTrue())
		})

		It("should return false when FuseRecovery feature is disabled", func() {
			err := utilfeature.DefaultMutableFeatureGate.Set(string(features.FuseRecovery) + "=false")
			Expect(err).NotTo(HaveOccurred())
			Expect(Enabled()).To(BeFalse())
		})
	})

	Context("when NODEPUBLISH_METHOD is not set", func() {
		BeforeEach(func() {
			os.Unsetenv("NODEPUBLISH_METHOD")
		})

		It("should return true when FuseRecovery feature is enabled", func() {
			err := utilfeature.DefaultMutableFeatureGate.Set(string(features.FuseRecovery) + "=true")
			Expect(err).NotTo(HaveOccurred())
			Expect(Enabled()).To(BeTrue())
		})

		It("should return false when FuseRecovery feature is disabled", func() {
			err := utilfeature.DefaultMutableFeatureGate.Set(string(features.FuseRecovery) + "=false")
			Expect(err).NotTo(HaveOccurred())
			Expect(Enabled()).To(BeFalse())
		})
	})

	Context("when NODEPUBLISH_METHOD is empty string", func() {
		BeforeEach(func() {
			os.Setenv("NODEPUBLISH_METHOD", "")
		})

		It("should return true when FuseRecovery feature is enabled", func() {
			err := utilfeature.DefaultMutableFeatureGate.Set(string(features.FuseRecovery) + "=true")
			Expect(err).NotTo(HaveOccurred())
			Expect(Enabled()).To(BeTrue())
		})

		It("should return false when FuseRecovery feature is disabled", func() {
			err := utilfeature.DefaultMutableFeatureGate.Set(string(features.FuseRecovery) + "=false")
			Expect(err).NotTo(HaveOccurred())
			Expect(Enabled()).To(BeFalse())
		})
	})

	Context("when NODEPUBLISH_METHOD has other values", func() {
		BeforeEach(func() {
			os.Setenv("NODEPUBLISH_METHOD", "mount")
		})

		It("should return true when FuseRecovery feature is enabled", func() {
			err := utilfeature.DefaultMutableFeatureGate.Set(string(features.FuseRecovery) + "=true")
			Expect(err).NotTo(HaveOccurred())
			Expect(Enabled()).To(BeTrue())
		})

		It("should return false when FuseRecovery feature is disabled", func() {
			err := utilfeature.DefaultMutableFeatureGate.Set(string(features.FuseRecovery) + "=false")
			Expect(err).NotTo(HaveOccurred())
			Expect(Enabled()).To(BeFalse())
		})
	})
})
