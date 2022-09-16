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

package thin

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

func (t *ThinEngine) transformResourcesForWorker(resources corev1.ResourceRequirements, value *ThinValue) {
	value.Worker.Resources = common.Resources{
		Requests: common.ResourceList{},
		Limits:   common.ResourceList{},
	}
	if resources.Limits != nil {
		t.Log.Info("setting worker Resources limit")
		if resources.Limits.Cpu() != nil {
			quantity := resources.Limits[corev1.ResourceCPU]
			value.Worker.Resources.Limits[corev1.ResourceCPU] = quantity.String()
		}
		if resources.Limits.Memory() != nil {
			quantity := resources.Limits[corev1.ResourceMemory]
			value.Worker.Resources.Limits[corev1.ResourceMemory] = quantity.String()
		}
	}

	if resources.Requests != nil {
		t.Log.Info("setting worker Resources request")
		if resources.Requests.Cpu() != nil {
			quantity := resources.Requests[corev1.ResourceCPU]
			value.Worker.Resources.Requests[corev1.ResourceCPU] = quantity.String()
		}
		if resources.Requests.Memory() != nil {
			quantity := resources.Requests[corev1.ResourceMemory]
			value.Worker.Resources.Requests[corev1.ResourceMemory] = quantity.String()
		}
	}
}

func (t *ThinEngine) transformResourcesForFuse(resources corev1.ResourceRequirements, value *ThinValue) {
	value.Fuse.Resources = common.Resources{
		Requests: common.ResourceList{},
		Limits:   common.ResourceList{},
	}
	if resources.Limits != nil {
		t.Log.Info("setting fuse Resources limit")
		if resources.Limits.Cpu() != nil {
			quantity := resources.Limits[corev1.ResourceCPU]
			value.Fuse.Resources.Limits[corev1.ResourceCPU] = quantity.String()
		}
		if resources.Limits.Memory() != nil {
			quantity := resources.Limits[corev1.ResourceMemory]
			value.Fuse.Resources.Limits[corev1.ResourceMemory] = quantity.String()
		}
	}

	if resources.Requests != nil {
		t.Log.Info("setting fuse Resources request")
		if resources.Requests.Cpu() != nil {
			quantity := resources.Requests[corev1.ResourceCPU]
			value.Fuse.Resources.Requests[corev1.ResourceCPU] = quantity.String()
		}
		if resources.Requests.Memory() != nil {
			quantity := resources.Requests[corev1.ResourceMemory]
			value.Fuse.Resources.Requests[corev1.ResourceMemory] = quantity.String()
		}
	}
}
