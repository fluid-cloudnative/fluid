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

package engine

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/validation"
)

var _ = Describe("getSecretVolumeName Tests", Label("pkg.ddc.cache.engine.util_test.go"), func() {
	Describe("getSecretVolumeName", func() {
		Context("when secret name is short", func() {
			It("should return prefix + secretName without truncation", func() {
				secretName := "test-secret"
				result := getSecretVolumeName(secretName)
				expected := secretVolumeNamePrefix + secretName

				Expect(result).To(Equal(expected))
				Expect(len(result)).To(BeNumerically("<=", validation.DNS1035LabelMaxLength))
			})

			It("should handle empty secret name", func() {
				secretName := ""
				result := getSecretVolumeName(secretName)
				expected := secretVolumeNamePrefix

				Expect(result).To(Equal(expected))
				Expect(len(result)).To(BeNumerically("<=", validation.DNS1035LabelMaxLength))
			})

			It("should handle single character secret name", func() {
				secretName := "a"
				result := getSecretVolumeName(secretName)
				expected := secretVolumeNamePrefix + "a"

				Expect(result).To(Equal(expected))
				Expect(len(result)).To(BeNumerically("<=", validation.DNS1035LabelMaxLength))
			})
		})

		Context("when secret name is at the boundary", func() {
			It("should not truncate when total length equals DNS1035LabelMaxLength", func() {
				// Calculate max secret name length that fits exactly
				maxSecretLen := validation.DNS1035LabelMaxLength - prefixSecretVolumeLength
				secretName := strings.Repeat("a", maxSecretLen)
				result := getSecretVolumeName(secretName)
				expected := secretVolumeNamePrefix + secretName

				Expect(result).To(Equal(expected))
				Expect(len(result)).To(Equal(validation.DNS1035LabelMaxLength))
			})

			It("should truncate when total length exceeds DNS1035LabelMaxLength by 1", func() {
				// Calculate secret name length that exceeds by 1
				maxSecretLen := validation.DNS1035LabelMaxLength - prefixSecretVolumeLength + 1
				secretName := strings.Repeat("a", maxSecretLen)
				result := getSecretVolumeName(secretName)

				Expect(len(result)).To(BeNumerically("<=", validation.DNS1035LabelMaxLength))
				Expect(result).To(HavePrefix(secretVolumeNamePrefix))
				// Should have hash suffix
				Expect(len(result)).To(Equal(validation.DNS1035LabelMaxLength))
			})
		})

		Context("when secret name is very long", func() {
			It("should truncate and add hash suffix", func() {
				secretName := strings.Repeat("very-long-secret-name-", 10) // 240 chars
				result := getSecretVolumeName(secretName)

				Expect(len(result)).To(BeNumerically("<=", validation.DNS1035LabelMaxLength))
				Expect(result).To(HavePrefix(secretVolumeNamePrefix))
				// Should end with 8-char hash
				Expect(len(result)).To(Equal(validation.DNS1035LabelMaxLength))

				// Verify hash part is present (last 8 chars)
				hashPart := result[len(result)-hashSuffixLength:]
				Expect(hashPart).To(MatchRegexp("^[0-9a-f]{8}$"))
			})

			It("should generate different hashes for different secret names with same prefix", func() {
				secretName1 := strings.Repeat("a", 100) + "suffix1"
				secretName2 := strings.Repeat("a", 100) + "suffix2"

				result1 := getSecretVolumeName(secretName1)
				result2 := getSecretVolumeName(secretName2)

				Expect(result1).NotTo(Equal(result2))
				// Both should be valid DNS names
				Expect(len(result1)).To(BeNumerically("<=", validation.DNS1035LabelMaxLength))
				Expect(len(result2)).To(BeNumerically("<=", validation.DNS1035LabelMaxLength))
			})

			It("should handle extremely long secret names", func() {
				secretName := strings.Repeat("x", 1000)
				result := getSecretVolumeName(secretName)

				Expect(len(result)).To(BeNumerically("<=", validation.DNS1035LabelMaxLength))
				Expect(result).To(HavePrefix(secretVolumeNamePrefix))
				Expect(len(result)).To(Equal(validation.DNS1035LabelMaxLength))
			})
		})

		Context("when secret name contains special characters", func() {
			It("should handle secret names with hyphens", func() {
				secretName := "my-test-secret-name"
				result := getSecretVolumeName(secretName)
				expected := secretVolumeNamePrefix + secretName

				Expect(result).To(Equal(expected))
			})

			It("should handle secret names with numbers", func() {
				secretName := "secret123"
				result := getSecretVolumeName(secretName)
				expected := secretVolumeNamePrefix + secretName

				Expect(result).To(Equal(expected))
			})

			It("should handle long secret names with special characters", func() {
				secretName := strings.Repeat("test-secret-123-", 10)
				result := getSecretVolumeName(secretName)

				Expect(len(result)).To(BeNumerically("<=", validation.DNS1035LabelMaxLength))
				Expect(result).To(HavePrefix(secretVolumeNamePrefix))
			})
		})

		Context("DNS label validation", func() {
			It("should always generate valid DNS-1035 label names", func() {
				testCases := []string{
					"short",
					"medium-length-secret-name",
					strings.Repeat("a", 50),
					strings.Repeat("b", 100),
					strings.Repeat("c", 200),
					"test-with-hyphens-and-numbers-123",
				}

				for _, secretName := range testCases {
					result := getSecretVolumeName(secretName)

					// Check length constraint
					Expect(len(result)).To(BeNumerically("<=", validation.DNS1035LabelMaxLength),
						"Result length %d exceeds max %d for secret: %s",
						len(result), validation.DNS1035LabelMaxLength, secretName)

					// Check starts with letter
					Expect(result[0]).To(Satisfy(func(c byte) bool {
						return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
					}), "Result should start with a letter for secret: %s", secretName)

					// Check only contains lowercase letters, numbers, and hyphens
					Expect(result).To(MatchRegexp("^[a-z0-9-]+$"),
						"Result contains invalid characters for secret: %s", secretName)
				}
			})
		})

		Context("collision prevention", func() {
			It("should generate unique volume names for different secrets that truncate to same prefix", func() {
				// Two different secrets that would truncate to the same prefix
				secretName1 := strings.Repeat("a", truncatedSecretMaxLength) + "different1"
				secretName2 := strings.Repeat("a", truncatedSecretMaxLength) + "different2"

				result1 := getSecretVolumeName(secretName1)
				result2 := getSecretVolumeName(secretName2)

				// Should be different due to hash
				Expect(result1).NotTo(Equal(result2))

				// Both should be valid length
				Expect(len(result1)).To(Equal(validation.DNS1035LabelMaxLength))
				Expect(len(result2)).To(Equal(validation.DNS1035LabelMaxLength))
			})
		})
	})
})
