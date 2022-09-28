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

package fuse

import (
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

func mutateVolumeMounts(containers []corev1.Container, datasetVolumeNames []string) (retContainers []corev1.Container, needInjection bool) {
	mountPropagationHostToContainer := corev1.MountPropagationHostToContainer

	for _, container := range containers {

		// Set HostToContainer to the dataset volume mount point
		for i, volumeMount := range container.VolumeMounts {
			if utils.ContainsString(datasetVolumeNames, volumeMount.Name) {
				container.VolumeMounts[i].MountPropagation = &mountPropagationHostToContainer
				needInjection = true
			}
		}
	}

	return containers, needInjection
}
