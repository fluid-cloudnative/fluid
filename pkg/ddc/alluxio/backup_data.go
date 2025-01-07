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

package alluxio

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/dataflow"
	"os"
	"strconv"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdatabackup "github.com/fluid-cloudnative/fluid/pkg/databackup"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	"gopkg.in/yaml.v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	dataset, err := utils.GetDataset(e.Client, dataBackup.Dataset, dataBackup.Namespace)
	if err != nil {
		return
	}
	dataBackup.OwnerDatasetId = utils.GetDatasetId(dataset.Namespace, dataset.Name, string(dataset.UID))

	pvcName, path, err := utils.ParseBackupRestorePath(databackup.Spec.BackupPath)
	if err != nil {
		return
	}
	dataBackup.PVCName = pvcName
	dataBackup.Path = path

	// inject the node affinity by previous operation pod.
	dataBackup.Affinity, err = dataflow.InjectAffinityByRunAfterOp(e.Client, databackup.Spec.RunAfter, databackup.Namespace, nil)
	if err != nil {
		return "", err
	}

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
		// avoid setting FSGroup with root group
		// dataBackupValue.UserInfo.FSGroup = 0
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
