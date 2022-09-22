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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

func injectFuseContainerToFirst(containers []corev1.Container, fuseContainerName string,
	template *common.FuseInjectionTemplate,
	volumeNamesConflict map[string]string) []corev1.Container {
	fuseContainer := template.FuseContainer
	fuseContainer.Name = fuseContainerName

	if strings.HasSuffix(fuseContainerName, common.InitFuseContainerName) {
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

//func (s *Injector) mutateContainers(keyName types.NamespacedName, fuseContainerName string,
//	containers []corev1.Container,
//	datasetVolumeNames []string,
//	template *common.FuseInjectionTemplate,
//	volumeNamesConflict map[string]string,
//	appScriptGenerator *poststart.ScriptGeneratorForApp) (result []corev1.Container,
//	injectFuseContainer bool) {
//
//	mountPropagationHostToContainer := corev1.MountPropagationHostToContainer
//	for ci, container := range containers {
//		// Skip injection for injected container
//		if container.Name == fuseContainerName {
//			warningStr := fmt.Sprintf("===> Skipping injection because %v has injected %q sidecar already\n",
//				keyName, fuseContainerName)
//			log.Info(warningStr)
//			break
//		}
//
//		// Add volumeMounts only when appScriptGenerator is non-null, which means fuse sidecar injection in privileged mode.
//		if appScriptGenerator != nil {
//			containers[ci].VolumeMounts = append(containers[ci].VolumeMounts, appScriptGenerator.GetVolumeMount())
//		}
//
//		// Set mountPropagationHostToContainer to the dataset volume mount point, and set Injection true
//		for i, volumeMount := range container.VolumeMounts {
//			if utils.ContainsString(datasetVolumeNames, volumeMount.Name) {
//				container.VolumeMounts[i].MountPropagation = &mountPropagationHostToContainer
//				// Inject postStartHook only when appScriptGenerator is non-null, which means fuse sidecar injection in privileged mode.
//				if appScriptGenerator != nil {
//					if postStart := appScriptGenerator.GetPostStartCommand(volumeMount.MountPath); postStart != nil {
//						if containers[ci].Lifecycle != nil && containers[ci].Lifecycle.PostStart != nil {
//							warningStr := fmt.Sprintf("===> Skipping inject post start command because container %v already have one", containers[ci].Name)
//							log.Info(warningStr)
//							continue
//						} else {
//							if containers[ci].Lifecycle == nil {
//								containers[ci].Lifecycle = &corev1.Lifecycle{}
//							}
//							containers[ci].Lifecycle.PostStart = appScriptGenerator.GetPostStartCommand(volumeMount.MountPath)
//						}
//					}
//				}
//
//				// ##### Why we need injectFuseContainer? IMO, it is guaranteed to has some container mounted at least one Fluid PVC
//				injectFuseContainer = true
//			}
//		}
//
//	}
//
//	if !injectFuseContainer {
//		return containers, injectFuseContainer
//	}
//
//	fuseContainer := template.FuseContainer
//	fuseContainer.Name = fuseContainerName
//	if fuseContainerName == common.InitFuseContainerName {
//		fuseContainer.Lifecycle = nil
//		// TODO(zhihao): for init container, it will inject a customized command
//		fuseContainer.Command = []string{"sleep"}
//		fuseContainer.Args = []string{"2s"}
//	}
//	for oldName, newName := range volumeNamesConflict {
//		for i, volumeMount := range fuseContainer.VolumeMounts {
//			if volumeMount.Name == oldName {
//				fuseContainer.VolumeMounts[i].Name = newName
//			}
//		}
//	}
//	if injectFuseContainer {
//		containers = append([]corev1.Container{fuseContainer}, containers...)
//	}
//
//	return containers, injectFuseContainer
//}
