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

package juicefs

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/juicefs/operations"
	"io/ioutil"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	"os"
	"path/filepath"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
)

type JuiceDataLoadValue struct {
	DataLoadInfo cdataload.DataLoadInfo `yaml:"dataloader"`
	PodNames     []string               `yaml:"podNames"`
	RuntimeName  string                 `yaml:"runtimeName"`
}

// CreateDataLoadJob creates the job to load data
func (e *JuiceFSEngine) CreateDataLoadJob(ctx cruntime.ReconcileRequestContext, targetDataload datav1alpha1.DataLoad) (err error) {
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
		valueFileName, err := e.generateDataLoadValueFile(ctx, targetDataload)
		if err != nil {
			log.Error(err, "failed to generate dataload chart's value file")
			return err
		}
		chartName := utils.GetChartsDirectory() + "/" + cdataload.DATALOAD_CHART + "/" + common.JuiceFSRuntime
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
func (e *JuiceFSEngine) generateDataLoadValueFile(r cruntime.ReconcileRequestContext, dataload datav1alpha1.DataLoad) (valueFileName string, err error) {
	targetDataset, err := utils.GetDataset(r.Client, dataload.Spec.Dataset.Name, dataload.Spec.Dataset.Namespace)
	if err != nil {
		return "", err
	}
	e.Log.Info("target dataset", "dataset", targetDataset)

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

	image := fmt.Sprintf("%s:%s", imageName, imageTag)

	dataloadInfo := cdataload.DataLoadInfo{
		BackoffLimit:  3,
		TargetDataset: dataload.Spec.Dataset.Name,
		LoadMetadata:  dataload.Spec.LoadMetadata,
		Image:         image,
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
	// resolve spec options
	if dataload.Spec.Options != nil {
		for key, value := range dataload.Spec.Options {
			options[key] = value
		}
	}
	cacheinfo, err := GetCacheInfoFromConfigmap(e.Client, dataload.Spec.Dataset.Name, dataload.Spec.Dataset.Namespace)
	if err != nil {
		return
	}
	for key, value := range cacheinfo {
		options[key] = value
	}

	dataloadInfo.Options = options

	dsName := e.getFuseDaemonsetName()
	pods, err := e.GetRunningPodsOfDaemonset(dsName, e.namespace)
	if err != nil || len(pods) == 0 {
		return
	}
	podNames := []string{}
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}

	dataLoadValue := JuiceDataLoadValue{
		DataLoadInfo: dataloadInfo,
		RuntimeName:  e.name,
		PodNames:     podNames,
	}

	data, err := yaml.Marshal(dataLoadValue)
	if err != nil {
		return
	}
	e.Log.Info("dataload value", "value", string(data))

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

func (e *JuiceFSEngine) CheckRuntimeReady() (ready bool) {
	dsName := e.getFuseDaemonsetName()
	pods, err := e.GetRunningPodsOfDaemonset(dsName, e.namespace)
	if err != nil || len(pods) == 0 {
		return false
	}
	for _, pod := range pods {
		if !podutil.IsPodReady(&pod) {
			return false
		}
	}
	return true
}

func (e *JuiceFSEngine) CheckExistenceOfPath(targetDataload datav1alpha1.DataLoad) (notExist bool, err error) {
	// get mount path
	cacheinfo, err := GetCacheInfoFromConfigmap(e.Client, targetDataload.Spec.Dataset.Name, targetDataload.Spec.Dataset.Namespace)
	if err != nil {
		return
	}
	mountPath := cacheinfo[MOUNTPATH]
	if mountPath == "" {
		return true, fmt.Errorf("fail to find mountpath in dataset %s %s", targetDataload.Spec.Dataset.Name, targetDataload.Spec.Dataset.Namespace)
	}

	// get fuse pod
	dsName := e.getFuseDaemonsetName()
	pods, err := e.GetRunningPodsOfDaemonset(dsName, e.namespace)
	if err != nil || len(pods) == 0 {
		return true, err
	}

	// check path exist
	for _, pod := range pods {
		fileUtils := operations.NewJuiceFileUtils(pod.Name, common.JuiceFSFuseContainer, e.namespace, e.Log)
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
	}
	return false, nil
}
