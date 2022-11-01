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
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	volumehelper "github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
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

	return t.wrapMountedPersistentVolumeClaim()
}

func (t *ThinEngine) wrapMountedPersistentVolumeClaim() (err error) {
	dataset, err := utils.GetDataset(t.Client, t.name, t.namespace)
	if err != nil {
		return err
	}

	for _, mount := range dataset.Spec.Mounts {
		if strings.HasPrefix(mount.MountPoint, common.VolumeScheme.String()) {
			pvcName := strings.TrimPrefix(mount.MountPoint, common.VolumeScheme.String())

			mountedPvc, err := kubeclient.GetPersistentVolumeClaim(t.Client, pvcName, t.namespace)
			if err != nil {
				return err
			}

			labelsToModify := common.LabelsToModify{}
			labelsToModify.Add(common.LabelAnnotationWrappedBy, t.name)
			_, err = utils.PatchLabels(t.Client, mountedPvc, labelsToModify)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
