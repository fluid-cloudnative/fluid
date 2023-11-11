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

package alluxio

import (
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdatabackup "github.com/fluid-cloudnative/fluid/pkg/databackup"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	"gopkg.in/yaml.v2"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
)

// generateDataBackupValueFile builds a DataBackupValueFile by extracted specifications from the given DataBackup, and
// marshals the DataBackup to a temporary yaml file where stores values that'll be used by fluid dataBackup helm chart
func (e *AlluxioEngine) generateDataBackupValueFile(ctx cruntime.ReconcileRequestContext, object client.Object) (valueFileName string, err error) {
	logger := ctx.Log.WithName("generateDataBackupValueFile")

	databackup, ok := object.(*datav1alpha1.DataBackup)
	if !ok {
		err = fmt.Errorf("object %v is not a DataBackup", object)
		return "", err
	}

	// get the runAs and initUsers imageInfo from runtime
	runtime, err := e.getRuntime()
	if err != nil {
		return "", err
	}

	masterPodName, containerName := e.getMasterPodInfo()
	if runtime.Spec.Replicas > 1 {
		fileUtils := operations.NewAlluxioFileUtils(masterPodName, containerName, runtime.GetNamespace(), ctx.Log)
		masterPodName, err = fileUtils.MasterPodName()
		if err != nil {
			return "", err
		}
	}

	masterPod, err := e.getMasterPod(masterPodName, e.namespace)
	if err != nil {
		return
	}

	nodeName, ip, rpcPort := utils.GetAddressOfMaster(masterPod)

	var imageEnv, defaultImage string

	imageName, imageTag := docker.GetWorkerImage(e.Client, databackup.Spec.Dataset, common.AlluxioRuntime, databackup.Namespace)
	javaEnv := "-Dalluxio.master.hostname=" + ip + " -Dalluxio.master.rpc.port=" + strconv.Itoa(int(rpcPort))

	if len(imageName) == 0 {
		imageEnv = common.AlluxioRuntimeImageEnv
		defaultImage = common.DefaultAlluxioRuntimeImage

		imageName = docker.GetImageRepoFromEnv(imageEnv)
		if len(imageName) == 0 {
			defaultImageInfo := strings.Split(defaultImage, ":")
			if len(defaultImageInfo) < 1 {
				logger.Error(fmt.Errorf("ImageInfo"), "invalid default databackup image!")
				return
			} else {
				imageName = defaultImageInfo[0]
			}
		}
	}

	if len(imageTag) == 0 {
		imageEnv = common.AlluxioRuntimeImageEnv
		defaultImage = common.DefaultAlluxioRuntimeImage

		imageTag = docker.GetImageTagFromEnv(imageEnv)
		if len(imageTag) == 0 {
			defaultImageInfo := strings.Split(defaultImage, ":")
			if len(defaultImageInfo) < 1 {
				logger.Error(fmt.Errorf("ImageInfo"), "invalid default databackup image!")
				return
			} else {
				imageTag = defaultImageInfo[1]
			}
		}
	}

	image := fmt.Sprintf("%s:%s", imageName, imageTag)

	workdir := os.Getenv("FLUID_WORKDIR")
	if workdir == "" {
		workdir = "/tmp"
	}

	// image pull secrets
	// if the environment variable is not set, it is still an empty slice
	imagePullSecrets := docker.GetImagePullSecretsFromEnv(common.EnvImagePullSecretsKey)

	dataBackup := cdatabackup.DataBackup{
		Namespace:        databackup.Namespace,
		Dataset:          databackup.Spec.Dataset,
		Name:             databackup.Name,
		NodeName:         nodeName,
		Image:            image,
		JavaEnv:          javaEnv,
		Workdir:          workdir,
		RuntimeType:      common.AlluxioRuntime,
		ImagePullSecrets: imagePullSecrets,
	}
	pvcName, path, err := utils.ParseBackupRestorePath(databackup.Spec.BackupPath)
	if err != nil {
		return
	}
	dataBackup.PVCName = pvcName
	dataBackup.Path = path

	dataBackupValue := cdatabackup.DataBackupValue{DataBackup: dataBackup}

	dataBackupValue.InitUsers = common.InitUsers{
		Enabled: false,
	}

	var runAs = runtime.Spec.RunAs
	// databackup.Spec.RunAs > runtime.Spec.RunAs > root
	if databackup.Spec.RunAs != nil {
		runAs = databackup.Spec.RunAs
	}

	if runAs != nil {
		dataBackupValue.UserInfo.User = int(*runAs.UID)
		dataBackupValue.UserInfo.Group = int(*runAs.GID)
		dataBackupValue.UserInfo.FSGroup = 0
		dataBackupValue.InitUsers = common.InitUsers{
			Enabled:  true,
			EnvUsers: utils.GetInitUserEnv(runAs),
			Dir:      utils.GetBackupUserDir(dataBackup.Namespace, dataBackup.Name),
		}
	}

	image = runtime.Spec.InitUsers.Image
	imageTag = runtime.Spec.InitUsers.ImageTag
	imagePullPolicy := runtime.Spec.InitUsers.ImagePullPolicy

	dataBackupValue.InitUsers.Image, dataBackupValue.InitUsers.ImageTag, dataBackupValue.InitUsers.ImagePullPolicy = docker.ParseInitImage(image, imageTag, imagePullPolicy, common.DefaultInitImageEnv)

	data, err := yaml.Marshal(dataBackupValue)
	if err != nil {
		return
	}

	valueFile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("%s-%s-%s-backuper-values.yaml", databackup.Namespace, databackup.Name, dataBackup.RuntimeType))
	if err != nil {
		return
	}

	err = os.WriteFile(valueFile.Name(), data, 0400)
	if err != nil {
		return
	}

	return valueFile.Name(), nil
}
