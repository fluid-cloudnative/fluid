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

package utils

import (
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// TrimVolumes trims the volumes
func TrimVolumes(inputs []corev1.Volume, excludeNames []string) (outputs []corev1.Volume) {
	outputs = []corev1.Volume{}
outer:
	for _, in := range inputs {
		for _, excludeName := range excludeNames {
			if strings.HasPrefix(in.Name, excludeName) {
				log.V(1).Info("Skip the volume", "volume", in, "exludeName", excludeName)
				continue outer
			}
		}

		outputs = append(outputs, in)
	}

	return
}

func TrimVolumeMounts(inputs []corev1.VolumeMount, excludeNames []string) (outputs []corev1.VolumeMount) {
	outputs = []corev1.VolumeMount{}
outer:
	for _, in := range inputs {
		for _, excludeName := range excludeNames {
			if strings.HasPrefix(in.Name, excludeName) {
				log.V(1).Info("Skip the volumeMount", "volumeMount", in, "exludeName", excludeName)
				continue outer
			}
		}

		outputs = append(outputs, in)

	}

	return
}

func FindVolumeByVolumeMount(volumeMount corev1.VolumeMount, volumes []corev1.Volume) *corev1.Volume {
	for _, vol := range volumes {
		if vol.Name == volumeMount.Name {
			return &vol
		}
	}

	return nil
}

func AppendOrOverrideVolume(volumes []corev1.Volume, vol corev1.Volume) []corev1.Volume {
	var existed bool
	for idx, v := range volumes {
		if v.Name == vol.Name {
			if !reflect.DeepEqual(v, vol) {
				// override existing volume
				volumes[idx] = vol
			}
			existed = true
			break
		}
	}

	if !existed {
		volumes = append(volumes, vol)
	}

	return volumes
}

func AppendOrOverrideVolumeMounts(volumeMounts []corev1.VolumeMount, vm corev1.VolumeMount) []corev1.VolumeMount {
	var existed bool
	for idx, m := range volumeMounts {
		if m.Name == vm.Name {
			if !reflect.DeepEqual(m, vm) {
				// override existing volume mount
				volumeMounts[idx] = vm
			}
			existed = true
			break
		}
	}

	if !existed {
		volumeMounts = append(volumeMounts, vm)
	}

	return volumeMounts
}

// FilterVolumesByVolumeMounts returns volumes that exists in the volumeMounts
func FilterVolumesByVolumeMounts(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount) []corev1.Volume {
	retVolumes := []corev1.Volume{}
	for _, vol := range volumes {
		var exists bool
		for _, volMount := range volumeMounts {
			if volMount.Name == vol.Name {
				exists = true
				break
			}
		}

		if exists {
			retVolumes = append(retVolumes, vol)
		}
	}

	return retVolumes
}
