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

package goosefs

import (
	volumeHelper "github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
)

// DeleteVolume deletes the GooseFS volume by performing the following steps:
// 1. Initializes the runtime if it is not already initialized.
// 2. Deletes the Fuse Persistent Volume Claim (PVC) associated with the volume.
// 3. Deletes the Fuse Persistent Volume (PV) associated with the volume.
// Returns an error if any of the steps fail.
func (e *GooseFSEngine) DeleteVolume() (err error) {

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
func (e *GooseFSEngine) deleteFusePersistentVolume() (err error) {
	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumeHelper.DeleteFusePersistentVolume(e.Client, runtimeInfo, e.Log)
}

// deleteFusePersistentVolume
func (e *GooseFSEngine) deleteFusePersistentVolumeClaim() (err error) {
	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumeHelper.DeleteFusePersistentVolumeClaim(e.Client, runtimeInfo, e.Log)
}
