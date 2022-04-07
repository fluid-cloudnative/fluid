/*
Copyright 2022 The Fluid Authors.

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

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

// Query the API Gateway status
func (e *GooseFSEngine) GetAPIGatewayStatus() (status *datav1alpha1.APIGatewayStatus, err error) {
	endpoint, err := e.queryAPIGatewayEndpoint()
	if err != nil {
		e.Log.Error(err, "Failed to get HCFS Endpoint")
		return status, err
	}

	status = &datav1alpha1.APIGatewayStatus{
		Endpoint: endpoint,
	}
	return
}

// query the api endpoint
func (e *GooseFSEngine) queryAPIGatewayEndpoint() (endpoint string, err error) {

	var (
		serviceName = fmt.Sprintf("%s-master-0", e.name)
		host        = fmt.Sprintf("%s.%s", serviceName, e.namespace)
	)

	svc, err := kubeclient.GetServiceByName(e.Client, serviceName, e.namespace)
	if err != nil {
		e.Log.Error(err, "Failed to get Endpoint")
		return endpoint, err
	}

	if svc == nil {
		e.Log.Error(fmt.Errorf("failed to find the svc %s in %s", e.name, e.namespace), "failed to find the svc, it's nil")
		return
	}

	for _, port := range svc.Spec.Ports {
		if port.Name == "rest" {
			endpoint = fmt.Sprintf("%s:%d", host, port.Port)
			return
		}
	}

	return
}
