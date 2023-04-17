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
	"context"
	"fmt"
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
)

func (e *EACEngine) transformResourcesForMaster(runtime *datav1alpha1.EFCRuntime, value *EAC) error {
	value.Master.Resources = common.Resources{
		Requests: common.ResourceList{},
		Limits:   common.ResourceList{},
	}
	if len(runtime.Spec.Master.Resources.Limits) > 0 || len(runtime.Spec.Master.Resources.Requests) > 0 {
		value.Master.Resources = utils.TransformRequirementsToResources(runtime.Spec.Master.Resources)
	}

	if len(value.Master.TieredStore.Levels) > 0 &&
		len(value.Master.TieredStore.Levels[0].Quota) > 0 &&
		value.Master.TieredStore.Levels[0].MediumType == string(common.Memory) {
		cacheQuota := utils.TransformEACUnitToQuantity(value.Master.TieredStore.Levels[0].Quota)
		needUpdated := false
		if runtime.Spec.Master.Resources.Requests == nil ||
			runtime.Spec.Master.Resources.Requests.Memory() == nil ||
			runtime.Spec.Master.Resources.Requests.Memory().IsZero() ||
			cacheQuota.Cmp(*runtime.Spec.Master.Resources.Requests.Memory()) > 0 {
			needUpdated = true
		}

		if runtime.Spec.Master.Resources.Limits != nil &&
			runtime.Spec.Master.Resources.Limits.Memory() != nil &&
			!runtime.Spec.Master.Resources.Limits.Memory().IsZero() &&
			cacheQuota.Cmp(*runtime.Spec.Master.Resources.Limits.Memory()) > 0 {
			return fmt.Errorf("the master memory tierdStore's size %v is greater than master limits memory %v",
				cacheQuota, runtime.Spec.Master.Resources.Limits.Memory())
		}

		if needUpdated {
			value.Master.Resources.Requests[corev1.ResourceMemory] = cacheQuota.String()

			err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				runtime, err := e.getRuntime()
				if err != nil {
					return err
				}
				runtimeToUpdate := runtime.DeepCopy()
				if len(runtimeToUpdate.Spec.Master.Resources.Requests) == 0 {
					runtimeToUpdate.Spec.Master.Resources.Requests = make(corev1.ResourceList)
				}
				runtimeToUpdate.Spec.Master.Resources.Requests[corev1.ResourceMemory] = *cacheQuota
				if !reflect.DeepEqual(runtimeToUpdate, runtime) {
					err = e.Client.Update(context.TODO(), runtimeToUpdate)
					if err != nil {
						return err
					}
				}
				return nil
			})

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *EACEngine) transformResourcesForFuse(runtime *datav1alpha1.EFCRuntime, value *EAC) error {
	value.Fuse.Resources = common.Resources{
		Requests: common.ResourceList{},
		Limits:   common.ResourceList{},
	}
	if len(runtime.Spec.Fuse.Resources.Limits) > 0 || len(runtime.Spec.Fuse.Resources.Requests) > 0 {
		value.Fuse.Resources = utils.TransformRequirementsToResources(runtime.Spec.Fuse.Resources)
	}

	if len(value.Fuse.TieredStore.Levels) > 0 &&
		len(value.Fuse.TieredStore.Levels[0].Quota) > 0 &&
		value.Fuse.TieredStore.Levels[0].MediumType == string(common.Memory) {
		cacheQuota := utils.TransformEACUnitToQuantity(value.Fuse.TieredStore.Levels[0].Quota)
		needUpdated := false
		if runtime.Spec.Fuse.Resources.Requests == nil ||
			runtime.Spec.Fuse.Resources.Requests.Memory() == nil ||
			runtime.Spec.Fuse.Resources.Requests.Memory().IsZero() ||
			cacheQuota.Cmp(*runtime.Spec.Fuse.Resources.Requests.Memory()) > 0 {
			needUpdated = true
		}

		if runtime.Spec.Fuse.Resources.Limits != nil &&
			runtime.Spec.Fuse.Resources.Limits.Memory() != nil &&
			!runtime.Spec.Fuse.Resources.Limits.Memory().IsZero() &&
			cacheQuota.Cmp(*runtime.Spec.Fuse.Resources.Limits.Memory()) > 0 {
			return fmt.Errorf("the fuse memory tierdStore's size %v is greater than fuse limits memory %v",
				cacheQuota, runtime.Spec.Fuse.Resources.Limits.Memory())
		}

		if needUpdated {
			value.Fuse.Resources.Requests[corev1.ResourceMemory] = cacheQuota.String()

			err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				runtime, err := e.getRuntime()
				if err != nil {
					return err
				}
				runtimeToUpdate := runtime.DeepCopy()
				if len(runtimeToUpdate.Spec.Fuse.Resources.Requests) == 0 {
					runtimeToUpdate.Spec.Fuse.Resources.Requests = make(corev1.ResourceList)
				}
				runtimeToUpdate.Spec.Fuse.Resources.Requests[corev1.ResourceMemory] = *cacheQuota
				if !reflect.DeepEqual(runtimeToUpdate, runtime) {
					err = e.Client.Update(context.TODO(), runtimeToUpdate)
					if err != nil {
						return err
					}
				}
				return nil
			})

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *EACEngine) transformResourcesForWorker(runtime *datav1alpha1.EFCRuntime, value *EAC) error {
	value.Worker.Resources = common.Resources{
		Requests: common.ResourceList{},
		Limits:   common.ResourceList{},
	}
	if len(runtime.Spec.Worker.Resources.Limits) > 0 || len(runtime.Spec.Worker.Resources.Requests) > 0 {
		value.Worker.Resources = utils.TransformRequirementsToResources(runtime.Spec.Worker.Resources)
	}

	if len(value.Worker.TieredStore.Levels) > 0 &&
		len(value.getTiredStoreLevel0Quota()) > 0 &&
		value.getTiredStoreLevel0MediumType() == string(common.Memory) {
		cacheQuota := utils.TransformEACUnitToQuantity(value.getTiredStoreLevel0Quota())
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
			return fmt.Errorf("the worker memory tierdStore's size %v is greater than worker limits memory %v",
				cacheQuota, runtime.Spec.Worker.Resources.Limits.Memory())
		}

		if needUpdated {
			value.Worker.Resources.Requests[corev1.ResourceMemory] = cacheQuota.String()

			err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				runtime, err := e.getRuntime()
				if err != nil {
					return err
				}
				runtimeToUpdate := runtime.DeepCopy()
				if len(runtimeToUpdate.Spec.Worker.Resources.Requests) == 0 {
					runtimeToUpdate.Spec.Worker.Resources.Requests = make(corev1.ResourceList)
				}
				runtimeToUpdate.Spec.Worker.Resources.Requests[corev1.ResourceMemory] = *cacheQuota
				if !reflect.DeepEqual(runtimeToUpdate, runtime) {
					err = e.Client.Update(context.TODO(), runtimeToUpdate)
					if err != nil {
						return err
					}
				}
				return nil
			})

			if err != nil {
				return err
			}
		}
	}

	return nil
}
