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

package jindofsx

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var propertiesToCheck = []string{
	// "client.storage.rpc.port" not included here cause it is same with "storage.rpc.port"
	"storage.rpc.port",
	"namespace.rpc.port",
}

// GetReservedPorts defines restoration logic for JindoRuntime
func GetReservedPorts(client client.Client) (ports []int, err error) {
	var datasets v1alpha1.DatasetList
	err = client.List(context.TODO(), &datasets)
	if err != nil {
		return nil, err
	}

	for _, dataset := range datasets.Items {
		if len(dataset.Status.Runtimes) != 0 {
			// Assume there is only one runtime and it is in category "Accelerate"
			accelerateRuntime := dataset.Status.Runtimes[0]
			if accelerateRuntime.Type != "jindo" {
				continue
			}
			configMapName := fmt.Sprintf("%s-jindofs-config", accelerateRuntime.Name)
			configMap, err := kubeclient.GetConfigmapByName(client, configMapName, accelerateRuntime.Namespace)
			if err != nil {
				return nil, err
			}

			if configMap == nil {
				continue
			}

			reservedPorts, err := parsePortsFromConfigMap(configMap)
			if err != nil {
				return nil, err
			}
			ports = append(ports, reservedPorts...)
		}
	}
	return ports, nil
}

// parsePortsFromConfigMap extracts port usage information given a configMap
func parsePortsFromConfigMap(configMap *v1.ConfigMap) (ports []int, err error) {
	if conf, ok := configMap.Data["jindofsx.cfg"]; ok {
		cfgConfs := strings.Split(conf, "\n")
		for _, cfgConf := range cfgConfs {
			for _, toCheck := range propertiesToCheck {
				if strings.HasPrefix(cfgConf, toCheck) {
					portStr := strings.Split(cfgConf, " = ")[1]
					portInt, err := strconv.Atoi(portStr)
					if err != nil {
						return nil, err
					}
					ports = append(ports, portInt)
				}
			}
		}
	}
	return ports, nil
}
