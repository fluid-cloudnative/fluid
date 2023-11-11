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

package alluxio

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	volumeHelper "github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
)

// CreateVolume creates volume
func (e *AlluxioEngine) CreateVolume() (err error) {
	if e.runtime == nil {
		e.runtime, err = e.getRuntime()
		if err != nil {
			return
		}
	}

	err = e.createFusePersistentVolume()
	if err != nil {
		return err
	}

	err = e.createFusePersistentVolumeClaim()
	if err != nil {
		return err
	}

	err = e.createHCFSPersistentVolume()
	if err != nil {
		return err
	}

	return nil

}

// createFusePersistentVolume
func (e *AlluxioEngine) createFusePersistentVolume() (err error) {
	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumeHelper.CreatePersistentVolumeForRuntime(e.Client,
		runtimeInfo,
		e.getMountPoint(),
		common.AlluxioMountType,
		e.Log)

}

// createFusePersistentVolume
func (e *AlluxioEngine) createFusePersistentVolumeClaim() (err error) {
	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumeHelper.CreatePersistentVolumeClaimForRuntime(e.Client, runtimeInfo, e.Log)

}

// createHCFSVolume (TODO: cheyang)
func (e *AlluxioEngine) createHCFSPersistentVolume() (err error) {
	return nil
}

// createHCFSVolume (TODO: cheyang)
// func (e *AlluxioEngine) createHCFSPersistentVolumeClaim() (err error) {
// 	return nil
// }
