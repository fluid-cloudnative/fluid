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

var _ = Describe("validateRuntimeClassTopology", func() {
	It("rejects a runtime class without a topology", func() {
		err := validateRuntimeClassTopology(&datav1alpha1.CacheRuntimeClass{})

		Expect(err).To(MatchError("at least one component should be defined in runtimeClass "))
	})

	It("rejects a topology without a master, worker, or client", func() {
		err := validateRuntimeClassTopology(&datav1alpha1.CacheRuntimeClass{
			ObjectMeta: metav1.ObjectMeta{Name: "empty-topology"},
			Topology:   &datav1alpha1.RuntimeTopology{},
		})

		Expect(err).To(MatchError("at least one component should be defined in runtimeClass empty-topology"))
	})

	DescribeTable("accepts a topology with at least one component",
		func(topology *datav1alpha1.RuntimeTopology) {
			err := validateRuntimeClassTopology(&datav1alpha1.CacheRuntimeClass{Topology: topology})
			Expect(err).NotTo(HaveOccurred())
		},
		Entry("master", &datav1alpha1.RuntimeTopology{Master: &datav1alpha1.RuntimeComponentDefinition{}}),
		Entry("worker", &datav1alpha1.RuntimeTopology{Worker: &datav1alpha1.RuntimeComponentDefinition{}}),
		Entry("client", &datav1alpha1.RuntimeTopology{Client: &datav1alpha1.RuntimeComponentDefinition{}}),
	)
})
