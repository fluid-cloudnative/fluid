/*

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
	"fmt"
	"time"

	"github.com/go-logr/logr"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
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
	found, err := kubeclient.IsPersistentVolumeExist(client, pvName, common.ExpectedFluidAnnotations)
	if err != nil {
		return err
	}

	if found {
		err = kubeclient.DeletePersistentVolume(client, pvName)
		if err != nil {
			return err
		}
		retries := 500
		for i := 0; i < retries; i++ {
			found, err = kubeclient.IsPersistentVolumeExist(client, pvName, common.ExpectedFluidAnnotations)
			if err != nil {
				return err
			}

			if found {
				time.Sleep(time.Duration(2 * time.Second))
			} else {
				break
			}
		}

		if found {
			return fmt.Errorf("the PV %s is not cleaned up",
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

	found, err := kubeclient.IsPersistentVolumeClaimExist(client, runtime.GetName(), runtime.GetNamespace(), common.ExpectedFluidAnnotations)
	if err != nil {
		return err
	}

	if found {
		err = kubeclient.DeletePersistentVolumeClaim(client, runtime.GetName(), runtime.GetNamespace())
		if err != nil {
			return err
		}

		should, err := kubeclient.ShouldRemoveProtectionFinalizer(client, runtime.GetName(), runtime.GetNamespace())
		if err != nil {
			return err
		}

		// NOTE: remove finalizer after PVC was ordered to be deleted
		if should {
			log.Info("Should remove pvc-protection finalizer")
			err = kubeclient.RemoveProtectionFinalizer(client, runtime.GetName(), runtime.GetNamespace())
			if err != nil {
				log.Info("Failed to remove finalizers")
				return err
			}
		}

		found, err := kubeclient.IsPersistentVolumeClaimExist(client, runtime.GetName(), runtime.GetNamespace(), common.ExpectedFluidAnnotations)
		if err != nil {
			return err
		}

		if found {
			return fmt.Errorf("the PVC %s in ns %s is not cleaned up",
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
