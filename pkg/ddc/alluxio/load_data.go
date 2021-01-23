/*

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
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/controllers/v1alpha1/requestcontext"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
)

// LoadData load the data
func (e *AlluxioEngine) CreateDataLoadJob (ctx requestcontext.ReconcileRequestContext) (releaseName string, jobName string, err error) {
	log := ctx.Log.WithName("createDataLoadJob")

	// 1. Check if the helm release already exists
	releaseName = utils.GetDataLoadReleaseName(ctx.Name)
	jobName = utils.GetDataLoadJobName(releaseName)
	var existed bool
	existed, err = helm.CheckRelease(releaseName, ctx.Namespace)
	if err != nil {
		log.Error(err, "failed to check if release exists", "releaseName", releaseName, "namespace", ctx.Namespace)
		return
	}

	// 2. install the helm chart if not exists and requeue
	if !existed {
		log.Info("DataLoad job helm chart not installed yet, will install")
		valueFileName, err := e.generateDataLoadValueFile(ctx.DataLoad)
		if err != nil {
			log.Error(err, "failed to generate dataload chart's value file")
			return
		}
		chartName := utils.GetChartsDirectory() + "/" + cdataload.DATALOAD_CHART
		err = helm.InstallRelease(releaseName, ctx.Namespace, valueFileName, chartName)
		if err != nil {
			log.Error(err, "failed to install dataload chart")
			return
		}
	}
	return
}

// generateDataLoadValueFile builds a DataLoadValue by extracted specifications from the given DataLoad, and
// marshals the DataLoadValue to a temporary yaml file where stores values that'll be used by fluid dataloader helm chart
func (e *AlluxioEngine) generateDataLoadValueFile(dataload v1alpha1.DataLoad) (valueFileName string, err error) {
	targetDataset, err := utils.GetDataset(e.Client, dataload.Spec.Dataset.Name, dataload.Spec.Dataset.Namespace)
	if err != nil {
		return "", err
	}

	imageName := "registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio"
	imageTag := "2.3.0-SNAPSHOT-238b7eb"
	imageName, imageTag = docker.GetImageRepoTagFromEnv(common.ALLUXIO_DATALOAD_IMAGE_ENV, imageName, imageTag)
	image := fmt.Sprintf("%s:%s", imageName, imageTag)

	dataloadInfo := cdataload.DataLoadInfo{
		BackoffLimit:  3,
		TargetDataset: dataload.Spec.Dataset.Name,
		LoadMetadata:  dataload.Spec.LoadMetadata,
		Image:         image,
	}

	targetPaths := []cdataload.TargetPath{}
	for _, target := range dataload.Spec.Target {
		fluidNative := isTargetPathUnderFluidNativeMounts(target.Path, *targetDataset)
		targetPaths = append(targetPaths, cdataload.TargetPath{
			Path:        target.Path,
			Replicas:    target.Replicas,
			FluidNative: fluidNative,
		})
	}
	dataloadInfo.TargetPaths = targetPaths
	dataLoadValue := cdataload.DataLoadValue{DataLoadInfo: dataloadInfo}
	data, err := yaml.Marshal(dataLoadValue)
	if err != nil {
		return
	}

	valueFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("%s-%s-loader-values.yaml", dataload.Namespace, dataload.Name))
	if err != nil {
		return
	}
	err = ioutil.WriteFile(valueFile.Name(), data, 0400)
	if err != nil {
		return
	}
	return valueFile.Name(), nil
}

// isTargetPathUnderFluidNativeMounts checks if targetPath is a subpath under some given native mount point.
// We check this for the reason that native mount points need extra metadata sync operations.
func isTargetPathUnderFluidNativeMounts(targetPath string, dataset v1alpha1.Dataset) bool {
	for _, mount := range dataset.Spec.Mounts {
		mountPointOnDDCEngine := fmt.Sprintf("/%s", mount.Name)
		if mount.Path != "" {
			mountPointOnDDCEngine = mount.Path
		}

		//todo(xuzhihao): HasPrefix is not enough.
		if strings.HasPrefix(targetPath, mountPointOnDDCEngine) &&
			(strings.HasPrefix(mount.MountPoint, common.PathScheme) || strings.HasPrefix(mount.MountPoint, common.VolumeScheme)) {
			return true
		}
	}
	return false
}