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

package goosefs

import (
	"context"
	"fmt"
	"strconv"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var propertiesToCheck = []string{
	"goosefs.master.rpc.port",
	"goosefs.master.web.port",
	"goosefs.worker.rpc.port",
	"goosefs.worker.web.port",
	"goosefs.job.master.rpc.port",
	"goosefs.job.master.web.port",
	"goosefs.job.worker.rpc.port",
	"goosefs.job.worker.web.port",
	"goosefs.job.worker.data.port",
	"goosefs.proxy.web.port",
	"goosefs.master.embedded.journal.port",
	"goosefs.job.master.embedded.journal.port",
}

// GetReservedPorts defines restoration logic for goosefsRuntime
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
			if accelerateRuntime.Type != "goosefs" {
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
func parsePortsFromConfigMap(configMap *v1.ConfigMap) (ports []int, err error) {
	var value GooseFS
	if v, ok := configMap.Data["data"]; ok {
		if err := yaml.Unmarshal([]byte(v), &value); err != nil {
			return nil, err
		}
		for _, property := range propertiesToCheck {
			if portStr, ok := value.Properties[property]; ok {
				portInt, err := strconv.Atoi(portStr)
				if err != nil {
					return nil, err
				}
				ports = append(ports, portInt)
			}
		}
	}
	return ports, nil
}
