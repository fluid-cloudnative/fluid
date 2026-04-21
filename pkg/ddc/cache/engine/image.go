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
	"fmt"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

// getDataOperationImage get data operation image for cache runtime by using worker image.
func (e *CacheEngine) getDataOperationImage(runtime *v1alpha1.CacheRuntime, runtimeClass *v1alpha1.CacheRuntimeClass) (image string, err error) {
	// Priority: DataOperationSpecs.image > runtime.Spec.Worker.RuntimeVersion > runtimeClass.Topology.Worker

	if runtimeClass.Topology.Worker != nil && len(runtimeClass.Topology.Worker.Template.Spec.Containers) > 0 {
		// container [0] is the cache engine image
		image = runtimeClass.Topology.Worker.Template.Spec.Containers[0].Image
	}
	if len(runtime.Spec.Worker.RuntimeVersion.Image) > 0 && len(runtime.Spec.Worker.RuntimeVersion.ImageTag) > 0 {
		image = runtime.Spec.Worker.RuntimeVersion.Image + ":" + runtime.Spec.Worker.RuntimeVersion.ImageTag
	}

	if len(image) == 0 {
		return "", fmt.Errorf("no image for runtime, name: %s, namespace: %s", runtime.Name, runtime.Namespace)
	}

	return image, nil
}
