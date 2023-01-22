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

package alluxio

import (
	"context"
	"reflect"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
)

// SetCleanCachePolicy set the CleanCachePolicy of AlluxioRuntime
func (e *AlluxioEngine) SetCleanCachePolicy() (err error) {
	maxRetryAttempts, err := e.getGracefulShutdownLimits()
	if err != nil {
		return
	}

	cleanCacheGracePeriodSeconds, err := e.getCleanCacheGracePeriodSeconds()
	if err != nil {
		return
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}
		runtimeToUpdate := runtime.DeepCopy()
		runtimeToUpdate.Spec.CleanCachePolicy = datav1alpha1.CleanCachePolicy{
			GracePeriodSeconds: &cleanCacheGracePeriodSeconds,
			MaxRetryAttempts:   &maxRetryAttempts,
		}

		if !reflect.DeepEqual(runtimeToUpdate, runtime) {
			err = e.Client.Update(context.TODO(), runtimeToUpdate)
			if err != nil {
				if apierrors.IsConflict(err) {
					time.Sleep(3 * time.Second)
				}
				return err
			}
			e.Log.Info("Update cleanCachePolicy successfully",
				"runtimeToUpdate",
				runtimeToUpdate)
		} else {
			e.Log.Info("No need to update runtime for CleanCachePolicy",
				"runtimeToUpdate",
				runtimeToUpdate)
		}

		return err
	})

	return err
}
