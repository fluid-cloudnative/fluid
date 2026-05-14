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
	config *CacheRuntimeComponentCommonConfig, value *common.CacheRuntimeValue) error {

	if runtimeClass.Topology == nil || runtimeClass.Topology.Client == nil || runtime.Spec.Client.Disabled {
		value.Client = &common.CacheRuntimeComponentValue{Enabled: false}
		return nil
	}

	component := runtimeClass.Topology.Client
	value.Client = &common.CacheRuntimeComponentValue{
		Name:            GetComponentName(e.name, common.ComponentTypeClient),
		Namespace:       e.namespace,
		Enabled:         true,
		ComponentType:   common.ComponentTypeClient,
		WorkloadType:    component.WorkloadType,
		PodTemplateSpec: component.Template,
		Owner:           config.Owner,
		Replicas:        1,
	}
	if runtimeClass.Topology.Client.Service.Headless != nil {
		value.Client.Service = &common.CacheRuntimeComponentServiceConfig{
			Name: GetComponentServiceName(e.name, common.ComponentTypeClient),
		}
	}

	err := e.addCommonConfigForComponent(config, value.Client, component)
	if err != nil {
		return err
	}

	// transform encrypt options to client volumes (default disabled for Client)
	if shouldMountSecrets(component.Dependencies.SecretMount, false) {
		e.transformEncryptOptionsToComponentVolumes(dataset, value.Client)
	}

	podTemplateSpec := &value.Client.PodTemplateSpec

	// TODO: transform runtime.Spec.Client, runtimeClass.Topology.Client, dataset.Spec into PodTemplateSpec

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
