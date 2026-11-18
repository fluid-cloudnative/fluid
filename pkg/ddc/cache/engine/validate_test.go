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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("validateRuntimeClassTopology", Label("pkg.ddc.cache.engine.validate_test.go"), func() {
	newClass := func(topology *datav1alpha1.RuntimeTopology) *datav1alpha1.CacheRuntimeClass {
		return &datav1alpha1.CacheRuntimeClass{
			ObjectMeta: metav1.ObjectMeta{Name: "test-class"},
			Topology:   topology,
		}
	}

	It("rejects a class without a topology", func() {
		Expect(validateRuntimeClassTopology(newClass(nil))).To(MatchError(ContainSubstring("test-class")))
	})

	It("rejects a topology with no components", func() {
		Expect(validateRuntimeClassTopology(newClass(&datav1alpha1.RuntimeTopology{}))).To(MatchError(ContainSubstring("at least one component")))
	})

	DescribeTable("accepts a topology with at least one component",
		func(topology *datav1alpha1.RuntimeTopology) {
			Expect(validateRuntimeClassTopology(newClass(topology))).To(Succeed())
		},
		Entry("master only", &datav1alpha1.RuntimeTopology{Master: &datav1alpha1.RuntimeComponentDefinition{}}),
		Entry("worker only", &datav1alpha1.RuntimeTopology{Worker: &datav1alpha1.RuntimeComponentDefinition{}}),
		Entry("client only", &datav1alpha1.RuntimeTopology{Client: &datav1alpha1.RuntimeComponentDefinition{}}),
	)
})
