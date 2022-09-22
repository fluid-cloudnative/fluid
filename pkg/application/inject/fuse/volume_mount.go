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
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/scripts/poststart"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

func mutateVolumeMounts(containers []corev1.Container, appScriptGenerator *poststart.ScriptGeneratorForApp, datasetVolumeNames []string) (retContainers []corev1.Container, needInjection bool) {
	mountPropagationHostToContainer := corev1.MountPropagationHostToContainer

	for ci, container := range containers {

		needAppScript := false
		// Set HostToContainer to the dataset volume mount point
		for i, volumeMount := range container.VolumeMounts {
			if utils.ContainsString(datasetVolumeNames, volumeMount.Name) {
				container.VolumeMounts[i].MountPropagation = &mountPropagationHostToContainer

				// Inject postStartHook only when appScriptGenerator is non-null, which means fuse sidecar injection in privileged mode.
				if appScriptGenerator != nil {
					if postStart := appScriptGenerator.GetPostStartCommand(volumeMount.MountPath); postStart != nil {
						if containers[ci].Lifecycle != nil && containers[ci].Lifecycle.PostStart != nil {
							warningStr := fmt.Sprintf("===> Skipping inject post start command because container %v already have one", containers[ci].Name)
							log.Info(warningStr)
							continue
						} else {
							if containers[ci].Lifecycle == nil {
								containers[ci].Lifecycle = &corev1.Lifecycle{}
							}
							containers[ci].Lifecycle.PostStart = appScriptGenerator.GetPostStartCommand(volumeMount.MountPath)
						}
					}

				}

				needInjection = true
				needAppScript = true
			}
		}

		// Add volumeMounts only when the container mounts some dataset pvc and appScriptGenerator is non-null,
		// which means fuse sidecar injection in privileged mode.
		if needAppScript && appScriptGenerator != nil {
			containers[ci].VolumeMounts = append(containers[ci].VolumeMounts, appScriptGenerator.GetVolumeMount())
		}
	}

	return containers, needInjection
}
