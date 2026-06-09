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
	// config maps mounted by all component pods
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

	runtimeValue := &common.CacheRuntimeValue{}

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

// getRuntimeStatusValue extracts minimal component status information from runtimeClass and runtime spec
// This is a lightweight alternative to transform() when only status update is needed
func (e *CacheEngine) getRuntimeStatusValue(runtime *datav1alpha1.CacheRuntime, runtimeClass *datav1alpha1.CacheRuntimeClass) (*common.CacheRuntimeStatusValue, error) {
	if runtimeClass.Topology == nil ||
		(runtimeClass.Topology.Master == nil && runtimeClass.Topology.Worker == nil && runtimeClass.Topology.Client == nil) {
		return nil, fmt.Errorf("at least one component should be defined in runtimeClass")
	}

	statusValue := &common.CacheRuntimeStatusValue{}

	// Extract Master status info
	if runtimeClass.Topology.Master != nil && !runtime.Spec.Master.Disabled {
		statusValue.Master = &common.ComponentStatusInfo{
			ComponentIdentity: common.ComponentIdentity{
				Name:      GetComponentName(e.name, common.ComponentTypeMaster),
				Namespace: e.namespace,
			},
			Enabled: true,
		}
	} else {
		statusValue.Master = &common.ComponentStatusInfo{Enabled: false}
	}

	// Extract Worker status info
	if runtimeClass.Topology.Worker != nil && !runtime.Spec.Worker.Disabled {
		statusValue.Worker = &common.ComponentStatusInfo{
			ComponentIdentity: common.ComponentIdentity{
				Name:      GetComponentName(e.name, common.ComponentTypeWorker),
				Namespace: e.namespace,
			},
			Enabled: true,
		}
	} else {
		statusValue.Worker = &common.ComponentStatusInfo{Enabled: false}
	}

	// Extract Client status info
	if runtimeClass.Topology.Client != nil && !runtime.Spec.Client.Disabled {
		statusValue.Client = &common.ComponentStatusInfo{
			ComponentIdentity: common.ComponentIdentity{
				Name:      GetComponentName(e.name, common.ComponentTypeClient),
				Namespace: e.namespace,
			},
			Enabled: true,
		}
	} else {
		statusValue.Client = &common.ComponentStatusInfo{Enabled: false}
	}

	return statusValue, nil
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
