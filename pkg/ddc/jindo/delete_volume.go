package jindo

import (
	"fmt"
	"time"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

func (e *JindoEngine) DeleteVolume() (err error) {
	if e.runtime == nil {
		e.runtime, err = e.getRuntime()
		if err != nil {
			return
		}
	}

	err = e.deleteFusePersistentVolumeClaim()
	if err != nil {
		return
	}

	err = e.deleteFusePersistentVolume()
	if err != nil {
		return
	}

	return
}

// deleteFusePersistentVolume
func (e *JindoEngine) deleteFusePersistentVolume() (err error) {

	found, err := kubeclient.IsPersistentVolumeExist(e.Client, e.runtime.Name, common.ExpectedFluidAnnotations)
	if err != nil {
		return err
	}

	if found {
		err = kubeclient.DeletePersistentVolume(e.Client, e.runtime.Name)
		if err != nil {
			return err
		}
		retries := 500
		for i := 0; i < retries; i++ {
			found, err = kubeclient.IsPersistentVolumeExist(e.Client, e.runtime.Name, common.ExpectedFluidAnnotations)
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
				e.runtime.Name)
		} else {
			e.Log.Info("the PV is deleted successfully", "name", e.runtime.Name)
		}
	}

	return err
}

// deleteFusePersistentVolume
func (e *JindoEngine) deleteFusePersistentVolumeClaim() (err error) {

	found, err := kubeclient.IsPersistentVolumeClaimExist(e.Client, e.runtime.Name, e.runtime.Namespace, common.ExpectedFluidAnnotations)
	if err != nil {
		return err
	}

	if found {
		err = kubeclient.DeletePersistentVolumeClaim(e.Client, e.runtime.Name, e.runtime.Namespace)
		if err != nil {
			return err
		}

		should, err := kubeclient.ShouldRemoveProtectionFinalizer(e.Client, e.runtime.Name, e.runtime.Namespace)
		if err != nil {
			return err
		}

		// NOTE: remove finalizer after PVC was ordered to be deleted
		if should {
			e.Log.Info("Should remove pvc-protection finalizer")
			err = kubeclient.RemoveProtectionFinalizer(e.Client, e.runtime.Name, e.runtime.Namespace)
			if err != nil {
				e.Log.Info("Failed to remove finalizers")
				return err
			}
		}

		found, err := kubeclient.IsPersistentVolumeClaimExist(e.Client, e.runtime.Name, e.runtime.Namespace, common.ExpectedFluidAnnotations)
		if err != nil {
			return err
		}

		if found {
			return fmt.Errorf("the PVC %s in ns %s is not cleaned up",
				e.runtime.Name,
				e.runtime.Namespace)
		} else {
			e.Log.Info("The PVC is deleted successfully",
				"name", e.runtime.Name,
				"namespace", e.runtime.Namespace)
		}
	}

	return err

}
