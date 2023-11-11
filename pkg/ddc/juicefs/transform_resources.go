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

package juicefs

import (
	"context"
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

func (j *JuiceFSEngine) transformResourcesForFuse(runtime *datav1alpha1.JuiceFSRuntime, value *JuiceFS) (err error) {
	value.Fuse.Resources = common.Resources{
		Requests: common.ResourceList{},
		Limits:   common.ResourceList{},
	}
	if runtime.Spec.Fuse.Resources.Limits != nil {
		j.Log.Info("setting fuse Resources limit")
		for k, v := range runtime.Spec.Fuse.Resources.Limits {
			value.Fuse.Resources.Limits[k] = v.String()
		}
	}

	if runtime.Spec.Fuse.Resources.Requests != nil {
		j.Log.Info("setting fuse Resources request")
		for k, v := range runtime.Spec.Fuse.Resources.Requests {
			value.Fuse.Resources.Requests[k] = v.String()
		}
	}

	// mem set request
	if j.hasTieredStore(runtime) && j.getTieredStoreType(runtime) == 0 && runtime.Spec.Fuse.Options["cache-size"] == "" {
		userQuota := runtime.Spec.TieredStore.Levels[0].Quota
		if userQuota == nil {
			return
		}
		needUpdated := false
		if runtime.Spec.Fuse.Resources.Requests == nil ||
			runtime.Spec.Fuse.Resources.Requests.Memory() == nil ||
			runtime.Spec.Fuse.Resources.Requests.Memory().IsZero() ||
			userQuota.Cmp(*runtime.Spec.Fuse.Resources.Requests.Memory()) > 0 {
			needUpdated = true
		}
		if !runtime.Spec.Fuse.Resources.Limits.Memory().IsZero() &&
			userQuota.Cmp(*runtime.Spec.Fuse.Resources.Limits.Memory()) > 0 {
			return fmt.Errorf("the fuse memory tierdStore's size %v is greater than master limits memory %v",
				userQuota, runtime.Spec.Fuse.Resources.Limits.Memory())
		}

		if needUpdated {
			value.Fuse.Resources.Requests[corev1.ResourceMemory] = userQuota.String()
			err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				runtime, err := j.getRuntime()
				if err != nil {
					return err
				}
				runtimeToUpdate := runtime.DeepCopy()
				if len(runtimeToUpdate.Spec.Fuse.Resources.Requests) == 0 {
					runtimeToUpdate.Spec.Fuse.Resources.Requests = make(corev1.ResourceList)
				}
				runtimeToUpdate.Spec.Fuse.Resources.Requests[corev1.ResourceMemory] = *userQuota
				if !reflect.DeepEqual(runtimeToUpdate, runtime) {
					err = j.Client.Update(context.TODO(), runtimeToUpdate)
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
	return
}

func (j *JuiceFSEngine) transformResourcesForWorker(runtime *datav1alpha1.JuiceFSRuntime, value *JuiceFS) (err error) {
	value.Worker.Resources = common.Resources{
		Requests: common.ResourceList{},
		Limits:   common.ResourceList{},
	}
	if runtime.Spec.Worker.Resources.Limits != nil {
		j.Log.Info("setting worker Resources limit")
		for k, v := range runtime.Spec.Worker.Resources.Limits {
			value.Worker.Resources.Limits[k] = v.String()
		}
	}

	if runtime.Spec.Worker.Resources.Requests != nil {
		j.Log.Info("setting worker Resources request")
		for k, v := range runtime.Spec.Worker.Resources.Requests {
			value.Worker.Resources.Requests[k] = v.String()
		}
	}

	// mem set request in enterprise edition
	if j.hasTieredStore(runtime) && j.getTieredStoreType(runtime) == 0 && value.Edition == EnterpriseEdition && runtime.Spec.Worker.Options["cache-size"] == "" {
		userQuota := runtime.Spec.TieredStore.Levels[0].Quota
		if userQuota == nil {
			return
		}
		needUpdated := false
		if runtime.Spec.Worker.Resources.Requests == nil ||
			runtime.Spec.Worker.Resources.Requests.Memory() == nil ||
			runtime.Spec.Worker.Resources.Requests.Memory().IsZero() ||
			userQuota.Cmp(*runtime.Spec.Worker.Resources.Requests.Memory()) > 0 {
			needUpdated = true
		}
		if !runtime.Spec.Worker.Resources.Limits.Memory().IsZero() &&
			userQuota.Cmp(*runtime.Spec.Worker.Resources.Limits.Memory()) > 0 {
			return fmt.Errorf("the worker memory tierdStore's size %v is greater than master limits memory %v",
				userQuota, runtime.Spec.Worker.Resources.Limits.Memory())
		}

		if needUpdated {
			value.Worker.Resources.Requests[corev1.ResourceMemory] = userQuota.String()
			err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				runtime, err := j.getRuntime()
				if err != nil {
					return err
				}
				runtimeToUpdate := runtime.DeepCopy()
				if len(runtimeToUpdate.Spec.Worker.Resources.Requests) == 0 {
					runtimeToUpdate.Spec.Worker.Resources.Requests = make(corev1.ResourceList)
				}
				runtimeToUpdate.Spec.Worker.Resources.Requests[corev1.ResourceMemory] = *userQuota
				if !reflect.DeepEqual(runtimeToUpdate, runtime) {
					err = j.Client.Update(context.TODO(), runtimeToUpdate)
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
	return
}
