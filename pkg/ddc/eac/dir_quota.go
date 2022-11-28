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

package eac

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/nas"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
)

func (e *EACEngine) setDirQuota() (response *nas.SetDirQuotaResponse, err error) {
	runtime, err := e.getRuntime()
	if err != nil {
		return nil, err
	}

	configMapName := e.getConfigmapName()
	configMap, err := kubeclient.GetConfigmapByName(e.Client, configMapName, runtime.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "GetConfigMapByName fail when setDirQuota")
	}

	serviceAddr, fileSystemId, dirPath, err := parseDirInfoFromConfigMap(configMap)
	if err != nil {
		return nil, errors.Wrap(err, "parseDirInfoFromConfigMap fail when setDirQuota")
	}

	config := sdk.NewConfig()
	accessKeyID, accessKeySecret, err := e.getEACSecret()
	if err != nil {
		return nil, err
	}
	credential := credentials.NewAccessKeyCredential(accessKeyID, accessKeySecret)
	client, err := nas.NewClientWithOptions(serviceAddr, config, credential)
	if err != nil {
		return nil, err
	}

	request := nas.CreateSetDirQuotaRequest()
	request.Scheme = "https"
	request.QuotaType = "Accounting"
	request.UserType = "AllUsers"
	request.FileSystemId = fileSystemId
	request.Path = dirPath

	response, err = client.SetDirQuota(request)
	if err != nil {
		return nil, err
	}
	e.Log.Info("SetDirQuota success", "ServiceAddr", serviceAddr, "FileSystemId", fileSystemId, "DirPath", dirPath, "Response", response)

	return
}

func (e *EACEngine) describeDirQuota() (response *nas.DescribeDirQuotasResponse, err error) {
	runtime, err := e.getRuntime()
	if err != nil {
		return nil, err
	}

	configMapName := e.getConfigmapName()
	configMap, err := kubeclient.GetConfigmapByName(e.Client, configMapName, runtime.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "GetConfigMapByName fail when describeDirQuota")
	}

	serviceAddr, fileSystemId, dirPath, err := parseDirInfoFromConfigMap(configMap)
	if err != nil {
		return nil, errors.Wrap(err, "parseDirInfoFromConfigMap fail when describeDirQuota")
	}

	config := sdk.NewConfig()
	accessKeyID, accessKeySecret, err := e.getEACSecret()
	if err != nil {
		return nil, err
	}
	credential := credentials.NewAccessKeyCredential(accessKeyID, accessKeySecret)
	client, err := nas.NewClientWithOptions(serviceAddr, config, credential)
	if err != nil {
		return nil, err
	}

	request := nas.CreateDescribeDirQuotasRequest()
	request.Scheme = "https"
	request.FileSystemId = fileSystemId
	request.Path = dirPath

	response, err = client.DescribeDirQuotas(request)
	if err != nil {
		return nil, err
	}
	e.Log.Info("DescribeDirQuota success", "ServiceAddr", serviceAddr, "FileSystemId", fileSystemId, "DirPath", dirPath, "Response", response)

	return
}
