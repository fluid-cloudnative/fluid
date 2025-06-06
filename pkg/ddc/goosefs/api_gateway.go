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

// queryAPIGatewayEndpoint retrieves the REST port information of GooseFS master from Kubernetes services
// and constructs the API gateway access address using the service name and namespace.
// It returns the constructed endpoint address if successful; otherwise, it returns an error.
// This function depends on the Kubernetes client to fetch service details.
//
// Parameters:
// - none
//
// Returns:
// - endpoint (string): The constructed API gateway endpoint address.
// - err (error): An error if the query fails, otherwise nil.
func (e *GooseFSEngine) queryAPIGatewayEndpoint() (endpoint string, err error) {
    // Define the service name in the format "{engine name}-master-0"
    var (
        serviceName = fmt.Sprintf("%s-master-0", e.name)
        // Construct the full hostname of the service in the format "{service name}.{namespace}"
        host = fmt.Sprintf("%s.%s", serviceName, e.namespace)
    )

    // Retrieve the service object with the specified name and namespace using the Kubernetes client
    svc, err := kubeclient.GetServiceByName(e.Client, serviceName, e.namespace)
    if err != nil {
        // Log an error and return if service retrieval fails
        e.Log.Error(err, "Failed to get Endpoint")
        return endpoint, err
    }

    // Check if the retrieved service object is nil
    if svc == nil {
        // Log an error indicating that the service was not found
        e.Log.Error(fmt.Errorf("failed to find the svc %s in %s", e.name, e.namespace), "failed to find the svc, it's nil")
        return
    }

    // Iterate through the service's port list to find the port named "rest"
    for _, port := range svc.Spec.Ports {
        if port.Name == "rest" {
            // Construct the endpoint address in the format "{host}:{port}"
            endpoint = fmt.Sprintf("%s:%d", host, port.Port)
            return
        }
    }

    // Return an empty endpoint if no matching port is found
    return
}
