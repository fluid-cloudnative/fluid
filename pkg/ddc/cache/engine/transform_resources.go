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

package engine

import (
	corev1 "k8s.io/api/core/v1"
)

func (t *CacheEngine) transformResourcesForContainer(resources corev1.ResourceRequirements, container *corev1.Container) {
	if container.Resources.Requests == nil {
		container.Resources.Requests = corev1.ResourceList{}
	}

	if container.Resources.Limits == nil {
		container.Resources.Limits = corev1.ResourceList{}
	}

	if resources.Limits != nil {
		if quantity, ok := resources.Limits[corev1.ResourceCPU]; ok {
			container.Resources.Limits[corev1.ResourceCPU] = quantity
		}
		if quantity, ok := resources.Limits[corev1.ResourceMemory]; ok {
			container.Resources.Limits[corev1.ResourceMemory] = quantity
		}
	}

	if resources.Requests != nil {
		if quantity, ok := resources.Requests[corev1.ResourceCPU]; ok {
			container.Resources.Requests[corev1.ResourceCPU] = quantity
		}
		if quantity, ok := resources.Requests[corev1.ResourceMemory]; ok {
			container.Resources.Requests[corev1.ResourceMemory] = quantity
		}
	}
}
