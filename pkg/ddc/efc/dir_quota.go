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

package efc

//func (mountInfo MountInfo) SetDirQuota() (response *nas.SetDirQuotaResponse, err error) {
//	serviceAddr, fileSystemId, dirPath := mountInfo.ServiceAddr, mountInfo.FileSystemId, mountInfo.DirPath
//	accessKeyID, accessKeySecret := mountInfo.AccessKeyID, mountInfo.AccessKeySecret
//
//	config := sdk.NewConfig()
//	credential := credentials.NewAccessKeyCredential(accessKeyID, accessKeySecret)
//	client, err := nas.NewClientWithOptions(serviceAddr, config, credential)
//	if err != nil {
//		return nil, err
//	}
//
//	request := nas.CreateSetDirQuotaRequest()
//	request.Scheme = "https"
//	request.QuotaType = "Accounting"
//	request.UserType = "AllUsers"
//	request.FileSystemId = fileSystemId
//	request.Path = dirPath
//
//	response, err = client.SetDirQuota(request)
//	if err != nil {
//		return nil, err
//	}
//
//	return
//}
//
//func (mountInfo MountInfo) DescribeDirQuota() (response *nas.DescribeDirQuotasResponse, err error) {
//	serviceAddr, fileSystemId, dirPath := mountInfo.ServiceAddr, mountInfo.FileSystemId, mountInfo.DirPath
//	accessKeyID, accessKeySecret := mountInfo.AccessKeyID, mountInfo.AccessKeySecret
//
//	config := sdk.NewConfig()
//	credential := credentials.NewAccessKeyCredential(accessKeyID, accessKeySecret)
//	client, err := nas.NewClientWithOptions(serviceAddr, config, credential)
//	if err != nil {
//		return nil, err
//	}
//
//	request := nas.CreateDescribeDirQuotasRequest()
//	request.Scheme = "https"
//	request.FileSystemId = fileSystemId
//	request.Path = dirPath
//
//	response, err = client.DescribeDirQuotas(request)
//	if err != nil {
//		return nil, err
//	}
//
//	return
//}
