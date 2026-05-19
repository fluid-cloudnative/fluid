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
	commonConfig *CacheRuntimeComponentCommonConfig, value *common.CacheRuntimeValue) error {
	runtimeMaster := runtime.Spec.Master
	// these two field (runtimeClass.Topology.Master and  runtimeMaster.Disabled） both indicate Master enabled or not.
	if runtimeClass.Topology == nil || runtimeClass.Topology.Master == nil || runtimeMaster.Disabled {
		value.Master = &common.CacheRuntimeComponentValue{Enabled: false}
		return nil
	}
	componentDefinition := runtimeClass.Topology.Master

	// Initialize component value with common fields
	var err error
	value.Master, err = e.initComponentValue(common.ComponentTypeMaster, componentDefinition, commonConfig.Owner, runtimeMaster.Replicas)
	if err != nil {
		return err
	}

	// TODO: TieredStore handling

	// transform container related config, currently only modify the first container
	e.transformComponentPodTemplate(runtimeMaster.RuntimeComponentCommonSpec, dataset, value.Master)

	// transform all volume-related configurations
	err = e.transformVolumes(runtime.Spec.Volumes, runtime.Spec.Master.VolumeMounts, dataset, componentDefinition, commonConfig, true, &value.Master.PodTemplateSpec.Spec)

	if err != nil {
		return err
	}

	return nil
}
