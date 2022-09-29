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
	"context"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/scripts/poststart"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

func (s *Injector) injectCheckMountReadyScript(pod common.FluidObject, runtimeInfos map[string]base.RuntimeInfoInterface) error {
	objMeta, err := pod.GetMetaObject()
	if err != nil {
		return err
	}

	var namespace string
	if len(runtimeInfos) == 0 {
		// Skip if no need to inject because no dataset pvc is mounted
		return nil
	}

	// Choose the first runtime info's namespace
	for _, v := range runtimeInfos {
		namespace = v.GetNamespace()
		break
	}

	appScriptGenerator, err := s.ensureScriptConfigMapExists(namespace)
	if err != nil {
		return err
	}

	volumes, err := pod.GetVolumes()
	if err != nil {
		return err
	}
	volumeToAdd := appScriptGenerator.GetVolume()
	conflictNames, volumes, err := s.appendVolumes(volumes, []corev1.Volume{volumeToAdd}, "")
	if err != nil {
		return err
	}

	containers, err := pod.GetContainers()
	if err != nil {
		return err
	}

	for ci := range containers {
		pathToRuntimeTypeMap := collectDatasetVolumeMountInfo(containers[ci].VolumeMounts, volumes, runtimeInfos)
		if len(pathToRuntimeTypeMap) == 0 {
			continue
		}

		volumeMountToAdd := appScriptGenerator.GetVolumeMount()
		if newName, found := conflictNames[volumeToAdd.Name]; found {
			volumeMountToAdd.Name = newName
		}

		containers[ci].VolumeMounts = append(containers[ci].VolumeMounts, volumeMountToAdd)
		if utils.AppContainerPostStartInjectEnabled(objMeta.Labels) {
			if containers[ci].Lifecycle != nil && containers[ci].Lifecycle.PostStart != nil {
				s.log.Info("container already has post start lifecycle, skip injection", "container name", containers[ci].Name)
			} else {
				if containers[ci].Lifecycle == nil {
					containers[ci].Lifecycle = &corev1.Lifecycle{}
				}

				mountPaths, mountTypes := assembleMountInfos(pathToRuntimeTypeMap)
				containers[ci].Lifecycle.PostStart = appScriptGenerator.GetPostStartCommand(mountPaths, mountTypes)
			}
		}
	}

	err = pod.SetContainers(containers)
	if err != nil {
		return err
	}

	initContainers, err := pod.GetInitContainers()
	if err != nil {
		return err
	}

	for ci := range initContainers {
		pathToRuntimeTypeMap := collectDatasetVolumeMountInfo(initContainers[ci].VolumeMounts, volumes, runtimeInfos)
		if len(pathToRuntimeTypeMap) == 0 {
			continue
		}

		volumeMountToAdd := appScriptGenerator.GetVolumeMount()
		if newName, found := conflictNames[volumeToAdd.Name]; found {
			volumeMountToAdd.Name = newName
		}

		initContainers[ci].VolumeMounts = append(initContainers[ci].VolumeMounts, volumeMountToAdd)
		if utils.AppContainerPostStartInjectEnabled(objMeta.Labels) {
			if initContainers[ci].Lifecycle != nil && initContainers[ci].Lifecycle.PostStart != nil {
				s.log.Info("container already has post start lifecycle, skip injection", "container name", initContainers[ci].Name)
			} else {
				if initContainers[ci].Lifecycle == nil {
					initContainers[ci].Lifecycle = &corev1.Lifecycle{}
				}

				mountPaths, mountTypes := assembleMountInfos(pathToRuntimeTypeMap)
				initContainers[ci].Lifecycle.PostStart = appScriptGenerator.GetPostStartCommand(mountPaths, mountTypes)
			}
		}
	}

	err = pod.SetInitContainers(initContainers)
	if err != nil {
		return err
	}

	err = pod.SetVolumes(volumes)
	if err != nil {
		return err
	}

	return nil
}

func (s *Injector) ensureScriptConfigMapExists(namespace string) (*poststart.ScriptGeneratorForApp, error) {
	appScriptGen := poststart.NewScriptGeneratorForApp(namespace)

	cm := appScriptGen.BuildConfigmap()
	cmFound, err := kubeclient.IsConfigMapExist(s.client, cm.Name, cm.Namespace)
	if err != nil {
		return nil, err
	}

	if !cmFound {
		err = s.client.Create(context.TODO(), cm)
		if err != nil {
			if otherErr := utils.IgnoreAlreadyExists(err); otherErr != nil {
				return nil, err
			}
		}
	}

	return appScriptGen, nil
}

func collectDatasetVolumeMountInfo(volMounts []corev1.VolumeMount, volumes []corev1.Volume, runtimeInfos map[string]base.RuntimeInfoInterface) map[string]string {
	path2RuntimeTypeMap := map[string]string{}
	for _, volMount := range volMounts {
		vol := utils.FindVolumeByVolumeMount(volMount, volumes)
		if vol == nil {
			// todo: log
			continue
		}

		if vol.PersistentVolumeClaim != nil {
			if ri, ok := runtimeInfos[vol.PersistentVolumeClaim.ClaimName]; ok {
				path2RuntimeTypeMap[volMount.MountPath] = ri.GetRuntimeType()
			}
		}
	}

	return path2RuntimeTypeMap
}

func assembleMountInfos(path2RuntimeTypeMap map[string]string) (mountPathStr, mountTypeStr string) {
	var (
		mountPaths []string
		mountTypes []string
	)

	for k, v := range path2RuntimeTypeMap {
		mountPaths = append(mountPaths, k)
		mountTypes = append(mountTypes, v)
	}

	mountPathStr = strings.Join(mountPaths, ":")
	mountTypeStr = strings.Join(mountTypes, ":")

	return
}
