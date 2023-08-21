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
	"errors"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
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

// checkAndOverrideInitPVC checks if the dataset PVC used in init phase has been specified and overrides them by emptyDir
func (s *Injector) checkAndOverrideInitPVC(dsName2SourceFiles map[string]string, runtimeInfos map[string]base.RuntimeInfoInterface, pod common.FluidObject) (err error) {
	initContainers, err := pod.GetInitContainers()
	if err != nil {
		return err
	}

	if len(initContainers) == 0 {
		return nil
	}
	volumes, err := pod.GetVolumes()
	if err != nil {
		return err
	}

	volumes2pvcName := map[string]string{}
	for _, volume := range volumes {
		if volume.PersistentVolumeClaim != nil {
			pvcNmae := volume.PersistentVolumeClaim.ClaimName
			if _, isExist := runtimeInfos[pvcNmae]; isExist {
				volumes2pvcName[volume.Name] = pvcNmae
			}
		}
	}

	for _, container := range initContainers {
		for index, volume := range container.VolumeMounts {
			volumeName := volume.Name
			pvcName, isDatasetPVC := volumes2pvcName[volumeName]
			_, isSpecified := dsName2SourceFiles[pvcName]

			// the PVC is dataset PVC used in init phase, but not specified
			if isDatasetPVC && !isSpecified {
				return errors.New(volumeName + " used in init phase, but not specified!")
			}
			if isDatasetPVC && isSpecified {
				container.VolumeMounts[index].Name = common.InitPrefix + pvcName
			}
		}
	}

	err = pod.SetInitContainers(initContainers)

	return err
}

// overrideVolumeMountName override the volumeMount name mapping key to value of volumeNamesConflict
func (s *Injector) overrideVolumeMountName(containers []corev1.Container, volumeNamesConflict map[string]string) (newContainers []corev1.Container) {
	if len(containers) == 0 {
		return newContainers
	}

	for _, container := range containers {
		volumeMounts := container.VolumeMounts
		for index, volumeMount := range volumeMounts {
			if newName, exist := volumeNamesConflict[volumeMount.Name]; exist {
				container.VolumeMounts[index].Name = newName
			}
		}
		newContainers = append(newContainers, container)
	}
	return newContainers
}
