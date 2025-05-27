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
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	for k, item := range testCases {
		e := mockAlluxioEngineWithClient(item.engineName, item.engineNamespace, item.port)
		got, _ := e.GetAPIGatewayStatus()

		if !reflect.DeepEqual(got, item.wantStatus) {
			t.Errorf("%s check failure,want:%v,got:%v", k, item.wantStatus, got)
		}

	}
}

func mockAlluxioEngineWithClient(name, ns string, port int32) *AlluxioEngine {

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

	e := &AlluxioEngine{
		Client:    mockClient,
		name:      name,
		namespace: ns,
	}
	return e
}

// TestQueryAPIGatewayEndpoint 测试 Alluxio 引擎的 queryAPIGatewayEndpoint 方法是否能正确返回 API 网关的访问地址。
// 它使用两个不同的测试用例，分别传入不同的 engineName、engineNamespace 和 port，
// 验证生成的 endpoint 字符串是否符合预期格式 "%s-master-0.%s:%d"。
// 测试用例覆盖了默认命名空间和系统命名空间两种情况。
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

	for k, item := range testCases {
		e := mockAlluxioEngineWithClient(item.engineName, item.engineNamespace, item.port)
		got, _ := e.queryAPIGatewayEndpoint()

		if !reflect.DeepEqual(got, item.wantEndpoint) {
			t.Errorf("%s check failure,want:%v,got:%v", k, item.wantEndpoint, got)
		}

	}
}
