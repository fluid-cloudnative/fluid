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

package juicefs

import volumehelper "github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"

func (j JuiceFSEngine) DeleteVolume() (err error) {
	if j.runtime == nil {
		j.runtime, err = j.getRuntime()
		if err != nil {
			return
		}
	}

	err = j.deleteFusePersistentVolumeClaim()
	if err != nil {
		return
	}

	err = j.deleteFusePersistentVolume()
	if err != nil {
		return
	}

	return
}

// deleteFusePersistentVolume
func (j *JuiceFSEngine) deleteFusePersistentVolume() (err error) {
	runtimeInfo, err := j.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumehelper.DeleteFusePersistentVolume(j.Client, runtimeInfo, j.Log)
}

// deleteFusePersistentVolume
func (j *JuiceFSEngine) deleteFusePersistentVolumeClaim() (err error) {
	runtimeInfo, err := j.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumehelper.DeleteFusePersistentVolumeClaim(j.Client, runtimeInfo, j.Log)
}
