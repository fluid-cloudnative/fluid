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

package juicefs

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

// GetReservedPorts defines restoration logic for JuiceFSRuntime
func GetReservedPorts(client client.Client) (ports []int, err error) {
	var datasets v1alpha1.DatasetList
	err = client.List(context.TODO(), &datasets)
	if err != nil {
		return nil, errors.Wrap(err, "can't list datasets when GetReservedPorts")
	}

	for _, dataset := range datasets.Items {
		if len(dataset.Status.Runtimes) != 0 {
			// Assume there is only one runtime with category "Accelerate"
			accelerateRuntime := dataset.Status.Runtimes[0]
			if accelerateRuntime.Type != common.JuiceFSRuntime {
				continue
			}
			configMapName := fmt.Sprintf("%s-%s-values", accelerateRuntime.Name, accelerateRuntime.Type)
			configMap, err := kubeclient.GetConfigmapByName(client, configMapName, accelerateRuntime.Namespace)
			if err != nil {
				return nil, errors.Wrap(err, "GetConfigMapByName when GetReservedPorts")
			}

			if configMap == nil {
				continue
			}

			reservedPorts, err := parsePortsFromConfigMap(configMap)
			if err != nil {
				return nil, errors.Wrap(err, "parsePortsFromConfigMap when GetReservedPorts")
			}
			ports = append(ports, reservedPorts...)
		}
	}
	return ports, nil
}

// parsePortsFromConfigMap extracts port usage information given a configMap
func parsePortsFromConfigMap(configMap *corev1.ConfigMap) (ports []int, err error) {
	var value JuiceFS
	if v, ok := configMap.Data["data"]; ok {
		if err := yaml.Unmarshal([]byte(v), &value); err != nil {
			return nil, err
		}

		if value.Worker.HostNetwork && value.Worker.MetricsPort != nil {
			ports = append(ports, *value.Worker.MetricsPort)
		}
		if value.Fuse.HostNetwork && value.Fuse.MetricsPort != nil {
			ports = append(ports, *value.Fuse.MetricsPort)
		}
	}
	return ports, nil
}
