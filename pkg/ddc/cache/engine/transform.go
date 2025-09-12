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
	"path/filepath"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/transformer"
)

func (e *CacheEngine) transform(dataset *datav1alpha1.Dataset, runtime *datav1alpha1.CacheRuntime, runtimeClass *datav1alpha1.CacheRuntimeClass) (*common.CacheRuntimeValue, error) {
	if dataset == nil {
		return nil, fmt.Errorf("the dataset is null")
	}

	if runtime == nil {
		return nil, fmt.Errorf("the cacheRuntime is null")
	}

	if err := e.precheckRuntimeClass(runtimeClass); err != nil {
		return nil, err
	}

	defer utils.TimeTrack(time.Now(), "CacheRuntime.Transform", "name", runtime.Name)

	runtimeValue := &common.CacheRuntimeValue{
		RuntimeIdentity: common.RuntimeIdentity{
			Namespace: runtime.Namespace,
			Name:      runtime.Name,
		},
	}

	runtimeCommonConfig, err := e.transformComponentCommonConfig(dataset, runtime)
	if err != nil {
		return nil, err
	}

	runtimeValue.FullnameOverride = e.name

	// set the placementMode
	e.transformPlacementMode(dataset, runtimeValue)

	// transform the workers
	if err := e.transformMasters(runtime, runtimeClass, runtimeCommonConfig, runtimeValue); err != nil {
		return nil, err
	}

	// transform the workers
	if err := e.transformWorkers(runtime, runtimeClass, runtimeCommonConfig, runtimeValue); err != nil {
		return nil, err
	}

	// transform the client
	if err := e.transformClient(runtime, runtimeClass, runtimeCommonConfig, runtimeValue); err != nil {
		return nil, err
	}

	return runtimeValue, nil
}

func (t *CacheEngine) precheckRuntimeClass(runtimeClass *datav1alpha1.CacheRuntimeClass) error {
	if runtimeClass == nil {
		return fmt.Errorf("cacheRuntimeClass is null")
	}

	if runtimeClass.Topology == nil {
		return fmt.Errorf("topology in cacheRuntimeClass is null")
	}

	if runtimeClass.Topology.Master == nil && runtimeClass.Topology.Worker == nil && runtimeClass.Topology.Client == nil {
		return fmt.Errorf("at least one component should be defined in runtimeClass")
	}
	return nil
}

func (t *CacheEngine) transformComponentCommonConfig(dataset *datav1alpha1.Dataset, runtime *datav1alpha1.CacheRuntime) (*common.CacheRuntimeComponentCommonConfig, error) {
	config := &common.CacheRuntimeComponentCommonConfig{
		Owner: transformer.GenerateOwnerReferenceFromObject(runtime),
	}

	config.ImagePullSecrets = runtime.Spec.ImagePullSecrets

	t.transformTolerations(dataset, config)
	t.transformEncryptOptionVolumes(dataset, config)
	t.transformRuntimeConfigVolume(config)

	return config, nil
}

func (t *CacheEngine) generateCommonEnvs(runtime *datav1alpha1.CacheRuntime, componentType common.ComponentType) []corev1.EnvVar {
	return []corev1.EnvVar{
		corev1.EnvVar{
			Name:  "FLUID_DATASET_NAME",
			Value: runtime.Name,
		},
		corev1.EnvVar{
			Name:  "FLUID_DATASET_NAMESPACE",
			Value: runtime.Namespace,
		},
		corev1.EnvVar{
			Name:  "FLUID_RUNTIME_CONFIG_PATH",
			Value: t.getRuntimeConfigPath(),
		},
		corev1.EnvVar{
			Name:  "FLUID_RUNTIME_MOUNT_PATH",
			Value: t.getTargetPath(),
		},
		corev1.EnvVar{
			Name:  "FLUID_RUNTIME_COMPONENT_TYPE",
			Value: string(componentType),
		},
	}
}

func (e *CacheEngine) transformTargetPathVolumes() *common.TargetPathVolumeConfig {
	volumeName := e.getRuntimeTargetPathVolumeName()
	targetPath := e.getTargetPath()
	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
	mountPropagation := corev1.MountPropagationBidirectional
	return &common.TargetPathVolumeConfig{
		TargetPath: targetPath,
		TargetPathHostVolume: corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: targetPath,
					Type: &hostPathDirectoryOrCreate,
				},
			},
		},
		TargetPathVolumeMount: corev1.VolumeMount{
			Name:             volumeName,
			MountPath:        targetPath,
			MountPropagation: &mountPropagation,
		},
	}
}

func (e *CacheEngine) transformRuntimeConfigVolume(value *common.CacheRuntimeComponentCommonConfig) {
	volumeName := e.getRuntimeConfigVolumeName()
	value.RuntimeConfigConfig = &common.RuntimeConfigVolumeConfig{
		RuntimeConfigPath: e.getRuntimeConfigPath(),
		RuntimeConfigVolume: corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: e.getRuntimeConfigCmName(),
					},
				},
			},
		},
		RuntimeConfigVolumeMount: corev1.VolumeMount{
			Name:      volumeName,
			MountPath: e.getRuntimeConfigDir(),
			ReadOnly:  true,
		}}
}

