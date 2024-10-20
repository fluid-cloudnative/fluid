/*
Copyright 2021 The Fluid Authors.

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
