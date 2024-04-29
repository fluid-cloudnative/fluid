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

package jindo

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/utils/transfromers"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindo/operations"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	jindoutils "github.com/fluid-cloudnative/fluid/pkg/utils/jindo"
)

// CreateDataLoadJob creates the job to load data
func (e *JindoEngine) CreateDataLoadJob(ctx cruntime.ReconcileRequestContext, targetDataload datav1alpha1.DataLoad) (err error) {
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
		valueFileName, err := e.generateDataLoadValueFile(ctx, &targetDataload)
		if err != nil {
			log.Error(err, "failed to generate dataload chart's value file")
			return err
		}
		chartName := utils.GetChartsDirectory() + "/" + cdataload.DataloadChart + "/" + common.JindoRuntime
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
func (e *JindoEngine) generateDataLoadValueFile(r cruntime.ReconcileRequestContext, object client.Object) (valueFileName string, err error) {
	dataload, ok := object.(*datav1alpha1.DataLoad)
	if !ok {
		err = fmt.Errorf("object %v is not a DataLoad", object)
		return "", err
	}

	targetDataset, err := utils.GetDataset(r.Client, dataload.Spec.Dataset.Name, dataload.Spec.Dataset.Namespace)
	if err != nil {
		return "", err
	}

	imageName, imageTag := docker.GetWorkerImage(r.Client, dataload.Spec.Dataset.Name, "jindo", dataload.Spec.Dataset.Namespace)

	if len(imageName) == 0 {
		defaultImageInfo := strings.Split(jindoutils.GetRuntimeImage(), ":")
		if len(defaultImageInfo) < 1 {
			panic("invalid default dataload image!")
		} else {
			imageName = defaultImageInfo[0]
		}
	}

	if len(imageTag) == 0 {
		defaultImageInfo := strings.Split(jindoutils.GetRuntimeImage(), ":")
		if len(defaultImageInfo) < 2 {
			panic("invalid default dataload image!")
		} else {
			imageTag = defaultImageInfo[1]
		}
	}

	image := fmt.Sprintf("%s:%s", imageName, imageTag)

	runtime, err := utils.GetJindoRuntime(r.Client, dataload.Spec.Dataset.Name, dataload.Spec.Dataset.Namespace)
	if err != nil {
		return
	}

	if len(runtime.Spec.TieredStore.Levels) == 0 {
		err = fmt.Errorf("the TieredStore is null")
		return
	}

	dataLoadValue := e.genDataLoadValue(image, runtime, targetDataset, dataload)
	data, err := yaml.Marshal(dataLoadValue)
	if err != nil {
		return
	}

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

func (e *JindoEngine) genDataLoadValue(image string, runtime *datav1alpha1.JindoRuntime, targetDataset *datav1alpha1.Dataset, dataload *datav1alpha1.DataLoad) *cdataload.DataLoadValue {
	imagePullSecrets := docker.GetImagePullSecretsFromEnv(common.EnvImagePullSecretsKey)

	hadoopConfig := runtime.Spec.HadoopConfig
	loadMemoryData := false

	if len(runtime.Spec.TieredStore.Levels) > 0 && runtime.Spec.TieredStore.Levels[0].MediumType == "MEM" {
		loadMemoryData = true
	}

	dataloadInfo := cdataload.DataLoadInfo{
		BackoffLimit:     3,
		TargetDataset:    dataload.Spec.Dataset.Name,
		LoadMetadata:     dataload.Spec.LoadMetadata,
		Image:            image,
		ImagePullSecrets: imagePullSecrets,
		Policy:           string(dataload.Spec.Policy),
		Schedule:         dataload.Spec.Schedule,
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
		targetPaths = append(targetPaths, cdataload.TargetPath{
			Path:        target.Path,
			Replicas:    target.Replicas,
			FluidNative: fluidNative,
		})
	}
	dataloadInfo.TargetPaths = targetPaths
	options := map[string]string{}
	if loadMemoryData {
		options["loadMemorydata"] = "true"
	} else {
		options["loadMemorydata"] = "false"
	}
	if hadoopConfig != "" {
		options["hdfsConfig"] = hadoopConfig
	}
	// resolve spec options
	if dataload.Spec.Options != nil {
		for key, value := range dataload.Spec.Options {
			options[key] = value
		}
	}
	dataloadInfo.Options = options

	dataLoadValue := &cdataload.DataLoadValue{
		Name:         dataload.Name,
		DataLoadInfo: dataloadInfo,
		Owner:        transfromer.GenerateOwnerReferenceFromObject(dataload),
	}
	return dataLoadValue
}

func (e *JindoEngine) CheckRuntimeReady() (ready bool) {
	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewJindoFileUtils(podName, containerName, e.namespace, e.Log)
	ready = fileUtils.Ready()
	if !ready {
		e.Log.Info("runtime not ready", "runtime", ready)
		return false
	}
	return true
}
