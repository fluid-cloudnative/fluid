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
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DeleteFusePersistentVolume
func DeleteFusePersistentVolume(client client.Client,
	runtime base.RuntimeInfoInterface,
	log logr.Logger) (err error) {

	pvName := runtime.GetPersistentVolumeName()

	err = deleteFusePersistentVolumeIfExists(client, pvName, log)
	if err != nil {
		return err
	}

	return err
}

func deleteFusePersistentVolumeIfExists(client client.Client, pvName string, log logr.Logger) (err error) {
	found, err := kubeclient.IsPersistentVolumeExist(client, pvName, common.GetExpectedFluidAnnotations())
	if err != nil {
		return err
	}

	if found {
		err = kubeclient.DeletePersistentVolume(client, pvName)
		if err != nil {
			return err
		}
		retries := 10
		for i := 0; i < retries; i++ {
			found, err = kubeclient.IsPersistentVolumeExist(client, pvName, common.GetExpectedFluidAnnotations())
			if err != nil {
				return err
			}

			if found {
				time.Sleep(1 * time.Second)
			} else {
				break
			}
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
func DeleteFusePersistentVolumeClaim(client client.Client,
	runtime base.RuntimeInfoInterface,
	log logr.Logger) (err error) {

	found, err := kubeclient.IsPersistentVolumeClaimExist(client, runtime.GetName(), runtime.GetNamespace(), common.GetExpectedFluidAnnotations())
	if err != nil {
		return err
	}

	if found {
		err = kubeclient.DeletePersistentVolumeClaim(client, runtime.GetName(), runtime.GetNamespace())
		if err != nil {
			return err
		}

		stillFound := false
		ctx, cancelFunc := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancelFunc()

		backoff := wait.Backoff{Duration: 100 * time.Millisecond, Steps: 10, Jitter: 0.2}
		err := wait.ExponentialBackoffWithContext(ctx, backoff, func(ctx context.Context) (done bool, err error) {
			stillFound, err = kubeclient.IsPersistentVolumeClaimExist(client, runtime.GetName(), runtime.GetNamespace(), common.GetExpectedFluidAnnotations())
			if err != nil {
				return false, err
			}

			if !stillFound {
				return true, nil
			}

			// WARN: This is a LEGACY MECHANISM and will be removed in the future.
			// force deletion of pvc-protection finalizer will not be done, we'll wait until the pvc is really deleted by the PV controller.
			if utils.GetBoolValueFromEnv(common.LegacyEnvForceCleanUpManagedPVC, false) {
				should, err := kubeclient.ShouldRemoveProtectionFinalizer(client, runtime.GetName(), runtime.GetNamespace())
				if err != nil {
					// ignore NotFound error and re-check existence if the pvc is already deleted
					if utils.IgnoreNotFound(err) == nil {
						return false, nil
					}
				}

				if should {
					log.Info("Should forcibly remove pvc-protection finalizer")
					err = kubeclient.RemoveProtectionFinalizer(client, runtime.GetName(), runtime.GetNamespace())
					if err != nil {
						// ignore NotFound error and re-check existence if the pvc is already deleted
						if utils.IgnoreNotFound(err) == nil {
							return false, nil
						}
						log.Info("Failed to remove finalizers", "name", runtime.GetName(), "namespace", runtime.GetNamespace())
						return false, err
					}
				}
			}

			return false, nil
		})

		if err != nil {
			if wait.Interrupted(err) {
				return errors.Wrapf(err, "timeout waiting for PVC %s to be deleted after 1-second retry", runtime.GetName())
			}
			return err
		}

		log.Info("The PVC is deleted successfully",
			"name", runtime.GetName(),
			"namespace", runtime.GetNamespace())
	}

	return nil
}
