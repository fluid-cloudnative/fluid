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
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestErrors(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Errors Suite")
}

func resource(resource string) schema.GroupResource {
	return schema.GroupResource{Group: "", Resource: resource}
}

var _ = Describe("Errors", func() {
	Describe("resource helper function", func() {
		It("should create GroupResource with empty group", func() {
			gr := resource("pods")

			Expect(gr.Group).To(Equal(""))
			Expect(gr.Resource).To(Equal("pods"))
		})

		It("should handle different resource names", func() {
			testCases := []string{
				"DataBackup",
				"datasets",
				"alluxioruntimes",
			}

			for _, resourceName := range testCases {
				gr := resource(resourceName)
				Expect(gr.Group).To(BeEmpty())
				Expect(gr.Resource).To(Equal(resourceName))
			}
		})
	})
})
