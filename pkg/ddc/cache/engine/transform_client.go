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
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

func (t *CacheEngine) transformClient(runtime *datav1alpha1.CacheRuntime, runtimeClass *datav1alpha1.CacheRuntimeClass, commonConfig *common.CacheRuntimeComponentCommonConfig, value *common.CacheRuntimeValue) (err error) {
	value.Client = &common.CacheRuntimeComponentValue{
		Name:          t.getComponentName(common.ComponentTypeClient),
		Namespace:     t.namespace,
		Enabled:       true,
		ComponentType: common.ComponentTypeClient,
	}
	if runtimeClass.Topology.Client == nil || runtime.Spec.Client.Disabled {
		value.Client.Enabled = false
		return nil
	}
	if len(value.Client.Namespace) == 0 {
		value.Client.Namespace = "default"
	}
	t.parseClientFromRuntimeClass(runtimeClass, value)

	t.addCommonConfigForClient(runtimeClass, commonConfig, value)

	if err := t.parseClientFromRuntime(runtime, value); err != nil {
		return err
	}
	return
}

func (e *CacheEngine) parseClientFromRuntimeClass(runtimeClass *datav1alpha1.CacheRuntimeClass, value *common.CacheRuntimeValue) {
	componentClient := runtimeClass.Topology.Client
	value.Client.WorkloadType = componentClient.WorkloadType

	value.Client.PodTemplateSpec = componentClient.PodTemplateSpec

	if runtimeClass.Topology.Client.Service.Headless != nil {
		value.Client.Service = e.transformHeadlessServiceValue(value)
	}
}

func (t *CacheEngine) parseClientFromRuntime(runtime *datav1alpha1.CacheRuntime, value *common.CacheRuntimeValue) error {
	podTemplateSpec := &value.Client.PodTemplateSpec

	// 1. image
	if len(runtime.Spec.Client.RuntimeVersion.Image) != 0 {
		podTemplateSpec.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", runtime.Spec.Client.RuntimeVersion.Image, runtime.Spec.Client.RuntimeVersion.ImageTag)
	}
	if len(runtime.Spec.Client.RuntimeVersion.ImagePullPolicy) != 0 {
		podTemplateSpec.Spec.Containers[0].ImagePullPolicy = corev1.PullPolicy(runtime.Spec.Client.RuntimeVersion.ImagePullPolicy)
	}
	if len(runtime.Spec.ImagePullSecrets) != 0 {
		podTemplateSpec.Spec.ImagePullSecrets = runtime.Spec.ImagePullSecrets
	}

	// 2. env
	if len(runtime.Spec.Client.Env) != 0 {
		podTemplateSpec.Spec.Containers[0].Env = append(value.Client.PodTemplateSpec.Spec.Containers[0].Env, runtime.Spec.Client.Env...)
	}
	podTemplateSpec.Spec.Containers[0].Env = append(podTemplateSpec.Spec.Containers[0].Env, t.generateCommonEnvs(runtime, value.Client.ComponentType)...)

	// 3. nodeSelector
	if len(runtime.Spec.Client.NodeSelector) != 0 {
		podTemplateSpec.Spec.NodeSelector = runtime.Spec.Client.NodeSelector
	}

	// 4. volume
	if len(runtime.Spec.Volumes) > 0 {
		podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, runtime.Spec.Volumes...)
	}

	if len(runtime.Spec.Client.VolumeMounts) > 0 {
		podTemplateSpec.Spec.Containers[0].VolumeMounts = append(podTemplateSpec.Spec.Containers[0].VolumeMounts, runtime.Spec.Client.VolumeMounts...)
	}

	// 5. metadate
	if len(runtime.Spec.PodMetadata.Annotations) != 0 {
		podTemplateSpec.Annotations = utils.UnionMapsWithOverride(value.Client.PodTemplateSpec.Annotations, runtime.Spec.PodMetadata.Annotations)
	}
	if len(runtime.Spec.PodMetadata.Labels) != 0 {
		podTemplateSpec.Labels = utils.UnionMapsWithOverride(value.Client.PodTemplateSpec.Labels, runtime.Spec.PodMetadata.Labels)
	}

	// 6. resources
	t.transformResourcesForContainer(runtime.Spec.Client.Resources, &value.Client.PodTemplateSpec.Spec.Containers[0])

	if len(runtime.Spec.Client.TieredStore.Levels) > 0 {
		tieredStoreConfig, err := t.transformTieredStore(runtime.Spec.Client.TieredStore.Levels)
		if err != nil {
			return err
		}
		for _, option := range tieredStoreConfig.CacheVolumeOptions {
			if option.MemQuantityRequirement != nil {
				if len(podTemplateSpec.Spec.Containers[0].Resources.Requests) == 0 {
					podTemplateSpec.Spec.Containers[0].Resources.Requests = make(corev1.ResourceList)
				}
				podTemplateSpec.Spec.Containers[0].Resources.Requests.Memory().Add(*option.MemQuantityRequirement)
			}
		}
		podTemplateSpec.Spec.Containers[0].VolumeMounts = append(
			podTemplateSpec.Spec.Containers[0].VolumeMounts, tieredStoreConfig.CacheVolumeMounts...)
		podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, tieredStoreConfig.CacheVolumes...)
		value.Client.TieredStore = tieredStoreConfig.CacheVolumeOptions
	}

	if podTemplateSpec.Spec.Containers[0].Lifecycle == nil {
		podTemplateSpec.Spec.Containers[0].Lifecycle = &corev1.Lifecycle{
			PreStop: &corev1.LifecycleHandler{
				Exec: &corev1.ExecAction{
					Command: []string{
						"umount",
						t.getTargetPath(),
					},
				},
			},
		}
	} else if podTemplateSpec.Spec.Containers[0].Lifecycle.PreStop == nil {
		podTemplateSpec.Spec.Containers[0].Lifecycle.PreStop = &corev1.LifecycleHandler{
			Exec: &corev1.ExecAction{
				Command: []string{
					"umount",
					t.getTargetPath(),
				},
			},
		}
	}

	targetPathVolumeConfig := t.transformTargetPathVolumes()
	podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, targetPathVolumeConfig.TargetPathHostVolume)
	podTemplateSpec.Spec.Containers[0].VolumeMounts = append(podTemplateSpec.Spec.Containers[0].VolumeMounts, targetPathVolumeConfig.TargetPathVolumeMount)
	value.Client.TargetPath = targetPathVolumeConfig.TargetPath

	if len(runtime.Spec.Client.Options) > 0 {
		value.Client.Options = utils.UnionMapsWithOverride(value.Client.Options, runtime.Spec.Client.Options)
	}

	podTemplateSpec.Spec.NodeSelector = map[string]string{}
	if len(runtime.Spec.Client.NodeSelector) > 0 {
		podTemplateSpec.Spec.NodeSelector = runtime.Spec.Client.NodeSelector
	}
	podTemplateSpec.Spec.NodeSelector[utils.GetFuseLabelName(runtime.Namespace, runtime.Name, "")] = "true"
	value.Client.NodeSelector = podTemplateSpec.Spec.NodeSelector

	return nil
}

