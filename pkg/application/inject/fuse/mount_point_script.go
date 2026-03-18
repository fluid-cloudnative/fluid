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
	"fmt"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/application/inject/fuse/mutator"
	"github.com/fluid-cloudnative/fluid/pkg/application/inject/fuse/poststart"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
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
	cmKey := fmt.Sprintf("%s/%s", cm.Namespace, cm.Name)

	existingCM, err := kubeclient.GetConfigmapByName(s.client, cm.Name, cm.Namespace)
	if err != nil {
		s.log.Error(err, "error when getting configMap", "cm.Name", cm.Name, "cm.Namespace", cm.Namespace)
		return nil, err
	}

	if existingCM == nil {
		// ConfigMap does not exist, create it
		s.log.V(1).Info("configMap not found, creating", "configMap", cmKey)
		err = s.client.Create(context.TODO(), cm)
		if err != nil {
			if otherErr := utils.IgnoreAlreadyExists(err); otherErr != nil {
				s.log.Error(err, "error when creating new configMap", "cm.Name", cm.Name, "cm.Namespace", cm.Namespace)
				return nil, err
			}
			s.log.V(1).Info("configmap already exists (concurrent creation), skip", "configMap", cmKey)
		}
		return appScriptGen, nil
	}

	// ConfigMap exists, check if the script SHA256 annotation matches; update with retry on conflict.
	latestSHA256 := appScriptGen.GetScriptSHA256()
	if err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		current, getErr := kubeclient.GetConfigmapByName(s.client, cm.Name, cm.Namespace)
		if getErr != nil {
			return getErr
		}
		if current == nil {
			// Deleted between Get calls; recreate
			return s.client.Create(context.TODO(), cm)
		}
		if current.Annotations != nil {
			if annotationSHA256, ok := current.Annotations[common.AnnotationCheckMountScriptSHA256]; ok && annotationSHA256 == latestSHA256 {
				s.log.V(1).Info("configmap script is up-to-date, skip update", "configMap", cmKey)
				return nil
			}
		}
		// SHA256 mismatch or annotation missing: update the ConfigMap with latest script and SHA256
		s.log.Info("configmap script SHA256 mismatch or annotation missing, updating", "configMap", cmKey, "expectedSHA256", latestSHA256)
		current.Data = cm.Data
		if current.Annotations == nil {
			current.Annotations = map[string]string{}
		}
		current.Annotations[common.AnnotationCheckMountScriptSHA256] = latestSHA256
		return s.client.Update(context.TODO(), current)
	}); err != nil {
		s.log.Error(err, "error when ensuring configMap is up-to-date", "cm.Name", cm.Name, "cm.Namespace", cm.Namespace)
		return nil, err
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
