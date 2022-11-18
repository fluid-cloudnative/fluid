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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

func injectFuseContainerToFirst(containers []corev1.Container, fuseContainerName string,
	template *common.FuseInjectionTemplate,
	volumeNamesConflict map[string]string) []corev1.Container {
	fuseContainer := template.FuseContainer
	fuseContainer.Name = fuseContainerName

	if strings.HasPrefix(fuseContainerName, common.InitFuseContainerName) {
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

	containers = append([]corev1.Container{fuseContainer}, containers...)
	return containers
}

func collectAllContainerNames(pod common.FluidObject) ([]string, error) {
	var allContainerNames []string

	containers, err := pod.GetContainers()
	if err != nil {
		return allContainerNames, err
	}

	for _, c := range containers {
		allContainerNames = append(allContainerNames, c.Name)
	}

	initContainers, err := pod.GetInitContainers()
	if err != nil {
		return allContainerNames, err
	}

	for _, c := range initContainers {
		allContainerNames = append(allContainerNames, c.Name)
	}

	return allContainerNames, nil
}
