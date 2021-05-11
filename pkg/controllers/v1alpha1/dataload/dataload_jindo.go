package dataload

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

// generateDataLoadValueFile builds a DataLoadValue by extracted specifications from the given DataLoad, and
// marshals the DataLoadValue to a temporary yaml file where stores values that'll be used by fluid dataloader helm chart
func (r *DataLoadReconcilerImplement) generateJindoDataLoadValueFile(dataload v1alpha1.DataLoad) (valueFileName string, err error) {
	targetDataset, err := utils.GetDataset(r.Client, dataload.Spec.Dataset.Name, dataload.Spec.Dataset.Namespace)
	if err != nil {
		return "", err
	}

	_, boundedRuntime := utils.GetRuntimeByCategory(targetDataset.Status.Runtimes, common.AccelerateCategory)

	imageName, imageTag := docker.GetWorkerImage(r.Client, dataload.Spec.Dataset.Name, boundedRuntime.Type, dataload.Spec.Dataset.Namespace)
	image := fmt.Sprintf("%s:%s", imageName, imageTag)

	runtime, err := utils.GetJindoRuntime(r.Client, boundedRuntime.Name, boundedRuntime.Namespace)
	if err != nil {
		return
	}
	hadoopConfig := runtime.Spec.HadoopConfig
	loadMemoryData := false
	if len(runtime.Spec.Tieredstore.Levels) == 0 {
		err = fmt.Errorf("the Tieredstore is null")
		return
	}
	if runtime.Spec.Tieredstore.Levels[0].MediumType == "MEM" {
		loadMemoryData = true
	}

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
	options := map[string]string{}
	if loadMemoryData {
		options["loadMemoryData"] = "true"
	} else {
		options["loadMemoryData"] = "false"
	}
	if hadoopConfig != "" {
		options["hdfsConfig"] = hadoopConfig
	}
	dataloadInfo.Options = options

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
