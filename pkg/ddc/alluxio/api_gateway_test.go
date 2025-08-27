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
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestGetAPIGatewayStatus(t *testing.T) {
	endpointFormat := "%s-master-0.%s:%d"
	testCases := map[string]struct {
		engineName      string
		engineNamespace string
		port            int32
		wantStatus      *datav1alpha1.APIGatewayStatus
	}{
		"test GetAPIGatewayStatus case 1": {
			engineName:      "fluid",
			engineNamespace: "default",
			port:            8080,
			wantStatus: &datav1alpha1.APIGatewayStatus{
				Endpoint: fmt.Sprintf(endpointFormat, "fluid", "default", 8080),
			},
		},
		"test GetAPIGatewayStatus case 2": {
			engineName:      "demo",
			engineNamespace: common.NamespaceFluidSystem,
			port:            80,
			wantStatus: &datav1alpha1.APIGatewayStatus{
				Endpoint: fmt.Sprintf(endpointFormat, "demo", common.NamespaceFluidSystem, 80),
			},
		},
	}

	for _, item := range testCases {
		namespacedName := types.NamespacedName{Name: item.engineName, Namespace: item.engineNamespace}
		dataset, runtime := mockFluidObjectsForTests(namespacedName)
		engine := mockAlluxioEngineForTests(dataset, runtime)
		mockedObjects := mockAlluxioObjectsForTests(dataset, runtime, engine)

		svc := mockedObjects.Services[fmt.Sprintf("%s-master-0", item.engineName)]
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{Name: "rest", Port: item.port})
		client := fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, svc)
		engine.Client = client

		// e := mockAlluxioEngineWithClient(item.engineName, item.engineNamespace, item.port)
		got, err := engine.GetAPIGatewayStatus()

		assert.NoError(t, err)
		assert.Equal(t, got, item.wantStatus)
		// if !reflect.DeepEqual(got, item.wantStatus) {
		// t.Errorf("%s check failure,want:%v,got:%v", k, item.wantStatus, got)
		// }

	}
}

// TestQueryAPIGatewayEndpoint tests whether the Alluxio engine's queryAPIGatewayEndpoint method
// correctly returns the API gateway access endpoint.
// It uses two different test cases with varying engineName, engineNamespace, and port values,
// and verifies that the generated endpoint string matches the expected format "%s-master-0.%s:%d".
// The test cases cover both the default namespace and the system namespace scenarios.
func TestQueryAPIGatewayEndpoint(t *testing.T) {
	endpointFormat := "%s-master-0.%s:%d"
	testCases := map[string]struct {
		engineName      string
		engineNamespace string
		port            int32
		wantEndpoint    string
	}{
		"test GetAPIGatewayStatus case 1": {
			engineName:      "fluid",
			engineNamespace: "default",
			port:            8080,
			wantEndpoint:    fmt.Sprintf(endpointFormat, "fluid", "default", 8080),
		},
		"test GetAPIGatewayStatus case 2": {
			engineName:      "demo",
			engineNamespace: common.NamespaceFluidSystem,
			port:            80,
			wantEndpoint:    fmt.Sprintf(endpointFormat, "demo", common.NamespaceFluidSystem, 80),
		},
	}

	for _, item := range testCases {
		namespacedName := types.NamespacedName{Name: item.engineName, Namespace: item.engineNamespace}
		dataset, runtime := mockFluidObjectsForTests(namespacedName)
		engine := mockAlluxioEngineForTests(dataset, runtime)
		mockedObjects := mockAlluxioObjectsForTests(dataset, runtime, engine)

		svc := mockedObjects.Services[fmt.Sprintf("%s-master-0", item.engineName)]
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{Name: "rest", Port: item.port})
		client := fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, svc)
		engine.Client = client

		got, err := engine.queryAPIGatewayEndpoint()

		assert.NoError(t, err)
		assert.Equal(t, got, item.wantEndpoint)
	}
}
