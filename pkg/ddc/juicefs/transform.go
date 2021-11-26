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

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

func (j *JuiceFSEngine) transform(runtime *datav1alpha1.JuiceFSRuntime) (value *JuiceFS, err error) {
	if runtime == nil {
		err = fmt.Errorf("the juicefsRuntime is null")
		return
	}

	dataset, err := utils.GetDataset(j.Client, j.name, j.namespace)
	if err != nil {
		return value, err
	}

	value = &JuiceFS{}

	value.FullnameOverride = j.name

	// transform the workers
	err = j.transformWorkers(runtime, value)
	if err != nil {
		return
	}

	// transform the fuse
	err = j.transformFuse(runtime, dataset, value)
	if err != nil {
		return
	}

	// set the placementMode
	j.transformPlacementMode(dataset, value)
	return
}

func (j *JuiceFSEngine) transformWorkers(runtime *datav1alpha1.JuiceFSRuntime, value *JuiceFS) (err error) {
	value.Worker = Worker{}

	image := runtime.Spec.JuiceFSVersion.Image
	imageTag := runtime.Spec.JuiceFSVersion.ImageTag
	imagePullPolicy := runtime.Spec.JuiceFSVersion.ImagePullPolicy

	value.Worker.Envs = runtime.Spec.Worker.Env
	value.Worker.Ports = runtime.Spec.Worker.Ports

	value.Image, value.ImageTag, value.ImagePullPolicy = j.parseRuntimeImage(image, imageTag, imagePullPolicy)

	if len(value.Worker.NodeSelector) == 0 {
		value.Worker.NodeSelector = map[string]string{}
	}

	if len(runtime.Spec.TieredStore.Levels) > 0 {
		if runtime.Spec.TieredStore.Levels[0].MediumType != common.Memory {
			value.Worker.CacheDir = runtime.Spec.TieredStore.Levels[0].Path
		}
	} else {
		value.Worker.CacheDir = DefaultCacheDir
	}

	j.transformResourcesForWorker(runtime, value)
	return
}

func (j *JuiceFSEngine) transformPlacementMode(dataset *datav1alpha1.Dataset, value *JuiceFS) {
	value.PlacementMode = string(dataset.Spec.PlacementMode)
	if len(value.PlacementMode) == 0 {
		value.PlacementMode = string(datav1alpha1.ExclusiveMode)
	}
}
