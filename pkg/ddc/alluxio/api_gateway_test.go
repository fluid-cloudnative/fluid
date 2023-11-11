/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
