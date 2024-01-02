/*
Copyright 2024 The Fluid Authors.
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

package vineyard

import (
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// transform master volumes
func (e *VineyardEngine) transformMasterVolumes(runtime *datav1alpha1.VineyardRuntime, value *Vineyard) (err error) {
	if len(runtime.Spec.Master.VolumeMounts) > 0 {
		for _, volumeMount := range runtime.Spec.Master.VolumeMounts {
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

			if len(value.Master.VolumeMounts) == 0 {
				value.Master.VolumeMounts = []corev1.VolumeMount{}
			}
			value.Master.VolumeMounts = append(value.Master.VolumeMounts, volumeMount)

			if len(value.Volumes) == 0 {
				value.Volumes = []corev1.Volume{}
			}
			value.Volumes = append(value.Volumes, *volume)
		}
	}

	return err
}

// transform worker volumes
func (e *VineyardEngine) transformWorkerVolumes(runtime *datav1alpha1.VineyardRuntime, value *Vineyard) (err error) {
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

			if len(value.Volumes) == 0 {
				value.Volumes = []corev1.Volume{}
			}
			value.Volumes = append(value.Volumes, *volume)
		}
	}

	return err
}
