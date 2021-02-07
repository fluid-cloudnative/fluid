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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"os"
	"strings"
)

// CreateDataLoadJob creates the job to load data
func (e *AlluxioEngine) CreateDataLoadJob(ctx cruntime.ReconcileRequestContext, targetDataload datav1alpha1.DataLoad) (err error) {
	log := ctx.Log.WithName("createDataLoadJob")

	// 1. Check if the helm release already exists
	releaseName := utils.GetDataLoadReleaseName(targetDataload.Name)
	jobName := utils.GetDataLoadJobName(releaseName)
	var existed bool
	existed, err = helm.CheckRelease(releaseName, targetDataload.Namespace)
	if err != nil {
		log.Error(err, "failed to check if release exists", "releaseName", releaseName, "namespace", targetDataload.Namespace)
		return err
	}

	// 2. install the helm chart if not exists
	if !existed {
		log.Info("DataLoad job helm chart not installed yet, will install")
		valueFileName, err := e.generateDataLoadValueFile(targetDataload)
		if err != nil {
			log.Error(err, "failed to generate dataload chart's value file")
			return err
		}
		chartName := utils.GetChartsDirectory() + "/" + cdataload.DATALOAD_CHART
		err = helm.InstallRelease(releaseName, targetDataload.Namespace, valueFileName, chartName)
		if err != nil {
			log.Error(err, "failed to install dataload chart")
			return err
		}
		log.Info("DataLoad job helm chart successfully installed", "namespace", targetDataload.Namespace, "releaseName", releaseName)
		ctx.Recorder.Eventf(&targetDataload, v1.EventTypeNormal, common.DataLoadJobStarted, "The DataLoad job %s started", jobName)
	}
	return err
}

// GetDataLoadJobStatus checks whether the DataLoad job is finished or not
func (e *AlluxioEngine) GetDataLoadJobStatus(ctx cruntime.ReconcileRequestContext, targetDataload datav1alpha1.DataLoad) (status batchv1.JobConditionType, err error){
	log := ctx.Log.WithName("checkDataLoadJobStatus")

	// Check running status of the DataLoad job
	releaseName := utils.GetDataLoadReleaseName(targetDataload.Name)
	jobName := utils.GetDataLoadJobName(releaseName)
	log.V(1).Info("DataLoad chart already existed, check its running status")
	job, err := utils.GetDataLoadJob(e.Client, jobName, targetDataload.Namespace)
	if err != nil {
		// helm release found but job missing, delete the helm release and requeue
		if utils.IgnoreNotFound(err) == nil {
			log.Info("Related Job missing, will delete helm chart and retry", "namespace", targetDataload.Namespace, "jobName", jobName)
			if err = helm.DeleteReleaseIfExists(releaseName, targetDataload.Namespace); err != nil {
				log.Error(err, "can't delete dataload release", "namespace", targetDataload.Namespace, "releaseName", releaseName)
				return "", err
			}
		}
		// other error
		log.Error(err, "can't get dataload job", "namespace", targetDataload.Namespace, "jobName", jobName)
		return "", err
	}
	if len(job.Status.Conditions) != 0 {
		status = job.Status.Conditions[0].Type
	}
	return status, err
}

// generateDataLoadValueFile builds a DataLoadValue by extracted specifications from the given DataLoad, and
// marshals the DataLoadValue to a temporary yaml file where stores values that'll be used by fluid dataloader helm chart
func (e *AlluxioEngine) generateDataLoadValueFile(targetDataload datav1alpha1.DataLoad) (valueFileName string, err error) {
	targetDataset, err := utils.GetDataset(e.Client, targetDataload.Spec.Dataset.Name, targetDataload.Spec.Dataset.Namespace)
	if err != nil {
		return "", err
	}

	imageName, imageTag := docker.GetWorkerImage(e.Client, targetDataload.Spec.Dataset.Name, "alluxio", targetDataload.Spec.Dataset.Namespace)
	image := fmt.Sprintf("%s:%s", imageName, imageTag)

	dataloadInfo := cdataload.DataLoadInfo{
		BackoffLimit:  3,
		TargetDataset: targetDataload.Spec.Dataset.Name,
		LoadMetadata:  targetDataload.Spec.LoadMetadata,
		Image:         image,
	}

	targetPaths := []cdataload.TargetPath{}
	for _, target := range targetDataload.Spec.Target {
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

	valueFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("%s-%s-loader-values.yaml", targetDataload.Namespace, targetDataload.Name))
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
func isTargetPathUnderFluidNativeMounts(targetPath string, dataset datav1alpha1.Dataset) bool {
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
