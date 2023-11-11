/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package fuse

import (
	"context"
	"fmt"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/application/inject/fuse/mutator"
	"github.com/fluid-cloudnative/fluid/pkg/application/inject/fuse/poststart"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
)

func (s *Injector) injectCheckMountReadyScript(podSpecs *mutator.MutatingPodSpecs, runtimeInfos map[string]base.RuntimeInfoInterface) error {
	objMeta := podSpecs.MetaObj

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

	podName := objMeta.Name
	if len(podName) == 0 {
		podName = objMeta.GenerateName
	}

	s.log.V(1).Info("before inject CheckMountReadyScript volume", "pod namespace", namespace, "pod name", podName)
	// Cannot use objMeta.namespace as the expected namespace because it may be empty and not trustworthy before K8s 1.24.
	// For more details, see https://github.com/kubernetes/website/issues/30574#issuecomment-974896246
	appScriptGenerator, err := s.ensureScriptConfigMapExists(namespace)
	if err != nil {
		return err
	}

	volumes := podSpecs.Volumes
	volumeToAdd := appScriptGenerator.GetVolume()
	conflictNames, volumes, err := s.appendVolumes(volumes, []corev1.Volume{volumeToAdd}, "")
	if err != nil {
		return err
	}

	s.log.V(1).Info("before inject CheckMountReadyScript volume mount to containers", "pod namespace", namespace, "pod name", podName)
	containers := podSpecs.Containers
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
	podSpecs.Containers = containers

	s.log.V(1).Info("before inject CheckMountReadyScript volume mount to initContainers", "pod namespace", namespace, "pod name", podName)
	initContainers := podSpecs.InitContainers
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

	podSpecs.InitContainers = initContainers
	podSpecs.Volumes = volumes

	return nil
}

func (s *Injector) ensureScriptConfigMapExists(namespace string) (*poststart.ScriptGeneratorForApp, error) {
	appScriptGen := poststart.NewScriptGeneratorForApp(namespace)

	cm := appScriptGen.BuildConfigmap()
	cmFound, err := kubeclient.IsConfigMapExist(s.client, cm.Name, cm.Namespace)
	if err != nil {
		s.log.Error(err, "error when checking configMap's existence", "cm.Name", cm.Name, "cm.Namespace", cm.Namespace)
		return nil, err
	}

	cmKey := fmt.Sprintf("%s/%s", cm.Namespace, cm.Name)
	s.log.V(1).Info("after check configMap existence", "configMap", cmKey, "existence", cmFound)
	if !cmFound {
		err = s.client.Create(context.TODO(), cm)
		if err != nil {
			if otherErr := utils.IgnoreAlreadyExists(err); otherErr != nil {
				s.log.Error(err, "error when creating new configMap", "cm.Name", cm.Name, "cm.Namespace", cm.Namespace)
				return nil, err
			} else {
				s.log.V(1).Info("configmap already exists, skip", "configMap", cmKey)
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
