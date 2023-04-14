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
