/*
Copyright 2025 The Fluid Authors.

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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/component-base/featuregate"
)

var _ = Describe("Runtime Feature Gates", func() {
	Context("Feature Gate Constants", func() {
		It("should have correct RuntimeFuseHostPID constant value", func() {
			Expect(string(RuntimeFuseHostPID)).To(Equal("RuntimeFuseHostPID"))
		})
	})

	Context("Default Feature Gates Configuration", func() {
		It("should have RuntimeFuseHostPID in defaultFeatureGates map", func() {
			spec, exists := defaultFeatureGates[RuntimeFuseHostPID]
			Expect(exists).To(BeTrue(), "RuntimeFuseHostPID should exist in defaultFeatureGates")
			Expect(spec).NotTo(BeNil())
		})

		It("should have correct default value for RuntimeFuseHostPID", func() {
			spec := defaultFeatureGates[RuntimeFuseHostPID]
			Expect(spec.Default).To(BeFalse(), "RuntimeFuseHostPID should be disabled by default for security")
		})

		It("should have correct pre-release stage for RuntimeFuseHostPID", func() {
			spec := defaultFeatureGates[RuntimeFuseHostPID]
			Expect(spec.PreRelease).To(Equal(featuregate.Alpha), "RuntimeFuseHostPID should be in Alpha stage")
		})
	})

	Context("Default Feature Gates Map Integrity", func() {
		It("should not be empty", func() {
			Expect(defaultFeatureGates).NotTo(BeEmpty(), "defaultFeatureGates should contain at least one feature")
		})

		It("should contain RuntimeFuseHostPID", func() {
			_, exists := defaultFeatureGates[RuntimeFuseHostPID]
			Expect(exists).To(BeTrue(), "RuntimeFuseHostPID should be present in defaultFeatureGates")
		})

		It("should have exactly one feature defined", func() {
			expectedCount := 1
			Expect(len(defaultFeatureGates)).To(Equal(expectedCount), "should have exactly %d feature(s)", expectedCount)
		})
	})

	Context("Feature Gate Spec Validation", func() {
		It("should have valid PreRelease stage", func() {
			spec := defaultFeatureGates[RuntimeFuseHostPID]
			validStages := []interface{}{
				featuregate.Alpha,
				featuregate.Beta,
				featuregate.GA,
				featuregate.Deprecated,
			}
			Expect(validStages).To(ContainElement(spec.PreRelease), "PreRelease stage should be valid")
		})

		It("should have a boolean Default value", func() {
			spec := defaultFeatureGates[RuntimeFuseHostPID]
			Expect(spec.Default).To(BeAssignableToTypeOf(false), "Default should be a boolean")
		})
	})

	Context("Initialization", func() {
		It("should have populated defaultFeatureGates after init", func() {
			Expect(len(defaultFeatureGates)).To(BeNumerically(">", 0), "init() should have populated defaultFeatureGates")
		})

		It("should have registered RuntimeFuseHostPID after init", func() {
			_, exists := defaultFeatureGates[RuntimeFuseHostPID]
			Expect(exists).To(BeTrue(), "RuntimeFuseHostPID should be registered after init()")
		})
	})

	Context("Feature Gate Properties", func() {
		var spec featuregate.FeatureSpec

		BeforeEach(func() {
			spec = defaultFeatureGates[RuntimeFuseHostPID]
		})

		It("should be disabled by default for security", func() {
			Expect(spec.Default).To(BeFalse(), "RuntimeFuseHostPID must be disabled by default for security")
		})

		It("should be in Alpha pre-release stage", func() {
			Expect(spec.PreRelease).To(Equal(featuregate.Alpha))
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

		It("should have RuntimeFuseHostPID as a key", func() {
			keys := make([]featuregate.Feature, 0, len(defaultFeatureGates))
			for key := range defaultFeatureGates {
				keys = append(keys, key)
			}
			Expect(keys).To(ContainElement(RuntimeFuseHostPID))
		})
	})
})
