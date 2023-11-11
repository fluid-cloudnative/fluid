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

package jindocache

import (
	volumeHelper "github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
)

// DeleteVolume
func (e *JindoCacheEngine) DeleteVolume() (err error) {

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
func (e *JindoCacheEngine) deleteFusePersistentVolume() (err error) {
	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumeHelper.DeleteFusePersistentVolume(e.Client, runtimeInfo, e.Log)
}

// deleteFusePersistentVolumeClaim
func (e *JindoCacheEngine) deleteFusePersistentVolumeClaim() (err error) {
	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumeHelper.DeleteFusePersistentVolumeClaim(e.Client, runtimeInfo, e.Log)
}
