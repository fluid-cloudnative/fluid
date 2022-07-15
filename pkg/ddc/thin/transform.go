/*
  Copyright 2022 The Fluid Authors.

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

package thin

import (
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/transfromer"
)

func (t *ThinEngine) transform(runtime *datav1alpha1.ThinRuntime, profile *datav1alpha1.ThinProfile) (value *ThinValue, err error) {
	if runtime == nil {
		err = fmt.Errorf("the juicefsRuntime is null")
		return
	}

	dataset, err := utils.GetDataset(t.Client, t.name, t.namespace)
	if err != nil {
		return value, err
	}

	value = &ThinValue{}

	value.FullnameOverride = t.name
	value.Owner = transfromer.GenerateOwnerReferenceFromObject(runtime)

	// transform the workers
	err = t.transformWorkers(runtime, profile, value)
	if err != nil {
		return
	}

	// transform the fuse
	err = t.transformFuse(runtime, profile, dataset, value)
	if err != nil {
		return
	}

	// set the placementMode
	t.transformPlacementMode(dataset, value)
	return
}

func (t *ThinEngine) transformWorkers(runtime *datav1alpha1.ThinRuntime, profile *datav1alpha1.ThinProfile, value *ThinValue) (err error) {
	value.Worker = Worker{}

	image := runtime.Spec.Version.Image
	imageTag := runtime.Spec.Version.ImageTag
	imagePullPolicy := runtime.Spec.Version.ImagePullPolicy

	value.Worker.Envs = runtime.Spec.Worker.Env
	value.Worker.Ports = runtime.Spec.Worker.Ports

	// todo
	value.Image, value.ImageTag, value.ImagePullPolicy = image, imageTag, imagePullPolicy

	if len(value.Worker.NodeSelector) == 0 {
		value.Worker.NodeSelector = map[string]string{}
	}

	if len(runtime.Spec.TieredStore.Levels) > 0 {
		if runtime.Spec.TieredStore.Levels[0].MediumType != common.Memory {
			value.Worker.CacheDir = runtime.Spec.TieredStore.Levels[0].Path
		}
	}

	t.transformResourcesForWorker(runtime, value)
	return
}

func (t *ThinEngine) transformPlacementMode(dataset *datav1alpha1.Dataset, value *ThinValue) {
	value.PlacementMode = string(dataset.Spec.PlacementMode)
	if len(value.PlacementMode) == 0 {
		value.PlacementMode = string(datav1alpha1.ExclusiveMode)
	}
}
