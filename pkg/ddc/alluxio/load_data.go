/*
Copyright 2021 The Fluid Authors.

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
	"os"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/dataflow"
	"github.com/fluid-cloudnative/fluid/pkg/utils/transformer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
)

// generateDataLoadValueFile builds a DataLoadValue by extracted specifications from the given DataLoad, and
// marshals the DataLoadValue to a temporary yaml file where stores values that'll be used by fluid dataloader helm chart
func (e *AlluxioEngine) generateDataLoadValueFile(r cruntime.ReconcileRequestContext, object client.Object) (valueFileName string, err error) {
	dataload, ok := object.(*datav1alpha1.DataLoad)
	if !ok {
		err = fmt.Errorf("object %v is not a DataLoad", object)
		return "", err
	}

	targetDataset, err := utils.GetDataset(r.Client, dataload.Spec.Dataset.Name, dataload.Spec.Dataset.Namespace)
	if err != nil {
		return "", err
	}

	imageName, imageTag := docker.GetWorkerImage(r.Client, dataload.Spec.Dataset.Name, "alluxio", dataload.Spec.Dataset.Namespace)

	if len(imageName) == 0 {
		imageName = docker.GetImageRepoFromEnv(common.AlluxioRuntimeImageEnv)
		if len(imageName) == 0 {
			defaultImageInfo := strings.Split(common.DefaultAlluxioRuntimeImage, ":")
			if len(defaultImageInfo) < 1 {
				panic("invalid default dataload image!")
			} else {
				imageName = defaultImageInfo[0]
			}
		}
	}

	if len(imageTag) == 0 {
		imageTag = docker.GetImageTagFromEnv(common.AlluxioRuntimeImageEnv)
		if len(imageTag) == 0 {
			defaultImageInfo := strings.Split(common.DefaultAlluxioRuntimeImage, ":")
			if len(defaultImageInfo) < 2 {
				panic("invalid default dataload image!")
			} else {
				imageTag = defaultImageInfo[1]
			}
		}
	}

	image := fmt.Sprintf("%s:%s", imageName, imageTag)

	dataLoadValue, err := e.genDataLoadValue(image, targetDataset, dataload)
	if err != nil {
		return
	}

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

// genDataLoadValue generates configuration values for a DataLoad operation
// Parameters:
//   - image: Container image to use for the data load operation
//   - targetDataset: Target dataset object where data will be loaded
//   - dataload: DataLoad custom resource specification
// Returns:
//   - *cdataload.DataLoadValue: Fully populated configuration object for data loading
//   - error: Any error that occurs during processing
func (e *AlluxioEngine) genDataLoadValue(image string, targetDataset *datav1alpha1.Dataset, dataload *datav1alpha1.DataLoad) (*cdataload.DataLoadValue, error) {
    // Retrieve image pull secrets from environment variables
    // Returns empty slice if environment variable isn't set
    imagePullSecrets := docker.GetImagePullSecretsFromEnv(common.EnvImagePullSecretsKey)

    // Build base DataLoad configuration structure
    dataloadInfo := cdataload.DataLoadInfo{
        BackoffLimit:     3,  // Number of retries for failed jobs
        TargetDataset:    dataload.Spec.Dataset.Name,  // Name of target dataset
        LoadMetadata:     dataload.Spec.LoadMetadata,  // Whether to load metadata
        Image:            image,  // Container image for data loader
        Labels:           dataload.Spec.PodMetadata.Labels,  // Pod labels
        // Merge annotations with affinity injection
        Annotations:      dataflow.InjectAffinityAnnotation(dataload.Annotations, dataload.Spec.PodMetadata.Annotations),
        ImagePullSecrets: imagePullSecrets,  // Credentials for pulling images
        Policy:           string(dataload.Spec.Policy),  // Execution policy (e.g., Once, Cron)
        Schedule:         dataload.Spec.Schedule,  // Cron schedule for periodic execution
        Resources:        dataload.Spec.Resources,  // CPU/memory requirements
    }

    // Apply affinity settings if specified in DataLoad CR
    if dataload.Spec.Affinity != nil {
        dataloadInfo.Affinity = dataload.Spec.Affinity
    }

    // Inject dependency-based affinity constraints
    // Ensures this operation runs after specified predecessor operations
    var err error
    dataloadInfo.Affinity, err = dataflow.InjectAffinityByRunAfterOp(
        e.Client,                // Kubernetes API client
        dataload.Spec.RunAfter,   // Operations that must complete first
        dataload.Namespace,       // Kubernetes namespace
        dataloadInfo.Affinity,    // Current affinity settings
    )
    if err != nil {
        return nil, err  // Propagate affinity injection errors
    }

    // Configure node selection constraints
    if dataload.Spec.NodeSelector != nil {
        // Initialize map if empty
        if dataloadInfo.NodeSelector == nil {
            dataloadInfo.NodeSelector = make(map[string]string)
        }
        dataloadInfo.NodeSelector = dataload.Spec.NodeSelector
    }

    // Configure tolerations for tainted nodes
    if len(dataload.Spec.Tolerations) > 0 {
        // Initialize slice if empty
        if dataloadInfo.Tolerations == nil {
            dataloadInfo.Tolerations = make([]v1.Toleration, 0)
        }
        dataloadInfo.Tolerations = dataload.Spec.Tolerations
    }

    // Assign custom scheduler if specified
    if len(dataload.Spec.SchedulerName) > 0 {
        dataloadInfo.SchedulerName = dataload.Spec.SchedulerName
    }

    // Process target data paths
    targetPaths := []cdataload.TargetPath{}
    for _, target := range dataload.Spec.Target {
        // Check if path is within Fluid-native mounts
        fluidNative := utils.IsTargetPathUnderFluidNativeMounts(target.Path, *targetDataset)
        
        // Append path configuration
        targetPaths = append(targetPaths, cdataload.TargetPath{
            Path:        target.Path,       // Absolute data path to load
            Replicas:    target.Replicas,   // Number of data replicas to create
            FluidNative: fluidNative,       // Whether path is Fluid-native
        })
    }
    dataloadInfo.TargetPaths = targetPaths

    // Assemble final configuration object
    dataLoadValue := &cdataload.DataLoadValue{
        Name:           dataload.Name,  // DataLoad job name
        OwnerDatasetId: utils.GetDatasetId(  // Unique dataset identifier
            targetDataset.Namespace, 
            targetDataset.Name, 
            string(targetDataset.UID),
        DataLoadInfo:   dataloadInfo,  // Core configuration
        // Owner reference for garbage collection
        Owner:          transformer.GenerateOwnerReferenceFromObject(dataload),
    }

    return dataLoadValue, nil
}

// CheckRuntimeReady checks if the Alluxio runtime is operational.
// It obtains master pod details, creates file utilities, and checks readiness.
//
// Returns:
//
//	ready bool - Runtime readiness status (true = ready, false = not ready).
func (e *AlluxioEngine) CheckRuntimeReady() (ready bool) {
	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)
	ready = fileUtils.Ready()
	if !ready {
		e.Log.Info("runtime not ready", "runtime", ready)
		return false
	}
	return true
}
