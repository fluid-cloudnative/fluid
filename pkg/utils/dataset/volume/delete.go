/*
Copyright 2023 The Fluid Authors.

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

package volume

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DeleteFusePersistentVolume
func DeleteFusePersistentVolume(ctx context.Context,
	client client.Client,
	runtime base.RuntimeInfoInterface,
	log logr.Logger) (err error) {

	pvName := runtime.GetPersistentVolumeName()

	err = deleteFusePersistentVolumeIfExists(ctx, client, pvName, log)
	if err != nil {
		return err
	}

	return err
}

func deleteFusePersistentVolumeIfExists(ctx context.Context, client client.Client, pvName string, log logr.Logger) (err error) {
	found, err := kubeclient.IsPersistentVolumeExistWithContext(ctx, client, pvName, common.GetExpectedFluidAnnotations())
	if err != nil {
		return err
	}

	if found {
		err = kubeclient.DeletePersistentVolumeWithContext(ctx, client, pvName)
		if err != nil {
			return err
		}
		pollCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		pollErr := wait.PollUntilContextCancel(pollCtx, time.Second, true, func(pollCtx context.Context) (bool, error) {
			found, err = kubeclient.IsPersistentVolumeExistWithContext(pollCtx, client, pvName, common.GetExpectedFluidAnnotations())
			if err != nil {
				return false, err
			}
			return !found, nil
		})
		if pollErr != nil && ctx.Err() != nil {
			return pollErr
		}

		if found {
			return fmt.Errorf("the PV %s is not cleaned up after 10-second retry",
				pvName)
		} else {
			log.Info("the PV is deleted successfully", "name", pvName)
		}
	}

	return err
}

// DeleteFusePersistentVolume
func DeleteFusePersistentVolumeClaim(ctx context.Context,
	client client.Client,
	runtime base.RuntimeInfoInterface,
	log logr.Logger) (err error) {

	found, err := kubeclient.IsPersistentVolumeClaimExistWithContext(ctx, client, runtime.GetName(), runtime.GetNamespace(), common.GetExpectedFluidAnnotations())
	if err != nil {
		return err
	}

	if found {
		err = kubeclient.DeletePersistentVolumeClaimWithContext(ctx, client, runtime.GetName(), runtime.GetNamespace())
		if err != nil {
			return err
		}

		stillFound := false
		pollCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		pollErr := wait.PollUntilContextCancel(pollCtx, time.Second, true, func(pollCtx context.Context) (bool, error) {
			stillFound, err = kubeclient.IsPersistentVolumeClaimExistWithContext(pollCtx, client, runtime.GetName(), runtime.GetNamespace(), common.GetExpectedFluidAnnotations())
			if err != nil {
				return false, err
			}

			if !stillFound {
				return true, nil
			}

			should, err := kubeclient.ShouldRemoveProtectionFinalizerWithContext(pollCtx, client, runtime.GetName(), runtime.GetNamespace())
			if err != nil {
				// ignore NotFound error and re-check existence if the pvc is already deleted
				if utils.IgnoreNotFound(err) == nil {
					return false, nil
				}
				return false, err
			}

			if should {
				log.Info("Should forcibly remove pvc-protection finalizer")
				err = kubeclient.RemoveProtectionFinalizerWithContext(pollCtx, client, runtime.GetName(), runtime.GetNamespace())
				if err != nil {
					// ignore NotFound error and re-check existence if the pvc is already deleted
					if utils.IgnoreNotFound(err) == nil {
						return false, nil
					}
					log.Info("Failed to remove finalizers", "name", runtime.GetName(), "namespace", runtime.GetNamespace())
					return false, err
				}
			}

			return false, nil
		})
		if pollErr != nil && ctx.Err() != nil {
			return pollErr
		}

		if stillFound {
			return fmt.Errorf("the PVC %s in ns %s is not cleaned up after 10-second retry",
				runtime.GetName(),
				runtime.GetNamespace())
		} else {
			log.Info("The PVC is deleted successfully",
				"name", runtime.GetName(),
				"namespace", runtime.GetNamespace())
		}
	}

	return err

}
