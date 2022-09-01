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

package juicefs

import (
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// transform worker volumes
func (j *JuiceFSEngine) transformWorkerVolumes(runtime *datav1alpha1.JuiceFSRuntime, value *JuiceFS) (err error) {
	if len(runtime.Spec.Worker.VolumeMounts) > 0 {
		for _, volumeMount := range runtime.Spec.Worker.VolumeMounts {
			var volume *corev1.Volume

			for _, v := range runtime.Spec.Volumes {
				if v.Name == volumeMount.Name {
					volume = &v
					break
				}
			}

			if volume == nil {
				return fmt.Errorf("failed to find the volume for volumeMount %s", volumeMount.Name)
			}

			if len(value.Worker.VolumeMounts) == 0 {
				value.Worker.VolumeMounts = []corev1.VolumeMount{}
			}
			value.Worker.VolumeMounts = append(value.Worker.VolumeMounts, volumeMount)

			if len(value.Worker.Volumes) == 0 {
				value.Worker.Volumes = []corev1.Volume{}
			}
			value.Worker.Volumes = append(value.Worker.Volumes, *volume)
		}
	}

	return err
}

// transform fuse volumes
func (j *JuiceFSEngine) transformFuseVolumes(runtime *datav1alpha1.JuiceFSRuntime, value *JuiceFS) (err error) {
	if len(runtime.Spec.Fuse.VolumeMounts) > 0 {
		for _, volumeMount := range runtime.Spec.Fuse.VolumeMounts {
			var volume *corev1.Volume
			for _, v := range runtime.Spec.Volumes {
				if v.Name == volumeMount.Name {
					volume = &v
					break
				}
			}

			if volume == nil {
				return fmt.Errorf("failed to find the volume for volumeMount %s", volumeMount.Name)
			}

			if len(value.Fuse.VolumeMounts) == 0 {
				value.Fuse.VolumeMounts = []corev1.VolumeMount{}
			}
			value.Fuse.VolumeMounts = append(value.Fuse.VolumeMounts, volumeMount)

			if len(value.Fuse.Volumes) == 0 {
				value.Fuse.Volumes = []corev1.Volume{}
			}
			value.Fuse.Volumes = append(value.Fuse.Volumes, *volume)
		}
	}

	return err
}
