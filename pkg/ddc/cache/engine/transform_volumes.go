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
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

// transformVolumes consolidates all volume-related transformations for a component
// This function handles:
// 1. Runtime config volume and volume mount
// 2. Extra config map volumes and volume mounts
// 3. Runtime spec volumes and volume mounts
// 4. Encrypt options to component volumes
func (e *CacheEngine) transformVolumes(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount,
	dataset *datav1alpha1.Dataset, componentDefinition *datav1alpha1.RuntimeComponentDefinition,
	commonConfig *CacheRuntimeComponentCommonConfig, defaultMountSecrets bool, podSpec *corev1.PodSpec) error {

	// 1. Transform runtime config volume and mount
	e.applyRuntimeConfigVolume(podSpec, commonConfig)

	// 2. Transform extra config map volumes and mounts
	err := e.transformExtraConfigMapVolumes(commonConfig, podSpec, componentDefinition.Dependencies.ExtraResources)
	if err != nil {
		return err
	}

	// 3. Transform runtime spec volumes
	err = e.transformRuntimeSpecVolumes(volumes, volumeMounts, podSpec)
	if err != nil {
		return err
	}

	// 4. Transform encrypt options to component volumes (default enabled for Worker/Master, disabled for Client)
	var shouldMount = shouldMountSecrets(componentDefinition.Dependencies.SecretMount, defaultMountSecrets)
	if shouldMount {
		e.transformEncryptOptionsToComponentVolumes(dataset, podSpec)
	}

	return nil
}

// applyRuntimeConfigVolume adds runtime config volume and mount to the component
func (e *CacheEngine) applyRuntimeConfigVolume(podSpec *corev1.PodSpec, commonConfig *CacheRuntimeComponentCommonConfig) {
	if commonConfig == nil || commonConfig.RuntimeConfigs.RuntimeConfigVolume.Name == "" {
		return
	}

	// Add runtime config volume
	podSpec.Volumes = append(
		podSpec.Volumes,
		commonConfig.RuntimeConfigs.RuntimeConfigVolume,
	)

	// Add runtime config volume mount to init container if exists
	if len(podSpec.InitContainers) > 0 {
		podSpec.InitContainers[0].VolumeMounts = append(
			podSpec.InitContainers[0].VolumeMounts,
			commonConfig.RuntimeConfigs.RuntimeConfigVolumeMount,
		)
	}

	// Add runtime config volume mount to main container
	if len(podSpec.Containers) > 0 {
		podSpec.Containers[0].VolumeMounts = append(
			podSpec.Containers[0].VolumeMounts,
			commonConfig.RuntimeConfigs.RuntimeConfigVolumeMount,
		)
	}
}

// transformExtraConfigMapVolumes transforms extra config map resources to volumes and volume mounts
func (e *CacheEngine) transformExtraConfigMapVolumes(
	commonConfig *CacheRuntimeComponentCommonConfig,
	podSpec *corev1.PodSpec,
	resources *datav1alpha1.ExtraResourcesComponentDependency,
) error {
	// other config maps defined in CacheRuntimeClass
	if resources == nil {
		return nil
	}
	names := commonConfig.RuntimeConfigs.ExtraConfigMapNames
	for _, cm := range resources.ConfigMaps {
		if !names[cm.Name] {
			return fmt.Errorf("component has undefined config map extra resource '%s', check the CacheRuntimeClass definition", cm.Name)
		}
		volumeName := e.getRuntimeClassExtraConfigMapVolumeName(cm.Name)
		podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cm.Name,
					},
				},
			},
		})
		if len(podSpec.InitContainers) > 0 {
			podSpec.InitContainers[0].VolumeMounts = append(podSpec.InitContainers[0].VolumeMounts,
				corev1.VolumeMount{
					Name:      volumeName,
					MountPath: cm.MountPath,
					ReadOnly:  true,
				})
		}
		podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: cm.MountPath,
			ReadOnly:  true,
		})
	}
	return nil
}

// transformEncryptOptionsToComponentVolumes transforms encrypt options from dataset spec to component pod volumes
// This function can be reused for both Master and Worker components
func (e *CacheEngine) transformEncryptOptionsToComponentVolumes(dataset *datav1alpha1.Dataset, podSpec *corev1.PodSpec) {
	// Helper to add secret volume and mount to the component
	addSecret := func(secretName string) {
		if secretName == "" {
			return
		}
		volName := getSecretVolumeName(secretName)
		volumeToAdd := corev1.Volume{
			Name: volName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secretName,
				},
			},
		}
		podSpec.Volumes = utils.AppendOrOverrideVolume(
			podSpec.Volumes, volumeToAdd)

		volumeMountToAdd := corev1.VolumeMount{
			Name:      volName,
			ReadOnly:  true,
			MountPath: getSecretMountPath(secretName),
		}
		podSpec.Containers[0].VolumeMounts = utils.AppendOrOverrideVolumeMounts(
			podSpec.Containers[0].VolumeMounts, volumeMountToAdd)
	}

	// 1. Process shared encrypt options once
	for _, encryptOpt := range dataset.Spec.SharedEncryptOptions {
		addSecret(encryptOpt.ValueFrom.SecretKeyRef.Name)
	}

	// 2. Process mount-specific encrypt options, override shared options
	for _, m := range dataset.Spec.Mounts {
		if common.IsFluidNativeScheme(m.MountPoint) {
			continue
		}
		for _, encryptOpt := range m.EncryptOptions {
			addSecret(encryptOpt.ValueFrom.SecretKeyRef.Name)
		}
	}
}

// shouldMountSecrets determines whether secrets should be mounted based on configuration and default behavior
// config: the SecretMount configuration from CacheRuntimeClass (can be nil)
// defaultEnabled: the default behavior when config is nil or not explicitly set
func shouldMountSecrets(config *datav1alpha1.SecretMountComponentDependency, defaultEnabled bool) bool {
	if config == nil {
		return defaultEnabled
	}
	return config.Enabled
}

// transformRuntimeSpecVolumes transforms volumes and volumeMounts from CacheRuntimeSpec to PodTemplateSpec
func (e *CacheEngine) transformRuntimeSpecVolumes(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount, podSpec *corev1.PodSpec) error {
	// podTemplateSpec will not be nil

	// Create a map to track existing volumes in PodTemplateSpec
	existingVolumeMap := make(map[string]bool)
	// First pass: add volumes that don't already exist
	for _, volume := range volumes {
		if !existingVolumeMap[volume.Name] {
			existingVolumeMap[volume.Name] = true
			podSpec.Volumes = append(podSpec.Volumes, volume)
		}
	}

	// Second pass: process volumeMounts
	for _, volumeMount := range volumeMounts {
		// Check if corresponding volume exists
		if !existingVolumeMap[volumeMount.Name] {
			return fmt.Errorf("volume not found for volumeMount %s, check the CacheRuntime Spec", volumeMount.Name)
		}

		// Add volumeMount to the first container
		if len(podSpec.Containers) > 0 {
			podSpec.Containers[0].VolumeMounts = append(
				podSpec.Containers[0].VolumeMounts, volumeMount,
			)
		}
	}

	return nil
}
