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
	"encoding/json"
	"errors"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

func (t *ThinEngine) transformConfig(runtime *datav1alpha1.ThinRuntime,
	dataset *datav1alpha1.Dataset, targetPath string) (config Config, err error) {
	mounts := []datav1alpha1.Mount{}
	pvAttributes := map[string]PVAttributes{}
	for _, m := range dataset.Spec.Mounts {
		if strings.HasPrefix(m.MountPoint, common.VolumeScheme.String()) {
			pvcName := strings.TrimPrefix(m.MountPoint, common.VolumeScheme.String())
			fsType, volumeAttrs, err := t.extractVolumeAttributes(pvcName)
			if err != nil {
				return config, err
			}

			pvAttributes[pvcName] = PVAttributes{
				FsType: fsType,
				VolumeAttributes: volumeAttrs,
			}
		}

		m.Options, err = t.genUFSMountOptions(m)
		if err != nil {
			return
		}
		// clean up the EncryptOptions in config.json
		m.EncryptOptions = nil
		mounts = append(mounts, m)
	}

	config.Mounts = mounts
	config.RuntimeOptions = runtime.Spec.Fuse.Options
	config.TargetPath = targetPath
	config.PersistentVolumeAttrs = pvAttributes
	return
}

func (t *ThinEngine) extractVolumeAttributes(pvcName string) (fsType string, volumeAttr map[string]string, err error) {
	pvc, err := kubeclient.GetPersistentVolumeClaim(t.Client, pvcName, t.namespace)
	if err != nil {
		return
	}

	if len(pvc.Spec.VolumeName) == 0 {
		err = errors.New(pvcName + " Not Bounded yet")
		return
	}

	pv, err := kubeclient.GetPersistentVolume(t.Client, pvc.Spec.VolumeName)
	if err != nil {
		return
	}

	fsType = pv.Spec.CSI.FSType
	volumeAttr = pv.Spec.CSI.VolumeAttributes

	return
}

func (t *ThinEngine) toRuntimeSetConfig(workers []string, fuses []string) (result string, err error) {
	if workers == nil {
		workers = []string{}
	}

	if fuses == nil {
		fuses = []string{}
	}

	status := RuntimeSetConfig{
		Workers: workers,
		Fuses:   fuses,
	}
	var runtimeStr []byte
	runtimeStr, err = json.Marshal(status)
	if err != nil {
		return
	}
	return string(runtimeStr), err
}
