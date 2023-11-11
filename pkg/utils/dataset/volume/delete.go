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

package volume

import (
	"fmt"
	"time"

	"github.com/go-logr/logr"

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
	found, err := kubeclient.IsPersistentVolumeExist(client, pvName, common.ExpectedFluidAnnotations)
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
			found, err = kubeclient.IsPersistentVolumeExist(client, pvName, common.ExpectedFluidAnnotations)
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

	found, err := kubeclient.IsPersistentVolumeClaimExist(client, runtime.GetName(), runtime.GetNamespace(), common.ExpectedFluidAnnotations)
	if err != nil {
		return err
	}

	if found {
		err = kubeclient.DeletePersistentVolumeClaim(client, runtime.GetName(), runtime.GetNamespace())
		if err != nil {
			return err
		}

		stillFound := false
		retries := 10
		for i := 0; i < retries; i++ {
			stillFound, err = kubeclient.IsPersistentVolumeClaimExist(client, runtime.GetName(), runtime.GetNamespace(), common.ExpectedFluidAnnotations)
			if err != nil {
				return err
			}

			if !stillFound {
				break
			}

			should, err := kubeclient.ShouldRemoveProtectionFinalizer(client, runtime.GetName(), runtime.GetNamespace())
			if err != nil {
				// ignore NotFound error and re-check existence if the pvc is already deleted
				if utils.IgnoreNotFound(err) == nil {
					continue
				}
			}

			if should {
				log.Info("Should forcibly remove pvc-protection finalizer")
				err = kubeclient.RemoveProtectionFinalizer(client, runtime.GetName(), runtime.GetNamespace())
				if err != nil {
					// ignore NotFound error and re-check existence if the pvc is already deleted
					if utils.IgnoreNotFound(err) == nil {
						continue
					}
					log.Info("Failed to remove finalizers", "name", runtime.GetName(), "namespace", runtime.GetNamespace())
					return err
				}
			}

			time.Sleep(1 * time.Second)
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
