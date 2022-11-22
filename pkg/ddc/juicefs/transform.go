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

package juicefs

import (
	"fmt"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/transfromer"
	corev1 "k8s.io/api/core/v1"
)

func (j *JuiceFSEngine) transform(runtime *datav1alpha1.JuiceFSRuntime) (value *JuiceFS, err error) {
	defer utils.TimeTrack(time.Now(), "JuiceFSRuntime.Transform", "name", runtime.Name)
	if runtime == nil {
		err = fmt.Errorf("the juicefsRuntime is null")
		return
	}

	dataset, err := utils.GetDataset(j.Client, j.name, j.namespace)
	if err != nil {
		return value, err
	}

	value = &JuiceFS{
		RuntimeIdentity: common.RuntimeIdentity{
			Namespace: runtime.Namespace,
			Name:      runtime.Name,
		},
	}

	value.FullnameOverride = j.name
	value.Owner = transfromer.GenerateOwnerReferenceFromObject(runtime)

	// transform toleration
	j.transformTolerations(dataset, value)

	// transform the fuse
	value.Fuse = Fuse{
		Privileged: true,
	}
	value.Worker = Worker{
		Privileged: true,
	}
	err = j.transformFuse(runtime, dataset, value)
	if err != nil {
		return
	}

	// transform the workers
	err = j.transformWorkers(runtime, value)
	if err != nil {
		return
	}

	// transform runtime pod metadata
	err = j.transformPodMetadata(runtime, value)
	if err != nil {
		return
	}

	// set the placementMode
	j.transformPlacementMode(dataset, value)
	return
}

func (j *JuiceFSEngine) transformWorkers(runtime *datav1alpha1.JuiceFSRuntime, value *JuiceFS) (err error) {

	image := runtime.Spec.JuiceFSVersion.Image
	imageTag := runtime.Spec.JuiceFSVersion.ImageTag
	imagePullPolicy := runtime.Spec.JuiceFSVersion.ImagePullPolicy

	value.Worker.Envs = runtime.Spec.Worker.Env
	value.Worker.Ports = runtime.Spec.Worker.Ports

	value.Image, value.ImageTag, value.ImagePullPolicy = j.parseRuntimeImage(image, imageTag, imagePullPolicy)

	// nodeSelector
	value.Worker.NodeSelector = map[string]string{}
	if len(runtime.Spec.Worker.NodeSelector) > 0 {
		value.Worker.NodeSelector = runtime.Spec.Worker.NodeSelector
	}

	err = j.transformResourcesForWorker(runtime, value)
	if err != nil {
		j.Log.Error(err, "failed to transform resource for worker")
		return
	}

	// transform volumes for worker
	err = j.transformWorkerVolumes(runtime, value)
	if err != nil {
		j.Log.Error(err, "failed to transform volumes for worker")
	}

	// parse work pod network mode
	value.Worker.HostNetwork = datav1alpha1.IsHostNetwork(runtime.Spec.Worker.NetworkMode)
	return
}

func (j *JuiceFSEngine) transformPlacementMode(dataset *datav1alpha1.Dataset, value *JuiceFS) {
	value.PlacementMode = string(dataset.Spec.PlacementMode)
	if len(value.PlacementMode) == 0 {
		value.PlacementMode = string(datav1alpha1.ExclusiveMode)
	}
}

func (j *JuiceFSEngine) transformTolerations(dataset *datav1alpha1.Dataset, value *JuiceFS) {
	if len(dataset.Spec.Tolerations) > 0 {
		// value.Tolerations = dataset.Spec.Tolerations
		value.Tolerations = []corev1.Toleration{}
		for _, toleration := range dataset.Spec.Tolerations {
			toleration.TolerationSeconds = nil
			value.Tolerations = append(value.Tolerations, toleration)
		}
	}
}

func (j *JuiceFSEngine) transformPodMetadata(runtime *datav1alpha1.JuiceFSRuntime, value *JuiceFS) (err error) {
	commonLabels := utils.UnionMapsWithOverride(map[string]string{}, runtime.Spec.PodMetadata.Labels)
	value.Worker.Labels = utils.UnionMapsWithOverride(commonLabels, runtime.Spec.Worker.PodMetadata.Labels)
	value.Fuse.Labels = utils.UnionMapsWithOverride(commonLabels, runtime.Spec.Fuse.PodMetadata.Labels)

	commonAnnotations := utils.UnionMapsWithOverride(map[string]string{}, runtime.Spec.PodMetadata.Annotations)
	value.Worker.Annotations = utils.UnionMapsWithOverride(commonAnnotations, runtime.Spec.Worker.PodMetadata.Annotations)
	value.Fuse.Annotations = utils.UnionMapsWithOverride(commonAnnotations, runtime.Spec.Fuse.PodMetadata.Annotations)

	return nil
}
