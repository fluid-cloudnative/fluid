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

func (e *CacheEngine) transformMaster(dataset *datav1alpha1.Dataset, runtime *datav1alpha1.CacheRuntime, runtimeClass *datav1alpha1.CacheRuntimeClass,
	config *CacheRuntimeComponentCommonConfig, value *common.CacheRuntimeValue) error {
	// TODO: these two field both indicate Master enabled or not, should be combined into one field.
	if runtimeClass.Topology == nil || runtimeClass.Topology.Master == nil || runtime.Spec.Master.Disabled {
		value.Master = &common.CacheRuntimeComponentValue{Enabled: false}
		return nil
	}

	component := runtimeClass.Topology.Master
	value.Master = &common.CacheRuntimeComponentValue{
		Name:            GetComponentName(e.name, common.ComponentTypeMaster),
		Namespace:       e.namespace,
		Enabled:         true,
		ComponentType:   common.ComponentTypeMaster,
		WorkloadType:    component.WorkloadType,
		PodTemplateSpec: component.Template,
		Owner:           config.Owner,
		Replicas:        runtime.Spec.Master.Replicas,
	}
	if runtimeClass.Topology.Master.Service.Headless != nil {
		value.Master.Service = &common.CacheRuntimeComponentServiceConfig{
			Name: GetComponentServiceName(e.name, common.ComponentTypeMaster),
		}
	}

	err := e.addCommonConfigForComponent(config, value.Master, component)
	if err != nil {
		return err
	}

	// transform encrypt options to master volumes (default enabled for Master)
	if shouldMountSecrets(component.Dependencies.SecretMount, true) {
		e.transformEncryptOptionsToComponentVolumes(dataset, value.Master)
	}

	// TODO: transform runtime.Spec.Master, runtimeClass.Topology.Master, dataset.Spec into PodTemplateSpec

	return nil
}
