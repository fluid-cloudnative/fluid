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
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

func (t *ThinEngine) transformResourcesForWorker(resources corev1.ResourceRequirements, value *ThinValue) {
	if value.Worker.Resources.Requests == nil {
		value.Worker.Resources.Requests = common.ResourceList{}
	}

	if value.Worker.Resources.Limits == nil {
		value.Worker.Resources.Limits = common.ResourceList{}
	}

	if resources.Limits != nil {
		t.Log.Info("setting worker Resources limit")
		if quantity, ok := resources.Limits[corev1.ResourceCPU]; ok {
			value.Worker.Resources.Limits[corev1.ResourceCPU] = quantity.String()
		}
		if quantity, ok := resources.Limits[corev1.ResourceMemory]; ok {
			value.Worker.Resources.Limits[corev1.ResourceMemory] = quantity.String()
		}
	}

	if resources.Requests != nil {
		t.Log.Info("setting worker Resources request")
		if quantity, ok := resources.Requests[corev1.ResourceCPU]; ok {
			value.Worker.Resources.Requests[corev1.ResourceCPU] = quantity.String()
		}
		if quantity, ok := resources.Requests[corev1.ResourceMemory]; ok {
			value.Worker.Resources.Requests[corev1.ResourceMemory] = quantity.String()
		}
	}
}

func (t *ThinEngine) transformResourcesForFuse(resources corev1.ResourceRequirements, value *ThinValue) {
	if value.Fuse.Resources.Requests == nil {
		value.Fuse.Resources.Requests = common.ResourceList{}
	}

	if value.Fuse.Resources.Limits == nil {
		value.Fuse.Resources.Limits = common.ResourceList{}
	}

	// Use the following implementation to allow resource override over the value defined in ThinRuntimeProfile
	cRes := utils.TransformCoreV1ResourcesToInternalResources(resources)
	if cRes.Limits != nil {
		for k, v := range cRes.Limits {
			value.Fuse.Resources.Limits[k] = v
		}
	}

	if cRes.Requests != nil {
		for k, v := range cRes.Requests {
			value.Fuse.Resources.Requests[k] = v
		}
	}
}
