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

package juicefs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/juicefs/operations"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
)

// CreateDataLoadJob creates the job to load data
func (j *JuiceFSEngine) CreateDataLoadJob(ctx cruntime.ReconcileRequestContext, targetDataload datav1alpha1.DataLoad) (err error) {
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
		valueFileName, err := j.generateDataLoadValueFile(ctx, targetDataload)
		if err != nil {
			log.Error(err, "failed to generate dataload chart's value file")
			return err
		}
		chartName := utils.GetChartsDirectory() + "/" + cdataload.DataloadChart + "/" + common.JuiceFSRuntime
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

// generateDataLoadValueFile builds a DataLoadValue by extracted specifications from the given DataLoad, and
// marshals the DataLoadValue to a temporary yaml file where stores values that'll be used by fluid dataloader helm chart
func (j *JuiceFSEngine) generateDataLoadValueFile(r cruntime.ReconcileRequestContext, dataload datav1alpha1.DataLoad) (valueFileName string, err error) {
	targetDataset, err := utils.GetDataset(r.Client, dataload.Spec.Dataset.Name, dataload.Spec.Dataset.Namespace)
	if err != nil {
		return "", err
	}
	j.Log.Info("target dataset", "dataset", targetDataset)

	imageName, imageTag := docker.GetWorkerImage(r.Client, dataload.Spec.Dataset.Name, "juicefs", dataload.Spec.Dataset.Namespace)

	if len(imageName) == 0 {
		defaultImageInfo := strings.Split(common.DefaultJuiceFSRuntimeImage, ":")
		if len(defaultImageInfo) < 1 {
			panic("invalid default dataload image!")
		} else {
			imageName = defaultImageInfo[0]
		}
	}

	if len(imageTag) == 0 {
		defaultImageInfo := strings.Split(common.DefaultJuiceFSRuntimeImage, ":")
		if len(defaultImageInfo) < 2 {
			panic("invalid default dataload image!")
		} else {
			imageTag = defaultImageInfo[1]
		}
	}

	cacheinfo, err := GetCacheInfoFromConfigmap(j.Client, dataload.Spec.Dataset.Name, dataload.Spec.Dataset.Namespace)
	if err != nil {
		return
	}

	stsName := j.getWorkerName()
	pods, err := j.GetRunningPodsOfStatefulSet(stsName, j.namespace)
	if err != nil || len(pods) == 0 {
		return
	}

	image := fmt.Sprintf("%s:%s", imageName, imageTag)
	dataLoadValue := j.genDataLoadValue(image, cacheinfo, pods, targetDataset, dataload)
	data, err := yaml.Marshal(dataLoadValue)
	if err != nil {
		return
	}
	j.Log.Info("dataload value", "value", string(data))

	valueFile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("%s-%s-loader-values.yaml", dataload.Namespace, dataload.Name))
	if err != nil {
		return
	}
	err = os.WriteFile(valueFile.Name(), data, 0o400)
	if err != nil {
		return
	}
	return valueFile.Name(), nil
}

func (j *JuiceFSEngine) genDataLoadValue(image string, cacheinfo map[string]string, pods []v1.Pod, targetDataset *datav1alpha1.Dataset, dataload datav1alpha1.DataLoad) *cdataload.DataLoadValue {
	imagePullSecrets := docker.GetImagePullSecretsFromEnv(common.EnvImagePullSecretsKey)

	dataloadInfo := cdataload.DataLoadInfo{
		BackoffLimit:     3,
		TargetDataset:    dataload.Spec.Dataset.Name,
		LoadMetadata:     dataload.Spec.LoadMetadata,
		Image:            image,
		Labels:           dataload.Spec.PodMetadata.Labels,
		Annotations:      dataload.Spec.PodMetadata.Annotations,
		ImagePullSecrets: imagePullSecrets,
	}

	// pod affinity
	if dataload.Spec.Affinity != nil {
		dataloadInfo.Affinity = dataload.Spec.Affinity
	}

	// node selector
	if dataload.Spec.NodeSelector != nil {
		if dataloadInfo.NodeSelector == nil {
			dataloadInfo.NodeSelector = make(map[string]string)
		}
		dataloadInfo.NodeSelector = dataload.Spec.NodeSelector
	}

	// pod tolerations
	if len(dataload.Spec.Tolerations) > 0 {
		if dataloadInfo.Tolerations == nil {
			dataloadInfo.Tolerations = make([]v1.Toleration, 0)
		}
		dataloadInfo.Tolerations = dataload.Spec.Tolerations
	}

	// scheduler name
	if len(dataload.Spec.SchedulerName) > 0 {
		dataloadInfo.SchedulerName = dataload.Spec.SchedulerName
	}

	targetPaths := []cdataload.TargetPath{}
	for _, target := range dataload.Spec.Target {
		fluidNative := utils.IsTargetPathUnderFluidNativeMounts(target.Path, *targetDataset)
		path := strings.TrimSpace(target.Path)
		targetPaths = append(targetPaths, cdataload.TargetPath{
			Path:        path,
			Replicas:    target.Replicas,
			FluidNative: fluidNative,
		})
	}
	dataloadInfo.TargetPaths = targetPaths

	options := map[string]string{}
	// resolve spec options
	if dataload.Spec.Options != nil {
		for key, value := range dataload.Spec.Options {
			options[key] = value
		}
	}

	for key, value := range cacheinfo {
		options[key] = value
	}

	podNames := []string{}
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	if cacheinfo[Edition] == CommunityEdition {
		options["podNames"] = strings.Join(podNames, ":")
	} else {
		options["podNames"] = podNames[0]
	}
	options["edition"] = cacheinfo[Edition]
	options["runtimeName"] = j.name
	if _, ok := options["timeout"]; !ok {
		options["timeout"] = DefaultDataLoadTimeout
	}

	dataloadInfo.Options = options

	dataLoadValue := &cdataload.DataLoadValue{
		DataLoadInfo: dataloadInfo,
	}

	return dataLoadValue
}

func (j *JuiceFSEngine) CheckRuntimeReady() (ready bool) {
	stsName := j.getWorkerName()
	pods, err := j.GetRunningPodsOfStatefulSet(stsName, j.namespace)
	if err != nil || len(pods) == 0 {
		return false
	}
	return true
}

func (j *JuiceFSEngine) CheckExistenceOfPath(targetDataload datav1alpha1.DataLoad) (notExist bool, err error) {
	// get mount path
	cacheinfo, err := GetCacheInfoFromConfigmap(j.Client, targetDataload.Spec.Dataset.Name, targetDataload.Spec.Dataset.Namespace)
	if err != nil {
		return
	}
	mountPath := cacheinfo[MountPath]
	if mountPath == "" {
		return true, fmt.Errorf("fail to find mountpath in dataset %s %s", targetDataload.Spec.Dataset.Name, targetDataload.Spec.Dataset.Namespace)
	}

	// get worker pod
	stsName := j.getWorkerName()
	pods, err := j.GetRunningPodsOfStatefulSet(stsName, j.namespace)
	if err != nil || len(pods) == 0 {
		return true, err
	}

	// check path exist
	pod := pods[0]
	fileUtils := operations.NewJuiceFileUtils(pod.Name, common.JuiceFSWorkerContainer, j.namespace, j.Log)
	for _, target := range targetDataload.Spec.Target {
		targetPath := filepath.Join(mountPath, target.Path)
		isExist, err := fileUtils.IsExist(targetPath)
		if err != nil {
			return true, err
		}
		if !isExist {
			return true, nil
		}
	}
	return false, err
}