func (t *CacheEngine) addCommonConfigForClient(runtimeClass *datav1alpha1.CacheRuntimeClass, commonConfig *common.CacheRuntimeComponentCommonConfig, value *common.CacheRuntimeValue) {
	componentClient := value.Client

	componentClient.PodTemplateSpec.Spec.ImagePullSecrets = append(
		componentClient.PodTemplateSpec.Spec.ImagePullSecrets, commonConfig.ImagePullSecrets...)

	componentClient.PodTemplateSpec.Spec.NodeSelector = utils.UnionMapsWithOverride(
		componentClient.PodTemplateSpec.Spec.NodeSelector, commonConfig.NodeSelector)

	componentClient.PodTemplateSpec.Spec.Tolerations = append(
		componentClient.PodTemplateSpec.Spec.Tolerations, commonConfig.Tolerations...)

	componentClient.PodTemplateSpec.Spec.Containers[0].Env = append(
		componentClient.PodTemplateSpec.Spec.Containers[0].Env, commonConfig.Envs...)

	componentClient.Owner = commonConfig.Owner
	componentClient.Options = utils.UnionMapsWithOverride(runtimeClass.Topology.Client.Options, commonConfig.Options)

	if runtimeClass.Topology.Client.Dependencies.EncryptOption != nil {
		for _, encryptOptionVolume := range commonConfig.EncryptOptionConfigs.EncryptOptionVolumes {
			value.Client.PodTemplateSpec.Spec.Volumes = append(value.Client.PodTemplateSpec.Spec.Volumes, encryptOptionVolume)
		}

		for _, encryptOptionVolumeMount := range commonConfig.EncryptOptionConfigs.EncryptOptionVolumeMounts {
			for i := range value.Client.PodTemplateSpec.Spec.Containers {
				value.Client.PodTemplateSpec.Spec.Containers[i].VolumeMounts = append(value.Client.PodTemplateSpec.Spec.Containers[i].VolumeMounts, encryptOptionVolumeMount)
			}
		}
		value.Master.EncryptOption = commonConfig.EncryptOptionConfigs.EncryptOptionConfig

	}
	value.Client.PodTemplateSpec.Spec.Volumes = append(value.Client.PodTemplateSpec.Spec.Volumes, commonConfig.RuntimeConfigConfig.RuntimeConfigVolume)
	value.Client.PodTemplateSpec.Spec.Containers[0].VolumeMounts = append(value.Client.PodTemplateSpec.Spec.Containers[0].VolumeMounts, commonConfig.RuntimeConfigConfig.RuntimeConfigVolumeMount)

}
