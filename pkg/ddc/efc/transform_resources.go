/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package efc

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

func (e *EFCEngine) transformResourcesForMaster(runtime *datav1alpha1.EFCRuntime, value *EFC) error {
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
		cacheQuota := utils.TransformEFCUnitToQuantity(value.Master.TieredStore.Levels[0].Quota)
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

func (e *EFCEngine) transformResourcesForFuse(runtime *datav1alpha1.EFCRuntime, value *EFC) error {
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
		cacheQuota := utils.TransformEFCUnitToQuantity(value.Fuse.TieredStore.Levels[0].Quota)
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

func (e *EFCEngine) transformResourcesForWorker(runtime *datav1alpha1.EFCRuntime, value *EFC) error {
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
		cacheQuota := utils.TransformEFCUnitToQuantity(value.getTiredStoreLevel0Quota())
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
