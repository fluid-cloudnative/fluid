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
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/dataflow"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluid-cloudnative/fluid/pkg/utils/transformer"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
)

// generateDataLoadValueFile builds a DataLoadValue by extracted specifications from the given DataLoad, and
// marshals the DataLoadValue to a temporary yaml file where stores values that'll be used by fluid dataloader helm chart
func (j *JuiceFSEngine) generateDataLoadValueFile(r cruntime.ReconcileRequestContext, object client.Object) (valueFileName string, err error) {
	dataload, ok := object.(*datav1alpha1.DataLoad)
	if !ok {
		err = fmt.Errorf("object %v is not a DataLoad", object)
		return "", err
	}

	targetDataset, err := utils.GetDataset(r.Client, dataload.Spec.Dataset.Name, dataload.Spec.Dataset.Namespace)
	if err != nil {
		return "", err
	}
	j.Log.Info("target dataset", "dataset", targetDataset)

	imageName, imageTag := docker.GetWorkerImage(r.Client, dataload.Spec.Dataset.Name, "juicefs", dataload.Spec.Dataset.Namespace)

	var defaultJuiceFSImage string
	if len(imageName) == 0 || len(imageTag) == 0 {
		defaultJuiceFSImage = common.DefaultCEImage
		edition := j.GetEdition()
		if edition == EnterpriseEdition {
			defaultJuiceFSImage = common.DefaultEEImage
		}
	}

	if len(imageName) == 0 {
		defaultImageInfo := strings.Split(defaultJuiceFSImage, ":")
		if len(defaultImageInfo) < 1 {
			err = fmt.Errorf("invalid default dataload image")
			return
		} else {
			imageName = defaultImageInfo[0]
		}
	}

	if len(imageTag) == 0 {
		defaultImageInfo := strings.Split(defaultJuiceFSImage, ":")
		if len(defaultImageInfo) < 2 {
			err = fmt.Errorf("invalid default dataload image")
			return
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
	dataLoadValue, err := j.genDataLoadValue(image, cacheinfo, pods, targetDataset, dataload)
	if err != nil {
		return
	}
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

func (j *JuiceFSEngine) genDataLoadValue(image string, cacheinfo map[string]string, pods []v1.Pod, targetDataset *datav1alpha1.Dataset, dataload *datav1alpha1.DataLoad) (*cdataload.DataLoadValue, error) {
	imagePullSecrets := docker.GetImagePullSecretsFromEnv(common.EnvImagePullSecretsKey)

	dataloadInfo := cdataload.DataLoadInfo{
		BackoffLimit:     3,
		TargetDataset:    dataload.Spec.Dataset.Name,
		LoadMetadata:     dataload.Spec.LoadMetadata,
		Image:            image,
		Labels:           dataload.Spec.PodMetadata.Labels,
		Annotations:      dataflow.InjectAffinityAnnotation(dataload.Annotations, dataload.Spec.PodMetadata.Annotations),
		ImagePullSecrets: imagePullSecrets,
		Policy:           string(dataload.Spec.Policy),
		Schedule:         dataload.Spec.Schedule,
		Resources:        dataload.Spec.Resources,
	}

	// pod affinity
	if dataload.Spec.Affinity != nil {
		dataloadInfo.Affinity = dataload.Spec.Affinity
	}

	// generate the node affinity by previous operation pod.
	var err error
	dataloadInfo.Affinity, err = dataflow.InjectAffinityByRunAfterOp(j.Client, dataload.Spec.RunAfter, dataload.Namespace, dataloadInfo.Affinity)
	if err != nil {
		return nil, err
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
	timeout := dataload.Spec.Options["timeout"]
	delete(dataload.Spec.Options, "timeout")
	if timeout == "" {
		timeout = DefaultDataLoadTimeout
	}
	warmupOtions := []string{}
	for k, v := range dataload.Spec.Options {
		if v != "" {
			warmupOtions = append(warmupOtions, fmt.Sprintf("--%s=%s", k, v))
		} else {
			warmupOtions = append(warmupOtions, fmt.Sprintf("--%s", k))
		}
	}
	options["option"] = strings.Join(warmupOtions, " ")
	options["timeout"] = timeout

	dataloadInfo.Options = options

	dataLoadValue := &cdataload.DataLoadValue{
		Name:           dataload.Name,
		OwnerDatasetId: utils.GetDatasetId(targetDataset.Namespace, targetDataset.Name, string(targetDataset.UID)),
		DataLoadInfo:   dataloadInfo,
		Owner:          transformer.GenerateOwnerReferenceFromObject(dataload),
	}

	return dataLoadValue, nil
}

func (j *JuiceFSEngine) CheckRuntimeReady() (ready bool) {
	stsName := j.getWorkerName()
	pods, err := j.GetRunningPodsOfStatefulSet(stsName, j.namespace)
	if err != nil || len(pods) == 0 {
		return false
	}
	return true
}
