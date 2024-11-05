/*
Copyright 2021 The Fluid Authors.

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
