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
	"github.com/pkg/errors"
)

func (t ThinEngine) DeleteVolume() (err error) {
	if t.runtime == nil {
		t.runtime, err = t.getRuntime()
		if err != nil {
			return
		}
	}

	err = t.deleteFusePersistentVolumeClaim()
	if err != nil {
		return
	}

	err = t.deleteFusePersistentVolume()
	if err != nil {
		return
	}

	return
}

// deleteFusePersistentVolume
func (t *ThinEngine) deleteFusePersistentVolume() (err error) {
	runtimeInfo, err := t.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumehelper.DeleteFusePersistentVolume(t.Client, runtimeInfo, t.Log)
}

// deleteFusePersistentVolume
func (t *ThinEngine) deleteFusePersistentVolumeClaim() (err error) {
	runtimeInfo, err := t.getRuntimeInfo()
	if err != nil {
		return err
	}

	err = t.unwrapMountedPersistentVolumeClaims()
	if err != nil {
		return errors.Wrapf(err, "failed to unwrap pvcs for runtime %s", t.name)
	}

	return volumehelper.DeleteFusePersistentVolumeClaim(t.Client, runtimeInfo, t.Log)
}
