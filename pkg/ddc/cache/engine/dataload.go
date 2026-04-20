package engine

import (
	"fmt"
	"os"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/dataflow"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	fluiderrors "github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/transformer"
	"sigs.k8s.io/yaml"

	"k8s.io/apimachinery/pkg/runtime/schema"
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

	// check runtime class defines the DataLoad or not.
	var supportDataLoad = false
	for _, op := range runtimeClass.DataOperationSpecs {
		if op.Name == string(dataoperation.DataLoadType) {
			supportDataLoad = true
			break
		}
	}
	if !supportDataLoad {
		return "", fluiderrors.NewNotSupported(
			schema.GroupResource{
				Group:    object.GetObjectKind().GroupVersionKind().Group,
				Resource: object.GetObjectKind().GroupVersionKind().Kind,
			}, "JuiceFSRuntime")
	}

	dataLoadValue, err := e.genDataLoadValue(targetDataset, runtime, runtimeClass, dataload)
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

func (e *CacheEngine) genDataLoadValue(targetDataset *v1alpha1.Dataset, runtime *v1alpha1.CacheRuntime,
	runtimeClass *v1alpha1.CacheRuntimeClass, dataload *v1alpha1.DataLoad) (value *cdataload.DataLoadValue, err error) {

	// get image pull secrets from runtime class worker pod template
	imagePullSecrets := runtimeClass.Topology.Worker.Template.Spec.ImagePullSecrets

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

	for _, op := range runtimeClass.DataOperationSpecs {
		if op.Name == string(dataoperation.DataLoadType) {
			dataloadInfo.Command = op.Command
			dataloadInfo.Args = op.Args
			// runtime class operation image has higher priority
			dataloadInfo.Image = op.Image
			if len(dataloadInfo.Image) == 0 {
				dataloadInfo.Image, err = e.getDataOperationImage(runtime, runtimeClass)
				if err != nil {
					return nil, err
				}
			}
			break
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
