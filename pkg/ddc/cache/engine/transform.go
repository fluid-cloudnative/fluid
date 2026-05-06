/*
  Copyright 2026 The Fluid Authors.

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
	"errors"
	"fmt"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/transformer"
	corev1 "k8s.io/api/core/v1"
)

// CacheRuntimeComponentCommonConfig common config for transform
type CacheRuntimeComponentCommonConfig struct {
	Owner *common.OwnerReference

	// TODO: add ImagePullSecrets, NodeSelector, Tolerations, Envs, PlacementMode etc.

	// configmaps mounted by all component pods
	RuntimeConfigs *RuntimeConfigVolumeConfig
}

type TargetPathVolumeConfig struct {
	TargetPathHostVolume  corev1.Volume
	TargetPathVolumeMount corev1.VolumeMount
}

type RuntimeConfigVolumeConfig struct {
	// runtime config's config map defined by fluid
	RuntimeConfigVolume      corev1.Volume
	RuntimeConfigVolumeMount corev1.VolumeMount
	// config map names defined in ClassRuntimeClass
	ExtraConfigMapNames map[string]bool
}

func (e *CacheEngine) transform(dataset *datav1alpha1.Dataset, runtime *datav1alpha1.CacheRuntime, runtimeClass *datav1alpha1.CacheRuntimeClass) (*common.CacheRuntimeValue, error) {

	if runtimeClass.Topology == nil ||
		(runtimeClass.Topology.Master == nil && runtimeClass.Topology.Worker == nil && runtimeClass.Topology.Client == nil) {
		return nil, fmt.Errorf("at least one component should be defined in runtimeClass")
	}
	defer utils.TimeTrack(time.Now(), "CacheRuntime.transform", "name", runtime.Name)

	runtimeValue := &common.CacheRuntimeValue{
		RuntimeIdentity: common.RuntimeIdentity{
			Namespace: runtime.Namespace,
			Name:      runtime.Name,
		},
	}

	// get common config for transform components
	runtimeCommonConfig, err := e.transformComponentCommonConfig(runtime, runtimeClass)
	if err != nil {
		return nil, err
	}

	// transform the master/worker/client
	err = e.transformMaster(dataset, runtime, runtimeClass, runtimeCommonConfig, runtimeValue)
	if err != nil {
		return nil, err
	}
	err = e.transformWorker(dataset, runtime, runtimeClass, runtimeCommonConfig, runtimeValue)
	if err != nil {
		return nil, err
	}
	err = e.transformClient(dataset, runtime, runtimeClass, runtimeCommonConfig, runtimeValue)
	if err != nil {
		return nil, err
	}

	return runtimeValue, nil
}

func (e *CacheEngine) transformComponentCommonConfig(runtime *datav1alpha1.CacheRuntime, runtimeClass *datav1alpha1.CacheRuntimeClass) (*CacheRuntimeComponentCommonConfig, error) {
	config := &CacheRuntimeComponentCommonConfig{
		Owner: transformer.GenerateOwnerReferenceFromObject(runtime),
	}
	e.transformRuntimeConfigVolume(config, runtimeClass)

	return config, nil
}

func (e *CacheEngine) transformRuntimeConfigVolume(config *CacheRuntimeComponentCommonConfig, runtimeClass *datav1alpha1.CacheRuntimeClass) {
	// create the runtime config mount info
	volumeName := e.getRuntimeConfigVolumeName()
	config.RuntimeConfigs = &RuntimeConfigVolumeConfig{
		RuntimeConfigVolume: corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: e.getRuntimeConfigConfigMapName(),
					},
				},
			},
		},
		RuntimeConfigVolumeMount: corev1.VolumeMount{
			Name:      volumeName,
			MountPath: e.getRuntimeConfigDir(),
			ReadOnly:  true,
		},
	}

	if len(runtimeClass.ExtraResources.ConfigMaps) == 0 {
		return
	}
	config.RuntimeConfigs.ExtraConfigMapNames = map[string]bool{}
	// TODO: 当前，这些 configmap 当前需要 component 中定义使用，是否对于所有 component 是通用的？
	for _, cm := range runtimeClass.ExtraResources.ConfigMaps {
		config.RuntimeConfigs.ExtraConfigMapNames[cm.Name] = true
	}
}

func (e *CacheEngine) addCommonConfigForComponent(commonConfig *CacheRuntimeComponentCommonConfig, componentValue *common.CacheRuntimeComponentValue,
	componentDefinition *datav1alpha1.RuntimeComponentDefinition) error {
	componentValue.PodTemplateSpec.Spec.Volumes = append(componentValue.PodTemplateSpec.Spec.Volumes, commonConfig.RuntimeConfigs.RuntimeConfigVolume)

	if len(componentValue.PodTemplateSpec.Spec.Containers) == 0 {
		return fmt.Errorf("component %s must define at least one container", componentValue.ComponentType)
	}

	// assume the first container uses the runtime config
	if len(componentValue.PodTemplateSpec.Spec.InitContainers) > 0 {
		componentValue.PodTemplateSpec.Spec.InitContainers[0].VolumeMounts = append(componentValue.PodTemplateSpec.Spec.InitContainers[0].VolumeMounts, commonConfig.RuntimeConfigs.RuntimeConfigVolumeMount)
	}
	componentValue.PodTemplateSpec.Spec.Containers[0].VolumeMounts = append(componentValue.PodTemplateSpec.Spec.Containers[0].VolumeMounts, commonConfig.RuntimeConfigs.RuntimeConfigVolumeMount)

	// other config maps defined in CacheRuntimeClass
	if componentDefinition.Dependencies.ExtraResources == nil {
		return nil
	}
	names := commonConfig.RuntimeConfigs.ExtraConfigMapNames
	for _, cm := range componentDefinition.Dependencies.ExtraResources.ConfigMaps {
		if !names[cm.Name] {
			e.Log.Error(errors.New("component has undefined config map extra resource"), "type", componentValue.ComponentType, "configMapName", cm.Name)
		}
		componentValue.PodTemplateSpec.Spec.Volumes = append(componentValue.PodTemplateSpec.Spec.Volumes, corev1.Volume{
			Name: e.getRuntimeClassExtraConfigMapVolumeName(cm.Name),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cm.Name,
					},
				},
			},
		})
		if len(componentValue.PodTemplateSpec.Spec.InitContainers) > 0 {
			componentValue.PodTemplateSpec.Spec.InitContainers[0].VolumeMounts = append(componentValue.PodTemplateSpec.Spec.InitContainers[0].VolumeMounts,
				corev1.VolumeMount{
					Name:      e.getRuntimeClassExtraConfigMapVolumeName(cm.Name),
					MountPath: cm.MountPath,
					ReadOnly:  true,
				})
		}
		componentValue.PodTemplateSpec.Spec.Containers[0].VolumeMounts = append(componentValue.PodTemplateSpec.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      e.getRuntimeClassExtraConfigMapVolumeName(cm.Name),
			MountPath: cm.MountPath,
			ReadOnly:  true,
		})
	}

	// add envs
	serviceName := ""
	if componentValue.Service != nil {
		serviceName = componentValue.Service.Name
	}
	addEnvs := []corev1.EnvVar{
		{
			Name:  "FLUID_DATASET_NAME",
			Value: e.name,
		},
		{
			Name:  "FLUID_DATASET_NAMESPACE",
			Value: e.namespace,
		},
		{
			Name:  "FLUID_RUNTIME_CONFIG_PATH",
			Value: e.getRuntimeConfigPath(),
		},
		{
			Name:  "FLUID_RUNTIME_MOUNT_PATH",
			Value: e.getFuseMountPoint(),
		},
		{
			Name:  "FLUID_RUNTIME_COMPONENT_TYPE",
			Value: string(componentValue.ComponentType),
		},
		{
			// curvine master sets the CURVINE_MASTER_HOSTNAME with service name
			Name:  "FLUID_RUNTIME_COMPONENT_SVC_NAME",
			Value: serviceName,
		},
	}
	// inject envs should come first.
	componentValue.PodTemplateSpec.Spec.Containers[0].Env = append(addEnvs, componentValue.PodTemplateSpec.Spec.Containers[0].Env...)
	if len(componentValue.PodTemplateSpec.Spec.InitContainers) > 0 {
		componentValue.PodTemplateSpec.Spec.InitContainers[0].Env = append(addEnvs, componentValue.PodTemplateSpec.Spec.InitContainers[0].Env...)
	}
	return nil
}
