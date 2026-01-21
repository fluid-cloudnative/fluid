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

package errors

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("NotSupported", func() {
	Describe("NewNotSupported", func() {
		Context("with valid inputs", func() {
			It("should create a not supported error with correct fields", func() {
				qualifiedResource := schema.GroupResource{
					Group:    "data.fluid.io",
					Resource: "datasets",
				}
				targetType := "AlluxioRuntime"

				err := NewNotSupported(qualifiedResource, targetType)

				Expect(err).NotTo(BeNil())
				Expect(err.Reason()).To(Equal(StatusReasonNotSupported))
				Expect(err.Details().Group).To(Equal("data.fluid.io"))
				Expect(err.Details().Kind).To(Equal("datasets"))
				Expect(err.Error()).To(ContainSubstring("datasets"))
				Expect(err.Error()).To(ContainSubstring("is not supported by"))
				Expect(err.Error()).To(ContainSubstring("AlluxioRuntime"))
			})

			It("should handle empty group", func() {
				qualifiedResource := schema.GroupResource{
					Group:    "",
					Resource: "pods",
				}
				targetType := "CustomRuntime"

				err := NewNotSupported(qualifiedResource, targetType)

				Expect(err).NotTo(BeNil())
				Expect(err.Reason()).To(Equal(StatusReasonNotSupported))
				Expect(err.Details().Group).To(Equal(""))
				Expect(err.Details().Kind).To(Equal("pods"))
				Expect(err.Error()).To(ContainSubstring("pods"))
				Expect(err.Error()).To(ContainSubstring("CustomRuntime"))
			})

			It("should handle empty target type", func() {
				qualifiedResource := schema.GroupResource{
					Group:    "apps",
					Resource: "deployments",
				}
				targetType := ""

				err := NewNotSupported(qualifiedResource, targetType)

				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("deployments"))
				Expect(err.Error()).To(ContainSubstring("is not supported by "))
			})

			It("should create valid error for multiple resource types", func() {
				testCases := []struct {
					resource   schema.GroupResource
					targetType string
					groupName  string
					kindName   string
				}{
					{
						resource:   schema.GroupResource{Group: "data.fluid.io", Resource: "alluxioruntimes"},
						targetType: "JindoRuntime",
						groupName:  "data.fluid.io",
						kindName:   "alluxioruntimes",
					},
					{
						resource:   schema.GroupResource{Group: "data.fluid.io", Resource: "jindoruntimes"},
						targetType: "AlluxioRuntime",
						groupName:  "data.fluid.io",
						kindName:   "jindoruntimes",
					},
					{
						resource:   schema.GroupResource{Group: "storage.k8s.io", Resource: "storageclasses"},
						targetType: "FluidRuntime",
						groupName:  "storage.k8s.io",
						kindName:   "storageclasses",
					},
				}

				for _, tc := range testCases {
					err := NewNotSupported(tc.resource, tc.targetType)
					Expect(err).NotTo(BeNil())
					Expect(err.Details().Group).To(Equal(tc.groupName))
					Expect(err.Details().Kind).To(Equal(tc.kindName))
					Expect(err.Reason()).To(Equal(StatusReasonNotSupported))
					Expect(err.Error()).To(ContainSubstring(tc.kindName))
					Expect(err.Error()).To(ContainSubstring(tc.targetType))
				}
			})
		})
	})

	Describe("IsNotSupported", func() {
		Context("with not supported error", func() {
			It("should return true for error created by NewNotSupported", func() {
				qualifiedResource := schema.GroupResource{
					Group:    "data.fluid.io",
					Resource: "datasets",
				}
				err := NewNotSupported(qualifiedResource, "TestRuntime")

				Expect(IsNotSupported(err)).To(BeTrue())
			})
		})

		Context("with other error types", func() {
			It("should return false for standard errors", func() {
				err := errors.New("standard error")

				Expect(IsNotSupported(err)).To(BeFalse())
			})

			It("should return false for nil error", func() {
				Expect(IsNotSupported(nil)).To(BeFalse())
			})

			It("should return false for FluidStatusError with different reason", func() {
				err := &FluidStatusError{
					reason:  StatusReasonDeprecated,
					details: &metav1.StatusDetails{},
					message: "test error",
				}

				Expect(IsNotSupported(err)).To(BeFalse())
			})

			It("should return false for temporary validation failed error", func() {
				err := NewTemporaryValidationFailed("test failure")

				Expect(IsNotSupported(err)).To(BeFalse())
			})
		})
	})

	Describe("NewTemporaryValidationFailed", func() {
		Context("with valid inputs", func() {
			It("should create a temporary validation failed error with correct fields", func() {
				failureMsg := "runtime not ready"

				err := NewTemporaryValidationFailed(failureMsg)

				Expect(err).NotTo(BeNil())
				Expect(err.Reason()).To(Equal(StatusReasonTemporaryValidationFailed))
				Expect(err.Details()).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("temporary validation failure"))
				Expect(err.Error()).To(ContainSubstring("runtime not ready"))
			})

			It("should handle empty failure message", func() {
				err := NewTemporaryValidationFailed("")

				Expect(err).NotTo(BeNil())
				Expect(err.Reason()).To(Equal(StatusReasonTemporaryValidationFailed))
				Expect(err.Error()).To(ContainSubstring("temporary validation failure"))
			})

			It("should create valid error for different failure messages", func() {
				testCases := []string{
					"resource quota exceeded",
					"node not schedulable",
					"volume mount failed",
					"network policy conflict",
				}

				for _, failureMsg := range testCases {
					err := NewTemporaryValidationFailed(failureMsg)
					Expect(err).NotTo(BeNil())
					Expect(err.Reason()).To(Equal(StatusReasonTemporaryValidationFailed))
					Expect(err.Error()).To(ContainSubstring(failureMsg))
					Expect(err.Error()).To(ContainSubstring("temporary validation failure"))
				}
			})
		})
	})

	Describe("IsTemporaryValidationFailed", func() {
		Context("with temporary validation failed error", func() {
			It("should return true for error created by NewTemporaryValidationFailed", func() {
				err := NewTemporaryValidationFailed("test failure")

				Expect(IsTemporaryValidationFailed(err)).To(BeTrue())
			})
		})

		Context("with other error types", func() {
			It("should return false for standard errors", func() {
				err := errors.New("standard error")

				Expect(IsTemporaryValidationFailed(err)).To(BeFalse())
			})

			It("should return false for nil error", func() {
				Expect(IsTemporaryValidationFailed(nil)).To(BeFalse())
			})

			It("should return false for FluidStatusError with different reason", func() {
				err := &FluidStatusError{
					reason:  StatusReasonNotSupported,
					details: &metav1.StatusDetails{},
					message: "test error",
				}

				Expect(IsTemporaryValidationFailed(err)).To(BeFalse())
			})

			It("should return false for not supported error", func() {
				qualifiedResource := schema.GroupResource{
					Group:    "data.fluid.io",
					Resource: "datasets",
				}
				err := NewNotSupported(qualifiedResource, "TestRuntime")

				Expect(IsTemporaryValidationFailed(err)).To(BeFalse())
			})
		})
	})

	// Optional: Add edge case tests
	Describe("Edge Cases", func() {
		It("should handle special characters in resource names", func() {
			qualifiedResource := schema.GroupResource{
				Group:    "data.fluid.io",
				Resource: "special-resource_v1",
			}
			err := NewNotSupported(qualifiedResource, "Runtime@v2")

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("special-resource_v1"))
			Expect(err.Error()).To(ContainSubstring("Runtime@v2"))
		})

		It("should handle very long failure messages", func() {
			longMsg := "This is a very long failure message that exceeds normal length and should still be properly handled by the error creation function without any truncation or issues"
			err := NewTemporaryValidationFailed(longMsg)

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring(longMsg))
		})
	})
})
