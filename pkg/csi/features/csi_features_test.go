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

package features

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/component-base/featuregate"
)

func TestFeatures(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Features Suite")
}

var _ = Describe("Feature Gates", func() {
	Context("Feature Gate Constants", func() {
		It("should have correct FuseRecovery constant value", func() {
			Expect(string(FuseRecovery)).To(Equal("FuseRecovery"))
		})
	})

	Context("Default Feature Gates Configuration", func() {
		It("should have FuseRecovery in defaultFeatureGates map", func() {
			spec, exists := defaultFeatureGates[FuseRecovery]
			Expect(exists).To(BeTrue(), "FuseRecovery should exist in defaultFeatureGates")
			Expect(spec).NotTo(BeNil())
		})

		It("should have correct default value for FuseRecovery", func() {
			spec := defaultFeatureGates[FuseRecovery]
			Expect(spec.Default).To(BeFalse(), "FuseRecovery should be disabled by default")
		})

		It("should have correct pre-release stage for FuseRecovery", func() {
			spec := defaultFeatureGates[FuseRecovery]
			Expect(spec.PreRelease).To(Equal(featuregate.Beta), "FuseRecovery should be in Beta stage")
		})
	})

	Context("Default Feature Gates Map Integrity", func() {
		It("should not be empty", func() {
			Expect(defaultFeatureGates).NotTo(BeEmpty(), "defaultFeatureGates should contain at least one feature")
		})

		It("should contain all expected features", func() {
			expectedFeatures := []featuregate.Feature{FuseRecovery}
			for _, feature := range expectedFeatures {
				_, exists := defaultFeatureGates[feature]
				Expect(exists).To(BeTrue(), "feature %q should be present in defaultFeatureGates", feature)
			}
		})

		It("should have exactly one feature defined", func() {
			expectedCount := 1 // Currently only FuseRecovery
			Expect(len(defaultFeatureGates)).To(Equal(expectedCount), "should have exactly %d feature(s)", expectedCount)
		})
	})

	Context("Feature Gate Spec Validation", func() {
		It("should have valid PreRelease stage", func() {
			spec := defaultFeatureGates[FuseRecovery]
			validStages := []interface{}{
				featuregate.Alpha,
				featuregate.Beta,
				featuregate.GA,
				featuregate.Deprecated,
			}
			Expect(validStages).To(ContainElement(spec.PreRelease), "PreRelease stage should be valid")
		})

		It("should have a boolean Default value", func() {
			spec := defaultFeatureGates[FuseRecovery]
			Expect(spec.Default).To(BeAssignableToTypeOf(bool(false)), "Default should be a boolean")
		})
	})

	Context("Initialization", func() {
		It("should have populated defaultFeatureGates after init", func() {
			Expect(len(defaultFeatureGates)).To(BeNumerically(">", 0), "init() should have populated defaultFeatureGates")
		})

		It("should have registered FuseRecovery after init", func() {
			_, exists := defaultFeatureGates[FuseRecovery]
			Expect(exists).To(BeTrue(), "FuseRecovery should be registered after init()")
		})
	})

	Context("Feature Gate Properties", func() {
		var spec featuregate.FeatureSpec

		BeforeEach(func() {
			spec = defaultFeatureGates[FuseRecovery]
		})

		It("should be disabled by default", func() {
			Expect(spec.Default).To(BeFalse())
		})

		It("should be in Beta pre-release stage", func() {
			Expect(spec.PreRelease).To(Equal(featuregate.Beta))
		})

		It("should have a non-nil FeatureSpec", func() {
			Expect(spec).NotTo(BeZero())
		})
	})

	Context("Feature Gate Map Structure", func() {
		It("should map Feature to FeatureSpec", func() {
			for feature, spec := range defaultFeatureGates {
				Expect(feature).To(BeAssignableToTypeOf(featuregate.Feature("")))
				Expect(spec).To(BeAssignableToTypeOf(featuregate.FeatureSpec{}))
			}
		})

		It("should have FuseRecovery as a key", func() {
			keys := make([]featuregate.Feature, 0, len(defaultFeatureGates))
			for key := range defaultFeatureGates {
				keys = append(keys, key)
			}
			Expect(keys).To(ContainElement(FuseRecovery))
		})
	})
})
