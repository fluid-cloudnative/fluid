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

	found, err := kubeclient.IsPersistentVolumeExist(client, runtime.GetName(), common.ExpectedFluidAnnotations)
	if err != nil {
		return err
	}

	if found {
		err = kubeclient.DeletePersistentVolume(client, runtime.GetName())
		if err != nil {
			return err
		}
		retries := 500
		for i := 0; i < retries; i++ {
			found, err = kubeclient.IsPersistentVolumeExist(client, runtime.GetName(), common.ExpectedFluidAnnotations)
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
				runtime.GetName())
		} else {
			log.Info("the PV is deleted successfully", "name", runtime.GetName())
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
