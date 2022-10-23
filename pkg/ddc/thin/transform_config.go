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
	"fmt"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
)

func (t *ThinEngine) transformConfig(runtime *datav1alpha1.ThinRuntime,
	dataset *datav1alpha1.Dataset, targetPath string) (config Config, err error) {
	mounts := []datav1alpha1.Mount{}
	pvAttributes := map[string]*corev1.CSIPersistentVolumeSource{}
	for _, m := range dataset.Spec.Mounts {
		if strings.HasPrefix(m.MountPoint, common.VolumeScheme.String()) {
			pvcName := strings.TrimPrefix(m.MountPoint, common.VolumeScheme.String())
			csiInfo, err := t.extractCSIVolumeSourceInfo(pvcName)
			if err != nil {
				return config, err
			}

			pvAttributes[pvcName] = csiInfo
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

func (t *ThinEngine) extractCSIVolumeSourceInfo(pvcName string) (csiInfo *corev1.CSIPersistentVolumeSource, err error) {
	pvc, err := kubeclient.GetPersistentVolumeClaim(t.Client, pvcName, t.namespace)
	if err != nil {
		return
	}

	if len(pvc.Spec.VolumeName) == 0 {
		err = fmt.Errorf("persistent volume claim %s not bounded yet", pvcName)
		return
	}

	pv, err := kubeclient.GetPersistentVolume(t.Client, pvc.Spec.VolumeName)
	if err != nil {
		return
	}

	if pv.Spec.CSI == nil {
		err = fmt.Errorf("persistent volume %s has unsupported volume source. only CSI is supported", pv.Name)
		return
	}

	return pv.Spec.CSI, nil
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
