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

package errors

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Deprecated", func() {
	Describe("FluidStatusError", func() {
		Context("Error method", func() {
			It("should return the error message", func() {
				err := FluidStatusError{
					message: "test error message",
				}
				Expect(err.Error()).To(Equal("test error message"))
			})

			It("should return empty string when message is empty", func() {
				err := FluidStatusError{
					message: "",
				}
				Expect(err.Error()).To(Equal(""))
			})
		})

		Context("Reason method", func() {
			It("should return the status reason", func() {
				err := FluidStatusError{
					reason: StatusReasonDeprecated,
				}
				Expect(err.Reason()).To(Equal(StatusReasonDeprecated))
			})

			It("should return custom status reason", func() {
				customReason := metav1.StatusReason("CustomReason")
				err := FluidStatusError{
					reason: customReason,
				}
				Expect(err.Reason()).To(Equal(customReason))
			})
		})

		Context("Details method", func() {
			It("should return the status details", func() {
				details := &metav1.StatusDetails{
					Group: "data.fluid.io",
					Kind:  "datasets",
				}
				err := FluidStatusError{
					details: details,
				}
				Expect(err.Details()).To(Equal(details))
			})

			It("should return nil when details is nil", func() {
				err := FluidStatusError{
					details: nil,
				}
				Expect(err.Details()).To(BeNil())
			})
		})
	})

	Describe("NewDeprecated", func() {
		Context("with valid inputs", func() {

			It("should handle empty group resource", func() {
				qualifiedResource := schema.GroupResource{
					Group:    "",
					Resource: "pods",
				}
				key := types.NamespacedName{
					Namespace: "kube-system",
					Name:      "test-pod",
				}

				err := NewDeprecated(qualifiedResource, key)

				Expect(err).NotTo(BeNil())
				Expect(err.Reason()).To(Equal(StatusReasonDeprecated))
				Expect(err.Details().Group).To(Equal(""))
				Expect(err.Details().Kind).To(Equal("pods"))
			})

			It("should handle empty namespace", func() {
				qualifiedResource := schema.GroupResource{
					Group:    "apps",
					Resource: "deployments",
				}
				key := types.NamespacedName{
					Namespace: "",
					Name:      "my-deployment",
				}

				err := NewDeprecated(qualifiedResource, key)

				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("namespace "))
				Expect(err.Error()).To(ContainSubstring("my-deployment"))
			})

			It("should create valid error for multiple resources", func() {
				testCases := []struct {
					resource  schema.GroupResource
					key       types.NamespacedName
					groupName string
					kindName  string
				}{
					{
						resource:  schema.GroupResource{Group: "data.fluid.io", Resource: "alluxioruntimes"},
						key:       types.NamespacedName{Namespace: "fluid", Name: "runtime-1"},
						groupName: "data.fluid.io",
						kindName:  "alluxioruntimes",
					},
					{
						resource:  schema.GroupResource{Group: "data.fluid.io", Resource: "jindoruntimes"},
						key:       types.NamespacedName{Namespace: "test", Name: "runtime-2"},
						groupName: "data.fluid.io",
						kindName:  "jindoruntimes",
					},
				}

				for _, tc := range testCases {
					err := NewDeprecated(tc.resource, tc.key)
					Expect(err).NotTo(BeNil())
					Expect(err.Details().Group).To(Equal(tc.groupName))
					Expect(err.Details().Kind).To(Equal(tc.kindName))
					Expect(err.Reason()).To(Equal(StatusReasonDeprecated))
				}
			})
		})
	})
})
