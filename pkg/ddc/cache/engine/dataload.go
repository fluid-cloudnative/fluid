/*
  Copyright 2026 The Fluid Authors.

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

package engine

import (
	"errors"
	"fmt"
	"os"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataflow"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	fluiderrors "github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/transformer"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (e *CacheEngine) generateDataLoadValueFile(ctx cruntime.ReconcileRequestContext, object client.Object) (string, error) {
	dataload, ok := object.(*v1alpha1.DataLoad)
	if !ok {
		return "", fmt.Errorf("object %v is not a DataLoad", object)
	}

	targetDataset, err := utils.GetDataset(ctx.Client, dataload.Spec.Dataset.Name, dataload.Spec.Dataset.Namespace)
	if err != nil {
		return "", err
	}

	runtime, err := e.getRuntime()
	if err != nil {
		return "", err
	}

	runtimeClass, err := e.getRuntimeClass(runtime.Spec.RuntimeClassName)
	if err != nil {
		return "", err
	}

	dataLoadValue, err := e.genDataLoadValue(ctx, targetDataset, runtime, runtimeClass, dataload)
	if err != nil {
		return "", err
	}

	data, err := yaml.Marshal(dataLoadValue)
	if err != nil {
		return "", err
	}

	valueFile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("%s-%s-loader-values.yaml", dataload.Namespace, dataload.Name))
	if err != nil {
		return "", err
	}
	err = os.WriteFile(valueFile.Name(), data, 0o400)
	if err != nil {
		return "", err
	}
	return valueFile.Name(), nil
}

func (e *CacheEngine) genDataLoadValue(ctx cruntime.ReconcileRequestContext, targetDataset *v1alpha1.Dataset, runtime *v1alpha1.CacheRuntime,
	runtimeClass *v1alpha1.CacheRuntimeClass, dataload *v1alpha1.DataLoad) (value *cdataload.DataLoadValue, err error) {

	// check runtime class defines the DataLoad or not.
	opSpec := findDataOperationSpec(runtimeClass.DataOperationSpecs, dataoperation.DataLoadType)
	if opSpec == nil {
		return nil, fluiderrors.NewNotSupported(
			schema.GroupResource{
				Group:    runtime.GetObjectKind().GroupVersionKind().Group,
				Resource: runtime.GetObjectKind().GroupVersionKind().Kind,
			}, "CacheRuntime["+e.name+"]")
	}

	// get image pull secrets from runtime class worker pod template
	var imagePullSecrets []corev1.LocalObjectReference
	if runtimeClass.Topology.Worker != nil {
		imagePullSecrets = runtimeClass.Topology.Worker.Template.Spec.ImagePullSecrets
	}

	dataloadInfo := cdataload.DataLoadInfo{
		BackoffLimit:     3,
		TargetDataset:    dataload.Spec.Dataset.Name,
		LoadMetadata:     dataload.Spec.LoadMetadata,
		Labels:           dataload.Spec.PodMetadata.Labels,
		Annotations:      dataflow.InjectAffinityAnnotation(dataload.Annotations, dataload.Spec.PodMetadata.Annotations),
		ImagePullSecrets: imagePullSecrets,
		Policy:           string(dataload.Spec.Policy),
		Schedule:         dataload.Spec.Schedule,
		Resources:        dataload.Spec.Resources,
	}

	dataloadInfo.Command = opSpec.Command
	dataloadInfo.Args = opSpec.Args
	if len(dataloadInfo.Command) == 0 && len(dataloadInfo.Args) == 0 {
		ctx.Recorder.Eventf(dataload, corev1.EventTypeWarning, common.DataOperationExecutionFailed, "dataLoad command and args defined in cache runtime class can not be both empty")
		return nil, errors.New("dataLoad command and args defined in cache runtime class can not be both empty")
	}

	// DataOperationSpecs image takes precedence; falls back to worker image if empty.
	dataloadInfo.Image = opSpec.Image
	if len(dataloadInfo.Image) == 0 {
		dataloadInfo.Image, err = e.getDataOperationImage(runtime, runtimeClass)
		if err != nil {
			return nil, err
		}
	}

	// pod affinity
	if dataload.Spec.Affinity != nil {
		dataloadInfo.Affinity = dataload.Spec.Affinity
	}

	// inject the node affinity by previous operation pod.
	dataloadInfo.Affinity, err = dataflow.InjectAffinityByRunAfterOp(e.Client, dataload.Spec.RunAfter, dataload.Namespace, dataloadInfo.Affinity)
	if err != nil {
		return nil, err
	}

	// node selector
	if dataload.Spec.NodeSelector != nil {
		dataloadInfo.NodeSelector = dataload.Spec.NodeSelector
	}

	// pod tolerations
	if len(dataload.Spec.Tolerations) > 0 {
		dataloadInfo.Tolerations = dataload.Spec.Tolerations
	}

	// scheduler name
	if len(dataload.Spec.SchedulerName) > 0 {
		dataloadInfo.SchedulerName = dataload.Spec.SchedulerName
	}

	var targetPaths []cdataload.TargetPath
	for _, target := range dataload.Spec.Target {
		targetPaths = append(targetPaths, cdataload.TargetPath{
			Path:     target.Path,
			Replicas: target.Replicas,
			// currently we don't support the FluidNative field.
		})
	}
	dataloadInfo.TargetPaths = targetPaths

	// injected envs
	dataloadInfo.Envs = []cdataload.Env{
		{
			Name:  "FLUID_RUNTIME_CONFIG_PATH",
			Value: e.getRuntimeConfigPath(),
		},
		// FLUID_DATALOAD_DATA_PATH and FLUID_DATALOAD_PATH_REPLICAS is generated and set in the helm job yaml.
	}

	dataLoadValue := &cdataload.DataLoadValue{
		Name:           dataload.Name,
		OwnerDatasetId: utils.GetDatasetId(targetDataset.Namespace, targetDataset.Name, string(targetDataset.UID)),
		DataLoadInfo:   dataloadInfo,
		Owner:          transformer.GenerateOwnerReferenceFromObject(dataload),
	}

	return dataLoadValue, nil
}