func (t *CacheEngine) transformTieredStore(Levels []datav1alpha1.CacheRuntimeTieredStoreLevel) (*common.CacheRuntimeComponentTieredStoreConfig, error) {
	baseIndex := 0
	tieredStoreVolumes := make([]corev1.Volume, 0)
	tieredStoreVolumeMounts := make([]corev1.VolumeMount, 0)
	cacheOpts := make([]common.TieredStoreOption, 0)

	for _, level := range Levels {
		cacheDirs := strings.Split(level.Path, ",") // mount path
		cacheQuotas := strings.Split(level.Quota, ",")

		if len(cacheQuotas) != len(cacheDirs) {
			return nil, fmt.Errorf("differene content length of quotas and paths")
		}

		for i := 0; i < len(cacheDirs); i++ {
			cacheDir := cacheDirs[i]
			cacheQuota := cacheQuotas[i]
			memQuantityRequirement := resource.MustParse("0Gi")
			if level.Medium.ProcessMemory != nil {
				memQuantityRequirement = resource.MustParse(cacheQuota)
			} else if volume := level.Medium.Volume; volume != nil {
				volumeName := fmt.Sprintf("fluid-cache-dir-%v", baseIndex)
				v := corev1.Volume{
					Name: volumeName,
				}

				if volume.HostPath != nil {
					v.VolumeSource.HostPath = volume.HostPath
				} else if volume.EmptyDir != nil {
					memQuantityRequirement = resource.MustParse(level.Quota)
					if volume.EmptyDir.Medium == corev1.StorageMediumMemory {
						volume.EmptyDir.SizeLimit = &memQuantityRequirement
					}
					v.VolumeSource.EmptyDir = volume.EmptyDir
				} else if volume.Ephemeral != nil {
					v.VolumeSource.Ephemeral = volume.Ephemeral
				}
				tieredStoreVolumes = append(tieredStoreVolumes, v)
				tieredStoreVolumeMounts = append(tieredStoreVolumeMounts, corev1.VolumeMount{
					Name:      volumeName,
					MountPath: cacheDir,
				})
			}

			cacheOpt := common.TieredStoreOption{
				CacheDir:               cacheDir,
				CacheCapacity:          cacheQuota,
				Low:                    level.Low,
				High:                   level.High,
				MemQuantityRequirement: &memQuantityRequirement,
			}

			cacheOpts = append(cacheOpts, cacheOpt)
			baseIndex++
		}
	}

	return &common.CacheRuntimeComponentTieredStoreConfig{
		CacheVolumeOptions: cacheOpts,
		CacheVolumes:       tieredStoreVolumes,
		CacheVolumeMounts:  tieredStoreVolumeMounts,
	}, nil
}

func (e *CacheEngine) transformEncryptOptionVolumes(dataset *datav1alpha1.Dataset, config *common.CacheRuntimeComponentCommonConfig) {
	encryptOptVolumes := make([]corev1.Volume, 0)
	encryptOptVolumeMounts := make([]corev1.VolumeMount, 0)
	encryptOptKeyAndPath := make(map[string]string)

	for _, m := range dataset.Spec.Mounts {
		if common.IsFluidNativeScheme(m.MountPoint) {
			continue
		}
		for _, encryptOpt := range append(dataset.Spec.SharedEncryptOptions, m.EncryptOptions...) {
			secretName := encryptOpt.ValueFrom.SecretKeyRef.Name
			secretMountPath := e.getRuntimeEncryptOptionPath(secretName)
			volName := e.getRuntimeEncryptVolumeName(secretName)
			volumeToAdd := corev1.Volume{
				Name: volName,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: secretName,
					},
				},
			}
			encryptOptVolumes = utils.AppendOrOverrideVolume(encryptOptVolumes, volumeToAdd)
			volumeMountToAdd := corev1.VolumeMount{
				Name:      volName,
				ReadOnly:  true,
				MountPath: secretMountPath,
			}
			encryptOptVolumeMounts = utils.AppendOrOverrideVolumeMounts(encryptOptVolumeMounts, volumeMountToAdd)
			encryptOptKeyAndPath[encryptOpt.Name] = filepath.Join(secretMountPath, encryptOpt.ValueFrom.SecretKeyRef.Key)
		}
	}

	config.EncryptOptionConfigs = &common.EncryptOptionVolumeConfig{
		EncryptOptionConfig:       encryptOptKeyAndPath,
		EncryptOptionVolumes:      encryptOptVolumes,
		EncryptOptionVolumeMounts: encryptOptVolumeMounts,
	}
	return
}

func (t *CacheEngine) transformPlacementMode(dataset *datav1alpha1.Dataset, value *common.CacheRuntimeValue) {
	value.PlacementMode = string(dataset.Spec.PlacementMode)
	if len(value.PlacementMode) == 0 {
		value.PlacementMode = string(datav1alpha1.ExclusiveMode)
	}
}

func (t *CacheEngine) transformTolerations(dataset *datav1alpha1.Dataset, value *common.CacheRuntimeComponentCommonConfig) {
	if len(dataset.Spec.Tolerations) > 0 {
		value.Tolerations = []corev1.Toleration{}
		for _, toleration := range dataset.Spec.Tolerations {
			toleration.TolerationSeconds = nil
			value.Tolerations = append(value.Tolerations, toleration)
		}
	}
}
