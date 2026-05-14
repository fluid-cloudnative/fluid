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
)

func (e *CacheEngine) transformWorker(dataset *datav1alpha1.Dataset, runtime *datav1alpha1.CacheRuntime, runtimeClass *datav1alpha1.CacheRuntimeClass,
	config *CacheRuntimeComponentCommonConfig, value *common.CacheRuntimeValue) error {

	if runtimeClass.Topology == nil || runtimeClass.Topology.Worker == nil || runtime.Spec.Worker.Disabled {
		value.Worker = &common.CacheRuntimeComponentValue{Enabled: false}
		return nil
	}

	component := runtimeClass.Topology.Worker
	value.Worker = &common.CacheRuntimeComponentValue{
		Name:            GetComponentName(e.name, common.ComponentTypeWorker),
		Namespace:       e.namespace,
		Enabled:         true,
		ComponentType:   common.ComponentTypeWorker,
		WorkloadType:    component.WorkloadType,
		PodTemplateSpec: component.Template,
		Owner:           config.Owner,
		Replicas:        runtime.Spec.Worker.Replicas,
	}
	if runtimeClass.Topology.Worker.Service.Headless != nil {
		value.Worker.Service = &common.CacheRuntimeComponentServiceConfig{
			Name: GetComponentServiceName(e.name, common.ComponentTypeWorker),
		}
	}

	err := e.addCommonConfigForComponent(config, value.Worker, component)
	if err != nil {
		return err
	}

	// transform encrypt options to worker volumes (default enabled for Worker)
	if shouldMountSecrets(component.Dependencies.SecretMount, true) {
		e.transformEncryptOptionsToComponentVolumes(dataset, value.Worker)
	}

	// TODO: transform runtime.Spec.Worker, runtimeClass.Topology.Worker, dataset.Spec into PodTemplateSpec

	return nil
}
