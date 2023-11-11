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

package thin

import (
	volumehelper "github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
)

func (t ThinEngine) CreateVolume() (err error) {
	if t.runtime == nil {
		t.runtime, err = t.getRuntime()
		if err != nil {
			return
		}
	}

	err = t.createFusePersistentVolume()
	if err != nil {
		return err
	}

	err = t.createFusePersistentVolumeClaim()
	if err != nil {
		return err
	}
	return
}

// createFusePersistentVolume
func (t *ThinEngine) createFusePersistentVolume() (err error) {
	runtimeInfo, err := t.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumehelper.CreatePersistentVolumeForRuntime(t.Client,
		runtimeInfo,
		t.getTargetPath(),
		t.runtimeProfile.Spec.FileSystemType,
		t.Log)
}

// createFusePersistentVolume
func (t *ThinEngine) createFusePersistentVolumeClaim() (err error) {
	runtimeInfo, err := t.getRuntimeInfo()
	if err != nil {
		return err
	}

	err = volumehelper.CreatePersistentVolumeClaimForRuntime(t.Client, runtimeInfo, t.Log)
	if err != nil {
		return err
	}

	// If the dataset contains pvc:// scheme mount point, set owner reference to the
	// dataset with the mounted pvc as its owner. If no pvc:// scheme mount point is specified,
	// it takes no effect.
	err = t.bindDatasetToMountedPersistentVolumeClaim()
	if err != nil {
		return err
	}

	// If the dataset contains pvc:// scheme mount point, wrap the mounted PVC, otherwise
	// it takes no effect.
	return t.wrapMountedPersistentVolumeClaim()
}
