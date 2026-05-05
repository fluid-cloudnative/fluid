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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

// transformEncryptOptionsToComponentVolumes transforms encrypt options from dataset spec to component pod volumes
// This function can be reused for both Master and Worker components
func (e *CacheEngine) transformEncryptOptionsToComponentVolumes(dataset *datav1alpha1.Dataset, component *common.CacheRuntimeComponentValue) {
	if component == nil || !component.Enabled || len(component.PodTemplateSpec.Spec.Containers) == 0 {
		return
	}

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
		component.PodTemplateSpec.Spec.Volumes = utils.AppendOrOverrideVolume(
			component.PodTemplateSpec.Spec.Volumes, volumeToAdd)

		volumeMountToAdd := corev1.VolumeMount{
			Name:      volName,
			ReadOnly:  true,
			MountPath: getSecretMountPath(secretName),
		}
		component.PodTemplateSpec.Spec.Containers[0].VolumeMounts = utils.AppendOrOverrideVolumeMounts(
			component.PodTemplateSpec.Spec.Containers[0].VolumeMounts, volumeMountToAdd)
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
