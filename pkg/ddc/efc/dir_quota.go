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
