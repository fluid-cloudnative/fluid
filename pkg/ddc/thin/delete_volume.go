/*
  Copyright 2022 The Fluid Authors.

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
