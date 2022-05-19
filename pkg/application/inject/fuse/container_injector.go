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

package fuse

import (
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (s *Injector) mutateContainers(keyName types.NamespacedName, fuseContainerName string,
	containers []corev1.Container, privileged bool,
	datasetVolumeNames []string,
	template *common.FuseInjectionTemplate,
	volumeNamesConflict map[string]string) (result []corev1.Container,
	injectFuseContainer bool) {

	mountPropagationHostToContainer := corev1.MountPropagationHostToContainer
	for _, container := range containers {
		// Skip injection for injected container
		if container.Name == fuseContainerName {
			warningStr := fmt.Sprintf("===> Skipping injection because %v has injected %q sidecar already\n",
				keyName, fuseContainerName)
			log.Info(warningStr)
			break
		}

		// Set mountPropagationHostToContainer to the dataset volume mount point, and set Injection true
		for i, volumeMount := range container.VolumeMounts {
			if utils.ContainsString(datasetVolumeNames, volumeMount.Name) {
				container.VolumeMounts[i].MountPropagation = &mountPropagationHostToContainer
				injectFuseContainer = true
			}
		}

	}

	if !injectFuseContainer {
		return containers, injectFuseContainer
	}

	fuseContainer := template.FuseContainer
	fuseContainer.Name = fuseContainerName
	if fuseContainerName == common.InitFuseContainerName {
		fuseContainer.Lifecycle = nil
		// TODO(zhihao): for init container, it will inject a customized command
		fuseContainer.Command = []string{"sleep"}
		fuseContainer.Args = []string{"2s"}
	}
	for oldName, newName := range volumeNamesConflict {
		for i, volumeMount := range fuseContainer.VolumeMounts {
			if volumeMount.Name == oldName {
				fuseContainer.VolumeMounts[i].Name = newName
			}
		}
	}
	if injectFuseContainer {
		containers = append([]corev1.Container{fuseContainer}, containers...)
	}

	return containers, injectFuseContainer
}
