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

package juicefs

import (
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

// GetCacheInfoFromConfigmap get cache info from configmap
func GetCacheInfoFromConfigmap(client client.Client, name string, namespace string) (cacheinfo map[string]string, err error) {

	configMapName := fmt.Sprintf("%s-juicefs-values", name)
	configMap, err := kubeclient.GetConfigmapByName(client, configMapName, namespace)
	if err != nil {
		return nil, errors.Wrap(err, "GetConfigMapByName error when GetCacheInfoFromConfigmap")
	}

	cacheinfo, err = parseCacheInfoFromConfigMap(configMap)
	if err != nil {
		return nil, errors.Wrap(err, "parsePortsFromConfigMap when GetReservedPorts")
	}

	return cacheinfo, nil
}

// parseCacheInfoFromConfigMap extracts port usage information given a configMap
func parseCacheInfoFromConfigMap(configMap *v1.ConfigMap) (cacheinfo map[string]string, err error) {
	var value JuiceFS
	configmapinfo := map[string]string{}
	if v, ok := configMap.Data["data"]; ok {
		if err := yaml.Unmarshal([]byte(v), &value); err != nil {
			return nil, err
		}
		configmapinfo[MountPath] = value.Fuse.MountPath
		configmapinfo[Edition] = value.Edition
	}
	return configmapinfo, nil
}

func GetFSInfoFromConfigMap(client client.Client, name string, namespace string) (info map[string]string, err error) {
	configMapName := fmt.Sprintf("%s-juicefs-values", name)
	configMap, err := kubeclient.GetConfigmapByName(client, configMapName, namespace)
	if err != nil {
		return nil, errors.Wrap(err, "GetConfigMapByName error when GetCacheInfoFromConfigmap")
	}
	return parseFSInfoFromConfigMap(configMap)
}

func parseFSInfoFromConfigMap(configMap *v1.ConfigMap) (info map[string]string, err error) {
	var value JuiceFS
	info = map[string]string{}
	if v, ok := configMap.Data["data"]; ok {
		if err = yaml.Unmarshal([]byte(v), &value); err != nil {
			return
		}
		info[MetaurlSecret] = value.Configs.MetaUrlSecret
		info[MetaurlSecretKey] = value.Configs.MetaUrlSecretKey
		info[TokenSecret] = value.Configs.TokenSecret
		info[TokenSecretKey] = value.Configs.TokenSecretKey
		info[AccessKeySecret] = value.Configs.AccessKeySecret
		info[AccessKeySecretKey] = value.Configs.AccessKeySecretKey
		info[SecretKeySecret] = value.Configs.SecretKeySecret
		info[SecretKeySecretKey] = value.Configs.SecretKeySecretKey
		info[FormatCmd] = value.Configs.FormatCmd
		info[Name] = value.Configs.Name
		info[Edition] = value.Edition
		return
	}
	return
}
