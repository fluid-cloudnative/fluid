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

package eac

import (
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

func (e *EACEngine) transformResourcesForMaster(runtime *datav1alpha1.EACRuntime, value *EAC) error {
	value.Master.Resources = common.Resources{
		Requests: common.ResourceList{},
		Limits:   common.ResourceList{},
	}
	if runtime.Spec.Master.Resources.Limits != nil {
		e.Log.Info("setting master Resources limit")
		if runtime.Spec.Master.Resources.Limits.Cpu() != nil {
			quantity := runtime.Spec.Master.Resources.Limits[corev1.ResourceCPU]
			value.Master.Resources.Limits[corev1.ResourceCPU] = quantity.String()
		}
		if runtime.Spec.Master.Resources.Limits.Memory() != nil {
			quantity := runtime.Spec.Master.Resources.Limits[corev1.ResourceMemory]
			value.Master.Resources.Limits[corev1.ResourceMemory] = quantity.String()
		}
	}

	if runtime.Spec.Master.Resources.Requests != nil {
		e.Log.Info("setting master Resources request")
		if runtime.Spec.Master.Resources.Requests.Cpu() != nil {
			quantity := runtime.Spec.Master.Resources.Requests[corev1.ResourceCPU]
			value.Master.Resources.Requests[corev1.ResourceCPU] = quantity.String()
		}
		if runtime.Spec.Master.Resources.Requests.Memory() != nil {
			quantity := runtime.Spec.Master.Resources.Requests[corev1.ResourceMemory]
			value.Master.Resources.Requests[corev1.ResourceMemory] = quantity.String()
		}
	}

	//cacheQuota := resource.MustParse(strings.TrimSuffix(value.Master.TieredStore.Levels[0].Quota, "B"))
	//needUpdated := false
	//if runtime.Spec.Master.Resources.Requests == nil ||
	//	runtime.Spec.Master.Resources.Requests.Memory() == nil ||
	//	runtime.Spec.Master.Resources.Requests.Memory().IsZero() ||
	//	cacheQuota.Cmp(*runtime.Spec.Master.Resources.Requests.Memory()) > 0 {
	//	needUpdated = true
	//}
	//
	//if runtime.Spec.Master.Resources.Limits != nil &&
	//	runtime.Spec.Master.Resources.Limits.Memory() != nil &&
	//	!runtime.Spec.Master.Resources.Limits.Memory().IsZero() &&
	//	cacheQuota.Cmp(*runtime.Spec.Master.Resources.Limits.Memory()) > 0 {
	//	return fmt.Errorf("the master memory tierdStore's size %v is greater than master limits memory %v",
	//		cacheQuota, runtime.Spec.Master.Resources.Limits.Memory())
	//}
	//
	//if needUpdated {
	//	value.Master.Resources.Requests[corev1.ResourceMemory] = cacheQuota.String()
	//}

	return nil
}

func (e *EACEngine) transformResourcesForFuse(runtime *datav1alpha1.EACRuntime, value *EAC) error {
	value.Fuse.Resources = common.Resources{
		Requests: common.ResourceList{},
		Limits:   common.ResourceList{},
	}
	if runtime.Spec.Fuse.Resources.Limits != nil {
		e.Log.Info("setting fuse Resources limit")
		if runtime.Spec.Fuse.Resources.Limits.Cpu() != nil {
			quantity := runtime.Spec.Fuse.Resources.Limits[corev1.ResourceCPU]
			value.Fuse.Resources.Limits[corev1.ResourceCPU] = quantity.String()
		}
		if runtime.Spec.Fuse.Resources.Limits.Memory() != nil {
			quantity := runtime.Spec.Fuse.Resources.Limits[corev1.ResourceMemory]
			value.Fuse.Resources.Limits[corev1.ResourceMemory] = quantity.String()
		}
	}

	if runtime.Spec.Fuse.Resources.Requests != nil {
		e.Log.Info("setting fuse Resources request")
		if runtime.Spec.Fuse.Resources.Requests.Cpu() != nil {
			quantity := runtime.Spec.Fuse.Resources.Requests[corev1.ResourceCPU]
			value.Fuse.Resources.Requests[corev1.ResourceCPU] = quantity.String()
		}
		if runtime.Spec.Fuse.Resources.Requests.Memory() != nil {
			quantity := runtime.Spec.Fuse.Resources.Requests[corev1.ResourceMemory]
			value.Fuse.Resources.Requests[corev1.ResourceMemory] = quantity.String()
		}
	}

	//cacheQuota := resource.MustParse(strings.TrimSuffix(value.Fuse.TieredStore.Levels[0].Quota, "B"))
	//needUpdated := false
	//if runtime.Spec.Fuse.Resources.Requests == nil ||
	//	runtime.Spec.Fuse.Resources.Requests.Memory() == nil ||
	//	runtime.Spec.Fuse.Resources.Requests.Memory().IsZero() ||
	//	cacheQuota.Cmp(*runtime.Spec.Fuse.Resources.Requests.Memory()) > 0 {
	//	needUpdated = true
	//}
	//
	//if runtime.Spec.Fuse.Resources.Limits != nil &&
	//	runtime.Spec.Fuse.Resources.Limits.Memory() != nil &&
	//	!runtime.Spec.Fuse.Resources.Limits.Memory().IsZero() &&
	//	cacheQuota.Cmp(*runtime.Spec.Fuse.Resources.Limits.Memory()) > 0 {
	//	return fmt.Errorf("the fuse memory tierdStore's size %v is greater than master limits memory %v",
	//		cacheQuota, runtime.Spec.Fuse.Resources.Limits.Memory())
	//}
	//
	//if needUpdated {
	//	value.Fuse.Resources.Requests[corev1.ResourceMemory] = cacheQuota.String()
	//}

	return nil
}

func (e *EACEngine) transformResourcesForWorker(runtime *datav1alpha1.EACRuntime, value *EAC) error {
	value.Worker.Resources = common.Resources{
		Requests: common.ResourceList{},
		Limits:   common.ResourceList{},
	}
	if runtime.Spec.Worker.Resources.Limits != nil {
		e.Log.Info("setting worker Resources limit")
		if runtime.Spec.Worker.Resources.Limits.Cpu() != nil {
			quantity := runtime.Spec.Worker.Resources.Limits[corev1.ResourceCPU]
			value.Worker.Resources.Limits[corev1.ResourceCPU] = quantity.String()
		}
		if runtime.Spec.Worker.Resources.Limits.Memory() != nil {
			quantity := runtime.Spec.Worker.Resources.Limits[corev1.ResourceMemory]
			value.Worker.Resources.Limits[corev1.ResourceMemory] = quantity.String()
		}
	}

	if runtime.Spec.Worker.Resources.Requests != nil {
		e.Log.Info("setting worker Resources request")
		if runtime.Spec.Worker.Resources.Requests.Cpu() != nil {
			quantity := runtime.Spec.Worker.Resources.Requests[corev1.ResourceCPU]
			value.Worker.Resources.Requests[corev1.ResourceCPU] = quantity.String()
		}
		if runtime.Spec.Worker.Resources.Requests.Memory() != nil {
			quantity := runtime.Spec.Worker.Resources.Requests[corev1.ResourceMemory]
			value.Worker.Resources.Requests[corev1.ResourceMemory] = quantity.String()
		}
	}

	cacheQuota := value.getTiredStoreLevel0Quota()
	needUpdated := false
	if runtime.Spec.Worker.Resources.Requests == nil ||
		runtime.Spec.Worker.Resources.Requests.Memory() == nil ||
		runtime.Spec.Worker.Resources.Requests.Memory().IsZero() ||
		cacheQuota.Cmp(*runtime.Spec.Worker.Resources.Requests.Memory()) > 0 {
		needUpdated = true
	}

	if runtime.Spec.Worker.Resources.Limits != nil &&
		runtime.Spec.Worker.Resources.Limits.Memory() != nil &&
		!runtime.Spec.Worker.Resources.Limits.Memory().IsZero() &&
		cacheQuota.Cmp(*runtime.Spec.Worker.Resources.Limits.Memory()) > 0 {
		return fmt.Errorf("the worker memory tierdStore's size %v is greater than master limits memory %v",
			cacheQuota, runtime.Spec.Worker.Resources.Limits.Memory())
	}

	if needUpdated {
		value.Worker.Resources.Requests[corev1.ResourceMemory] = cacheQuota.String()
	}

	return nil
}
