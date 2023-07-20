/*
Copyright 2023 The Fluid Authors.

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

package jindocache

import (
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

// Query the hcfs status
func (e *JindoCacheEngine) GetHCFSStatus() (status *datav1alpha1.HCFSStatus, err error) {

	endpoint, err := e.queryHCFSEndpoint()
	if err != nil {
		e.Log.Error(err, "Failed to get HCFS Endpoint")
		return status, err
	}

	status = &datav1alpha1.HCFSStatus{
		Endpoint:                    endpoint,
		UnderlayerFileSystemVersion: "",
	}
	return
}

func (e *JindoCacheEngine) GetHCFSStatusWithoutMaster() (status *datav1alpha1.HCFSStatus, err error) {
	return &datav1alpha1.HCFSStatus{
		Endpoint:                    "N/A",
		UnderlayerFileSystemVersion: "",
	}, err
}

// query the hcfs endpoint
func (e *JindoCacheEngine) queryHCFSEndpoint() (endpoint string, err error) {

	var (
		serviceName = fmt.Sprintf("%s-jindofs-master-0", e.name)
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
		if port.Name == "rpc" {
			endpoint = fmt.Sprintf("%s:%d", host, port.Port)
			return
		}
	}

	return
}
