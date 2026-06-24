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
	corev1 "k8s.io/api/core/v1"
)

func (e *CacheEngine) transformClient(dataset *datav1alpha1.Dataset, runtime *datav1alpha1.CacheRuntime, runtimeClass *datav1alpha1.CacheRuntimeClass,
	commonConfig *CacheRuntimeComponentCommonConfig, value *common.CacheRuntimeValue) error {
	runtimeClient := runtime.Spec.Client

	if runtimeClass.Topology == nil || runtimeClass.Topology.Client == nil || runtimeClient.Disabled {
		value.Client = &common.CacheRuntimeComponentValue{Enabled: false}
		return nil
	}

	componentDefinition := runtimeClass.Topology.Client

	// Initialize component value with common fields (Client always has 1 replica)
	var err error
	value.Client, err = e.initComponentValue(common.ComponentTypeClient, componentDefinition, commonConfig.Owner, 1)
	if err != nil {
		return err
	}

	podTemplateSpec := &value.Client.PodTemplateSpec

	// TODO: TieredStore handling

	// transform container related config, currently only modify the first container
	e.transformComponentPodTemplate(runtimeClient.RuntimeComponentCommonSpec, dataset, value.Client)

	// transform tiered store configuration into pod resource request or volumes .
	err = e.TransformRuntimeTieredStore(&runtimeClient.TieredStore, &value.Client.PodTemplateSpec.Spec)
	if err != nil {
		return err
	}

	// transform all volume-related configurations
	// Client default does NOT mount secrets (defaultMountSecrets=false)
	err = e.transformVolumes(runtime.Spec.Volumes, runtime.Spec.Client.VolumeMounts, dataset, componentDefinition, commonConfig, false, &value.Client.PodTemplateSpec.Spec)

	if err != nil {
		return err
	}

	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	// fuse label, keep the same key/value with CSI node server
	if podTemplateSpec.Spec.NodeSelector == nil {
		podTemplateSpec.Spec.NodeSelector = map[string]string{}
	}
	podTemplateSpec.Spec.NodeSelector[runtimeInfo.GetFuseLabelName()] = "true"

	// fuse volume mount
	e.transformFuseMountPointVolumes(podTemplateSpec)

	return nil
}
func (e *CacheEngine) transformFuseMountPointVolumes(podTemplate *corev1.PodTemplateSpec) {
	volumeName := e.getFuseMountPointVolumeName()
	targetPath := e.getFuseMountPoint()
	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
	mountPropagation := corev1.MountPropagationBidirectional

	podTemplate.Spec.Volumes = append(podTemplate.Spec.Volumes, corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: targetPath,
				Type: &hostPathDirectoryOrCreate,
			},
		},
	})
	podTemplate.Spec.Containers[0].VolumeMounts = append(podTemplate.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
		Name:             volumeName,
		MountPath:        targetPath,
		MountPropagation: &mountPropagation,
	})
}
