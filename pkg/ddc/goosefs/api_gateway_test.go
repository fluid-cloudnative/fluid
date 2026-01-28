/*
Copyright 2022 The Fluid Author.

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

package goosefs

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func mockGooseFSEngineWithClient(name, ns string, port int32) *GooseFSEngine {
	var mockClient client.Client

	mockSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-master-0", name),
			Namespace: ns,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "rest",
					Port: port,
				},
			},
		},
	}

	mockClient = fake.NewFakeClient(mockSvc)

	e := &GooseFSEngine{
		Client:    mockClient,
		name:      name,
		namespace: ns,
	}
	return e
}

var _ = Describe("APIGateway", func() {
	endpointFormat := "%s-master-0.%s:%d"

	type testCase struct {
		name            string
		engineName      string
		engineNamespace string
		port            int32
		wantEndpoint    string
	}

	Describe("GetAPIGatewayStatus", func() {
		DescribeTable("should return correct API gateway status",
			func(tc testCase) {
				e := mockGooseFSEngineWithClient(tc.engineName, tc.engineNamespace, tc.port)
				got, err := e.GetAPIGatewayStatus()

				Expect(err).NotTo(HaveOccurred())
				expectedStatus := &datav1alpha1.APIGatewayStatus{
					Endpoint: fmt.Sprintf(endpointFormat, tc.engineName, tc.engineNamespace, tc.port),
				}
				Expect(got).To(Equal(expectedStatus))
			},
			Entry("fluid engine in default namespace",
				testCase{
					name:            "case 1",
					engineName:      "fluid",
					engineNamespace: "default",
					port:            8080,
					wantEndpoint:    fmt.Sprintf(endpointFormat, "fluid", "default", 8080),
				},
			),
			Entry("demo engine in fluid-system namespace",
				testCase{
					name:            "case 2",
					engineName:      "demo",
					engineNamespace: common.NamespaceFluidSystem,
					port:            80,
					wantEndpoint:    fmt.Sprintf(endpointFormat, "demo", common.NamespaceFluidSystem, 80),
				},
			),
		)
	})

	Describe("queryAPIGatewayEndpoint", func() {
		DescribeTable("should return correct endpoint",
			func(tc testCase) {
				e := mockGooseFSEngineWithClient(tc.engineName, tc.engineNamespace, tc.port)
				got, err := e.queryAPIGatewayEndpoint()

				Expect(err).NotTo(HaveOccurred())
				Expect(got).To(Equal(tc.wantEndpoint))
			},
			Entry("fluid engine in default namespace",
				testCase{
					name:            "case 1",
					engineName:      "fluid",
					engineNamespace: "default",
					port:            8080,
					wantEndpoint:    fmt.Sprintf(endpointFormat, "fluid", "default", 8080),
				},
			),
			Entry("demo engine in fluid-system namespace",
				testCase{
					name:            "case 2",
					engineName:      "demo",
					engineNamespace: common.NamespaceFluidSystem,
					port:            80,
					wantEndpoint:    fmt.Sprintf(endpointFormat, "demo", common.NamespaceFluidSystem, 80),
				},
			),
		)
	})
})
