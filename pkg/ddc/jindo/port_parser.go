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

package jindo

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
	if conf, ok := configMap.Data["bigboot.cfg"]; ok {
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
