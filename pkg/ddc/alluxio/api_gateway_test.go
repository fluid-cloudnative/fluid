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

package alluxio

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Alluxio API gateway", Label("pkg.ddc.alluxio.api_gateway_test.go"), func() {
	const endpointFormat = "%s-master-0.%s:%d"

	DescribeTable("GetAPIGatewayStatus",
		func(engineName string, engineNamespace string, port int32, wantStatus *datav1alpha1.APIGatewayStatus) {
			namespacedName := types.NamespacedName{Name: engineName, Namespace: engineNamespace}
			dataset, runtime := mockFluidObjectsForTests(namespacedName)
			engine := mockAlluxioEngineForTests(dataset, runtime)
			mockedObjects := mockAlluxioObjectsForTests(dataset, runtime, engine)

			svc := mockedObjects.Services[fmt.Sprintf("%s-master-0", engineName)]
			svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{Name: "rest", Port: port})
			engine.Client = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, svc)

			status, err := engine.GetAPIGatewayStatus()

			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(wantStatus))
		},
		Entry("returns the endpoint from the master service in the default namespace",
			"fluid",
			"default",
			int32(8080),
			&datav1alpha1.APIGatewayStatus{Endpoint: fmt.Sprintf(endpointFormat, "fluid", "default", 8080)},
		),
		Entry("returns the endpoint from the master service in the fluid system namespace",
			"demo",
			common.NamespaceFluidSystem,
			int32(80),
			&datav1alpha1.APIGatewayStatus{Endpoint: fmt.Sprintf(endpointFormat, "demo", common.NamespaceFluidSystem, 80)},
		),
	)

	DescribeTable("queryAPIGatewayEndpoint",
		func(engineName string, engineNamespace string, port int32, wantEndpoint string) {
			namespacedName := types.NamespacedName{Name: engineName, Namespace: engineNamespace}
			dataset, runtime := mockFluidObjectsForTests(namespacedName)
			engine := mockAlluxioEngineForTests(dataset, runtime)
			mockedObjects := mockAlluxioObjectsForTests(dataset, runtime, engine)

			svc := mockedObjects.Services[fmt.Sprintf("%s-master-0", engineName)]
			svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{Name: "rest", Port: port})
			engine.Client = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, svc)

			endpoint, err := engine.queryAPIGatewayEndpoint()

			Expect(err).NotTo(HaveOccurred())
			Expect(endpoint).To(Equal(wantEndpoint))
		},
		Entry("returns the endpoint from the default namespace master service",
			"fluid",
			"default",
			int32(8080),
			fmt.Sprintf(endpointFormat, "fluid", "default", 8080),
		),
		Entry("returns the endpoint from the fluid system namespace master service",
			"demo",
			common.NamespaceFluidSystem,
			int32(80),
			fmt.Sprintf(endpointFormat, "demo", common.NamespaceFluidSystem, 80),
		),
	)
})
